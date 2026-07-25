package sigma

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	tokenRefreshWindow = 5 * time.Minute
	tokenRequestGap    = time.Second
)

type authState struct {
	mu          sync.Mutex
	accessToken string
	expiresAt   time.Time
	lastRequest time.Time
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

func (client *Client) accessToken(ctx context.Context) (string, error) {
	client.auth.mu.Lock()
	defer client.auth.mu.Unlock()

	now := client.now()
	if client.auth.accessToken != "" && now.Before(client.auth.expiresAt.Add(-tokenRefreshWindow)) {
		return client.auth.accessToken, nil
	}

	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {client.clientID},
		"client_secret": {client.clientSecret},
	}
	endpoint, err := client.resolveEndpoint("/v2/auth/token")
	if err != nil {
		return "", err
	}
	var token tokenResponse
	for attempt := 0; attempt <= client.maxRetries; attempt++ {
		if wait := tokenRequestGap - client.now().Sub(client.auth.lastRequest); !client.auth.lastRequest.IsZero() && wait > 0 {
			if err := client.sleep(ctx, wait); err != nil {
				return "", err
			}
		}

		request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), strings.NewReader(form.Encode()))
		if err != nil {
			return "", fmt.Errorf("create Sigma token request: %w", err)
		}
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("Accept", "application/json")

		client.auth.lastRequest = client.now()
		response, err := client.httpClient.Do(request)
		if err != nil {
			return "", fmt.Errorf("send Sigma token request: %w", err)
		}
		if shouldRetry(response.StatusCode) && attempt < client.maxRetries {
			_, _ = io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			if err := client.sleep(ctx, retryDelay(response, client.backoff, attempt)); err != nil {
				return "", err
			}
			continue
		}
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			err := decodeAPIError(response)
			_ = response.Body.Close()
			return "", err
		}
		if err := json.NewDecoder(response.Body).Decode(&token); err != nil {
			_ = response.Body.Close()
			return "", fmt.Errorf("decode Sigma token response: %w", err)
		}
		_ = response.Body.Close()
		break
	}
	if token.AccessToken == "" || token.ExpiresIn <= 0 {
		return "", fmt.Errorf("sigma token response is missing access_token or expires_in")
	}

	issuedAt := client.now()
	client.auth.accessToken = token.AccessToken
	client.auth.expiresAt = issuedAt.Add(time.Duration(token.ExpiresIn) * time.Second)
	return token.AccessToken, nil
}
