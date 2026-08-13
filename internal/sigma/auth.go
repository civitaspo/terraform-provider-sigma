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
	refresh     *tokenRefresh
}

type tokenRefresh struct {
	done  chan struct{}
	token string
	err   error
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

func (client *Client) accessToken(ctx context.Context) (string, error) {
	client.auth.mu.Lock()
	now := client.now()
	if client.auth.accessToken != "" && now.Before(client.auth.expiresAt.Add(-tokenRefreshWindow)) {
		token := client.auth.accessToken
		client.auth.mu.Unlock()
		return token, nil
	}
	if inFlight := client.auth.refresh; inFlight != nil {
		client.auth.mu.Unlock()
		return waitForRefresh(ctx, inFlight)
	}

	refresh := &tokenRefresh{done: make(chan struct{})}
	client.auth.refresh = refresh
	client.auth.mu.Unlock()

	token, expiresIn, err := client.fetchAccessToken(ctx)

	client.auth.mu.Lock()
	refresh.token = token
	refresh.err = err
	if err == nil {
		client.auth.accessToken = token
		client.auth.expiresAt = client.now().Add(time.Duration(expiresIn) * time.Second)
	}
	client.auth.refresh = nil
	close(refresh.done)
	client.auth.mu.Unlock()

	return token, err
}

func waitForRefresh(ctx context.Context, refresh *tokenRefresh) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-refresh.done:
		return refresh.token, refresh.err
	}
}

func (client *Client) fetchAccessToken(ctx context.Context) (string, int64, error) {
	client.auth.mu.Lock()
	lastRequest := client.auth.lastRequest
	client.auth.mu.Unlock()

	if wait := tokenRequestGap - client.now().Sub(lastRequest); !lastRequest.IsZero() && wait > 0 {
		if err := client.sleep(ctx, wait); err != nil {
			return "", 0, err
		}
	}

	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {client.clientID},
		"client_secret": {client.clientSecret},
	}
	endpoint, err := client.resolveEndpoint(tokenPath)
	if err != nil {
		return "", 0, err
	}

	var token tokenResponse
	for attempt := 0; attempt <= client.maxRetries; attempt++ {
		client.auth.mu.Lock()
		client.auth.lastRequest = client.now()
		client.auth.mu.Unlock()

		request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), strings.NewReader(form.Encode()))
		if err != nil {
			return "", 0, fmt.Errorf("create Sigma token request: %w", err)
		}
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("Accept", jsonContentType)
		request.Header.Set("User-Agent", client.userAgent)

		response, err := client.httpClient.Do(request)
		if err != nil {
			return "", 0, fmt.Errorf("send Sigma token request: %w", err)
		}
		if shouldRetry(response.StatusCode) && attempt < client.maxRetries {
			drainAndClose(response)
			if err := client.sleep(ctx, retryDelay(response, client.backoff, attempt, client.now())); err != nil {
				return "", 0, err
			}
			continue
		}
		body, readErr := io.ReadAll(response.Body)
		_ = response.Body.Close()
		if readErr != nil {
			return "", 0, fmt.Errorf("read Sigma token response: %w", readErr)
		}
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			return "", 0, apiErrorFrom(response, body)
		}
		if err := json.Unmarshal(body, &token); err != nil {
			return "", 0, fmt.Errorf("decode Sigma token response: %w", err)
		}
		break
	}
	if token.AccessToken == "" || token.ExpiresIn <= 0 {
		return "", 0, fmt.Errorf("sigma token response is missing access_token or expires_in")
	}
	return token.AccessToken, token.ExpiresIn, nil
}

// resolveEndpoint joins an API path with the configured base URL, preserving any
// path prefix on base_url (for reverse proxies such as https://proxy.example/sigma).
func (client *Client) resolveEndpoint(rawURL string) (*url.URL, error) {
	reference, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse Sigma API path: %w", err)
	}
	if reference.IsAbs() {
		if err := client.validateRequestURL(reference); err != nil {
			return nil, err
		}
		return reference, nil
	}

	base := *client.baseURL
	if base.Path == "" {
		base.Path = "/"
	} else if !strings.HasSuffix(base.Path, "/") {
		base.Path += "/"
	}
	reference.Path = strings.TrimPrefix(reference.Path, "/")
	return base.ResolveReference(reference), nil
}
