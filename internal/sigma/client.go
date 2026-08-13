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

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma/openapi"
)

const defaultMaxRetries = 5
const defaultBackoff = 100 * time.Millisecond

// Client is a client for the Sigma REST API.
type Client struct {
	baseURL      *url.URL
	clientID     string
	clientSecret string
	httpClient   *http.Client
	userAgent    string
	now          func() time.Time
	sleep        func(context.Context, time.Duration) error
	maxRetries   int
	backoff      time.Duration

	auth      authState
	transport *apiTransport
	api       *openapi.Client
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

// WithUserAgent configures the User-Agent header sent with API requests.
func WithUserAgent(userAgent string) Option {
	return func(client *Client) {
		if userAgent != "" {
			client.userAgent = userAgent
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
		httpClient:   newHTTPClient(),
		userAgent:    defaultUserAgent,
		now:          time.Now,
		sleep:        sleepContext,
		maxRetries:   defaultMaxRetries,
		backoff:      defaultBackoff,
	}
	for _, opt := range opts {
		opt(client)
	}
	client.httpClient.CheckRedirect = client.checkRedirect

	transport := &apiTransport{client: client}
	client.transport = transport
	generated, err := openapi.NewClient(parsedBaseURL.String(), openapi.WithHTTPClient(transport))
	if err != nil {
		return nil, fmt.Errorf("create generated Sigma API client: %w", err)
	}
	client.api = generated

	return client, nil
}

func (client *Client) invalidateAccessToken() {
	client.auth.mu.Lock()
	defer client.auth.mu.Unlock()
	client.auth.accessToken = ""
	client.auth.expiresAt = time.Time{}
}

func (client *Client) doAPI(call func() (*http.Response, error), requireObject bool) ([]byte, error) {
	response, err := call()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read Sigma API response: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, apiErrorFrom(response, body)
	}
	if requireObject && len(bytes.TrimSpace(body)) == 0 {
		return nil, fmt.Errorf("sigma API returned an empty body when a response object is required")
	}
	return body, nil
}

func (client *Client) doDecode(call func() (*http.Response, error), target any) error {
	body, err := client.doAPI(call, true)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("decode Sigma API response: %w", err)
	}
	return nil
}

func (client *Client) doVoid(call func() (*http.Response, error)) error {
	_, err := client.doAPI(call, false)
	return err
}

func encodeBody(payload any) (io.Reader, error) {
	if payload == nil {
		return http.NoBody, nil
	}
	if raw, ok := payload.(json.RawMessage); ok {
		return bytes.NewReader(raw), nil
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode Sigma API request: %w", err)
	}
	return bytes.NewReader(body), nil
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
	if delay <= 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
