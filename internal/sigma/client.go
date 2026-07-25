package sigma

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultMaxRetries = 5
	defaultBackoff    = 100 * time.Millisecond
)

// Client is a client for the Sigma REST API.
type Client struct {
	baseURL      *url.URL
	clientID     string
	clientSecret string
	httpClient   *http.Client
	now          func() time.Time
	sleep        func(context.Context, time.Duration) error
	maxRetries   int
	backoff      time.Duration

	auth authState
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient configures the HTTP client used for API requests.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *Client) {
		if httpClient != nil {
			client.httpClient = httpClient
		}
	}
}

// NewClient creates a Sigma REST API client.
func NewClient(baseURL, clientID, clientSecret string, opts ...Option) (*Client, error) {
	parsedBaseURL, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse Sigma base URL: %w", err)
	}
	if (parsedBaseURL.Scheme != "http" && parsedBaseURL.Scheme != "https") || parsedBaseURL.Host == "" {
		return nil, fmt.Errorf("sigma base URL must be an absolute HTTP or HTTPS URL")
	}

	client := &Client{
		baseURL:      parsedBaseURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   http.DefaultClient,
		now:          time.Now,
		sleep:        sleepContext,
		maxRetries:   defaultMaxRetries,
		backoff:      defaultBackoff,
	}
	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

func (client *Client) invalidateAccessToken() {
	client.auth.mu.Lock()
	defer client.auth.mu.Unlock()
	client.auth.accessToken = ""
	client.auth.expiresAt = time.Time{}
}

func (client *Client) do(ctx context.Context, method, path string, body []byte) (*http.Response, error) {
	retryable := isIdempotent(method)
	attempts := 1
	if retryable {
		attempts += client.maxRetries
	}

	unauthorizedRetried := false
	for attempt := 0; attempt < attempts; attempt++ {
		token, err := client.accessToken(ctx)
		if err != nil {
			return nil, err
		}

		endpoint, err := client.resolveEndpoint(path)
		if err != nil {
			return nil, err
		}
		request, err := http.NewRequestWithContext(ctx, method, endpoint.String(), bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("create Sigma API request: %w", err)
		}
		request.Header.Set("Authorization", "Bearer "+token)
		request.Header.Set("Accept", "application/json")
		if len(body) > 0 {
			request.Header.Set("Content-Type", "application/json")
		}

		response, err := client.httpClient.Do(request)
		if err != nil {
			return nil, fmt.Errorf("send Sigma API request: %w", err)
		}
		if response.StatusCode == http.StatusUnauthorized && !unauthorizedRetried {
			unauthorizedRetried = true
			_, _ = io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			client.invalidateAccessToken()
			// Retry once after invalidating the cached token. This does not consume
			// idempotent retry budget so POST/PATCH also recover from revoked tokens.
			attempt--
			continue
		}
		if !shouldRetry(response.StatusCode) || attempt == attempts-1 {
			return response, nil
		}

		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()
		delay := retryDelay(response, client.backoff, attempt)
		if err := client.sleep(ctx, delay); err != nil {
			return nil, err
		}
	}

	panic("unreachable")
}

// resolveEndpoint joins an API path with the configured base URL, preserving any
// path prefix on base_url (for reverse proxies such as https://proxy.example/sigma).
func (client *Client) resolveEndpoint(rawURL string) (*url.URL, error) {
	reference, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse Sigma API path: %w", err)
	}
	if reference.IsAbs() {
		return reference, nil
	}

	base := *client.baseURL
	if base.Path == "" {
		base.Path = "/"
	} else if !strings.HasSuffix(base.Path, "/") {
		base.Path += "/"
	}
	// Root-absolute paths would replace base.Path; treat them as relative to the prefix.
	reference.Path = strings.TrimPrefix(reference.Path, "/")
	return base.ResolveReference(reference), nil
}

func (client *Client) getJSON(ctx context.Context, path string, target any) error {
	body, err := client.getRaw(ctx, path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("decode Sigma API response: %w", err)
	}
	return nil
}

func (client *Client) getRaw(ctx context.Context, path string) ([]byte, error) {
	response, err := client.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, decodeAPIError(response)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read Sigma API response: %w", err)
	}
	return body, nil
}

func (client *Client) sendJSON(ctx context.Context, method, path string, payload, target any) error {
	var body []byte
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encode Sigma API request: %w", err)
		}
	}
	response, err := client.do(ctx, method, path, body)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return decodeAPIError(response)
	}
	if target == nil || response.StatusCode == http.StatusNoContent {
		_, _ = io.Copy(io.Discard, response.Body)
		return nil
	}
	if err := json.NewDecoder(response.Body).Decode(target); err != nil && err != io.EOF {
		return fmt.Errorf("decode Sigma API response: %w", err)
	}
	return nil
}

func isIdempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodOptions:
		return true
	default:
		return false
	}
}

func shouldRetry(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || statusCode >= http.StatusInternalServerError
}

func retryDelay(response *http.Response, base time.Duration, attempt int) time.Duration {
	if response.StatusCode == http.StatusTooManyRequests {
		if delay, ok := parseRetryAfter(response.Header.Get("Retry-After"), time.Now()); ok {
			return delay
		}
	}
	return base * time.Duration(1<<attempt)
}

func parseRetryAfter(value string, now time.Time) (time.Duration, bool) {
	if seconds, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second, true
	}
	if retryAt, err := http.ParseTime(value); err == nil {
		delay := retryAt.Sub(now)
		if delay < 0 {
			delay = 0
		}
		return delay, true
	}
	return 0, false
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
