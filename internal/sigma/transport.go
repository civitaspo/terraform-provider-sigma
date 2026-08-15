package sigma

import (
	"bytes"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma/openapi"
)

const (
	dialTimeout           = 10 * time.Second
	tlsHandshakeTimeout   = 10 * time.Second
	responseHeaderTimeout = 30 * time.Second
	wholeRequestTimeout   = 60 * time.Second
	maxRetrySleep         = 60 * time.Second
	defaultUserAgent      = "terraform-provider-sigma"
	jsonContentType       = "application/json"
	tokenPath             = "/v2/auth/token"
)

var _ openapi.HttpRequestDoer = (*apiTransport)(nil)

type apiTransport struct {
	client *Client
}

func newHTTPClient() *http.Client {
	dialer := &net.Dialer{Timeout: dialTimeout}
	return &http.Client{
		Timeout: wholeRequestTimeout,
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dialer.DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   tlsHandshakeTimeout,
			ResponseHeaderTimeout: responseHeaderTimeout,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

func (transport *apiTransport) Do(request *http.Request) (*http.Response, error) {
	if err := transport.client.validateRequestURL(request.URL); err != nil {
		return nil, err
	}

	body, err := snapshotBody(request)
	if err != nil {
		return nil, err
	}

	retryable := isIdempotent(request.Method)
	attempts := 1
	if retryable {
		attempts += transport.client.maxRetries
	}

	unauthorizedRetried := false
	for attempt := 0; attempt < attempts; attempt++ {
		token, err := transport.client.accessToken(request.Context())
		if err != nil {
			return nil, err
		}

		outbound := cloneRequest(request, body)
		outbound.Header.Set("Authorization", "Bearer "+token)
		outbound.Header.Set("Accept", jsonContentType)
		outbound.Header.Set("User-Agent", transport.client.userAgent)
		if len(body) > 0 && outbound.Header.Get("Content-Type") == "" {
			outbound.Header.Set("Content-Type", jsonContentType)
		}

		response, err := transport.client.httpClient.Do(outbound)
		if err != nil {
			return nil, fmt.Errorf("send Sigma API request: %w", err)
		}
		if response.StatusCode == http.StatusUnauthorized && !unauthorizedRetried {
			unauthorizedRetried = true
			drainAndClose(response)
			transport.client.invalidateAccessToken()
			attempt--
			continue
		}
		if !shouldRetry(response.StatusCode) || attempt == attempts-1 {
			return response, nil
		}

		drainAndClose(response)
		delay := retryDelay(response, transport.client.backoff, attempt, transport.client.now())
		if err := transport.client.sleep(request.Context(), delay); err != nil {
			return nil, err
		}
	}

	panic("unreachable")
}

func (client *Client) validateRequestURL(requestURL *url.URL) error {
	if requestURL == nil {
		return fmt.Errorf("sigma API request is missing a URL")
	}
	if !requestURL.IsAbs() {
		return fmt.Errorf("sigma API request URL %q is not absolute", requestURL)
	}
	if requestURL.Scheme != client.baseURL.Scheme || !strings.EqualFold(requestURL.Host, client.baseURL.Host) {
		return fmt.Errorf("sigma API request URL %q is not the configured host %s://%s", requestURL, client.baseURL.Scheme, client.baseURL.Host)
	}
	basePath := strings.TrimSuffix(client.baseURL.Path, "/")
	if basePath == "" {
		return nil
	}
	path := requestURL.EscapedPath()
	if path != basePath && !strings.HasPrefix(path, basePath+"/") {
		return fmt.Errorf("sigma API request path %q is outside the configured base path %q", path, basePath)
	}
	return nil
}

func (client *Client) checkRedirect(request *http.Request, via []*http.Request) error {
	if err := client.validateRequestURL(request.URL); err != nil {
		return err
	}
	if len(via) >= 10 {
		return fmt.Errorf("stopped after 10 redirects")
	}
	return nil
}

func snapshotBody(request *http.Request) ([]byte, error) {
	if request.Body == nil || request.Body == http.NoBody {
		return nil, nil
	}
	defer func() {
		_ = request.Body.Close()
	}()
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, fmt.Errorf("read Sigma API request body: %w", err)
	}
	return body, nil
}

func cloneRequest(request *http.Request, body []byte) *http.Request {
	cloned := request.Clone(request.Context())
	if len(body) == 0 {
		cloned.Body = http.NoBody
		cloned.ContentLength = 0
		cloned.GetBody = func() (io.ReadCloser, error) {
			return http.NoBody, nil
		}
		return cloned
	}
	cloned.Body = io.NopCloser(bytes.NewReader(body))
	cloned.ContentLength = int64(len(body))
	cloned.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	return cloned
}

func drainAndClose(response *http.Response) {
	if response == nil || response.Body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, response.Body)
	_ = response.Body.Close()
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

func retryDelay(response *http.Response, base time.Duration, attempt int, now time.Time) time.Duration {
	var delay time.Duration
	if response != nil && response.StatusCode == http.StatusTooManyRequests {
		if parsed, ok := parseRetryAfter(response.Header.Get("Retry-After"), now); ok {
			delay = parsed
		}
	}
	if delay == 0 {
		delay = base * time.Duration(1<<attempt)
		if response != nil && response.StatusCode == http.StatusTooManyRequests && delay < 2*time.Second {
			// Cloudflare 1015 often omits Retry-After. A 100ms exponential
			// backoff burns the retry budget before the edge unblocks.
			delay = 2 * time.Second
		}
		delay = time.Duration(float64(delay) * (0.5 + rand.Float64()))
	}
	if delay > maxRetrySleep {
		return maxRetrySleep
	}
	if delay < 0 {
		return 0
	}
	return delay
}
