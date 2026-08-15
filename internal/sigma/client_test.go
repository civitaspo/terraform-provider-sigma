package sigma

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientRefreshesTokenFiveMinutesBeforeExpiry(t *testing.T) {
	t.Parallel()

	var tokenCalls atomic.Int64
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth/token", func(response http.ResponseWriter, request *http.Request) {
		call := tokenCalls.Add(1)
		if err := request.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if got := request.Form.Get("grant_type"); got != "client_credentials" {
			t.Errorf("grant_type = %q", got)
		}
		if got := request.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type = %q", got)
		}
		_ = json.NewEncoder(response).Encode(map[string]any{
			"access_token": "token-" + string(rune('0'+call)),
			"token_type":   "Bearer",
			"expires_in":   601,
		})
	})
	mux.HandleFunc("/v2/whoami", func(response http.ResponseWriter, request *http.Request) {
		_ = json.NewEncoder(response).Encode(map[string]string{
			"userId":         "user-1",
			"organizationId": "org-1",
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	now := time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC)
	client, err := NewClient(server.URL, "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	client.now = func() time.Time { return now }
	client.sleep = func(context.Context, time.Duration) error { return nil }

	if _, err := client.Whoami(context.Background()); err != nil {
		t.Fatal(err)
	}
	now = now.Add(300 * time.Second)
	if _, err := client.Whoami(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := tokenCalls.Load(); got != 1 {
		t.Fatalf("token calls before refresh window = %d, want 1", got)
	}
	now = now.Add(2 * time.Second)
	if _, err := client.Whoami(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := tokenCalls.Load(); got != 2 {
		t.Fatalf("token calls after refresh window = %d, want 2", got)
	}
}

func TestClientRetries429UsingRetryAfter(t *testing.T) {
	t.Parallel()

	var whoamiCalls atomic.Int64
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth/token", validTokenHandler)
	mux.HandleFunc("/v2/whoami", func(response http.ResponseWriter, request *http.Request) {
		if whoamiCalls.Add(1) == 1 {
			response.Header().Set("Retry-After", "2")
			http.Error(response, "rate limited", http.StatusTooManyRequests)
			return
		}
		if got := request.Header.Get("Authorization"); got != "Bearer token" {
			t.Errorf("Authorization = %q", got)
		}
		if got := request.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q", got)
		}
		_ = json.NewEncoder(response).Encode(map[string]string{
			"userId":         "user-1",
			"organizationId": "org-1",
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	var slept time.Duration
	client.sleep = func(_ context.Context, delay time.Duration) error {
		slept = delay
		return nil
	}
	if _, err := client.Whoami(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := whoamiCalls.Load(); got != 2 {
		t.Errorf("whoami calls = %d, want 2", got)
	}
	if slept != 2*time.Second {
		t.Errorf("retry delay = %s, want 2s", slept)
	}
}

func TestClientSerializesConcurrentTokenRequests(t *testing.T) {
	t.Parallel()

	var tokenCalls atomic.Int64
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth/token", func(response http.ResponseWriter, request *http.Request) {
		tokenCalls.Add(1)
		time.Sleep(50 * time.Millisecond)
		validTokenHandler(response, request)
	})
	mux.HandleFunc("/v2/whoami", func(response http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(response).Encode(map[string]string{
			"userId":         "user-1",
			"organizationId": "org-1",
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	var wait sync.WaitGroup
	errs := make(chan error, 10)
	for range 10 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, callErr := client.Whoami(context.Background())
			errs <- callErr
		}()
	}
	wait.Wait()
	close(errs)
	for callErr := range errs {
		if callErr != nil {
			t.Error(callErr)
		}
	}
	if got := tokenCalls.Load(); got != 1 {
		t.Errorf("token calls = %d, want 1", got)
	}
}

func TestWhoami(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth/token", validTokenHandler)
	mux.HandleFunc("/v2/whoami", func(response http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(response).Encode(map[string]string{
			"userId":         "user-123",
			"organizationId": "org-456",
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	identity, err := client.Whoami(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if identity.UserID != "user-123" || identity.OrganizationID != "org-456" {
		t.Errorf("identity = %#v", identity)
	}
}

func TestClientRetriesOnceAfterUnauthorized(t *testing.T) {
	t.Parallel()

	var tokenCalls atomic.Int64
	var whoamiCalls atomic.Int64
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth/token", func(response http.ResponseWriter, request *http.Request) {
		call := tokenCalls.Add(1)
		_ = json.NewEncoder(response).Encode(map[string]any{
			"access_token": "token-" + string(rune('0'+call)),
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	})
	mux.HandleFunc("/v2/whoami", func(response http.ResponseWriter, request *http.Request) {
		call := whoamiCalls.Add(1)
		if call == 1 {
			if got := request.Header.Get("Authorization"); got != "Bearer token-1" {
				t.Errorf("first Authorization = %q", got)
			}
			http.Error(response, `{"message":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		if got := request.Header.Get("Authorization"); got != "Bearer token-2" {
			t.Errorf("retry Authorization = %q", got)
		}
		_ = json.NewEncoder(response).Encode(map[string]string{
			"userId":         "user-1",
			"organizationId": "org-1",
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	client.sleep = func(context.Context, time.Duration) error { return nil }

	if _, err := client.Whoami(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := tokenCalls.Load(); got != 2 {
		t.Fatalf("token calls = %d, want 2", got)
	}
	if got := whoamiCalls.Load(); got != 2 {
		t.Fatalf("whoami calls = %d, want 2", got)
	}
}

func TestClientPreservesBaseURLPathPrefix(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/sigma/v2/auth/token", validTokenHandler)
	mux.HandleFunc("/sigma/v2/whoami", func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/sigma/v2/whoami" {
			t.Errorf("path = %q, want /sigma/v2/whoami", request.URL.Path)
		}
		_ = json.NewEncoder(response).Encode(map[string]string{
			"userId":         "user-1",
			"organizationId": "org-1",
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL+"/sigma", "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	identity, err := client.Whoami(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if identity.UserID != "user-1" {
		t.Errorf("identity = %#v", identity)
	}
}

func TestResolveEndpointJoinsBasePath(t *testing.T) {
	t.Parallel()

	client, err := NewClient("https://proxy.example.com/sigma", "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	endpoint, err := client.resolveEndpoint("/v2/members?limit=1")
	if err != nil {
		t.Fatal(err)
	}
	if got := endpoint.String(); got != "https://proxy.example.com/sigma/v2/members?limit=1" {
		t.Errorf("endpoint = %q", got)
	}
}

func TestClientRejectsCrossHostRequest(t *testing.T) {
	t.Parallel()

	client, err := NewClient("https://api.sigmacomputing.com", "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequest(http.MethodGet, "https://evil.example/v2/whoami", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.transport.Do(request)
	if err == nil || !strings.Contains(err.Error(), "not the configured host") {
		t.Fatalf("error = %v, want cross-host rejection", err)
	}
}

func TestClientDoesNotRetryPOST(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth/token", validTokenHandler)
	mux.HandleFunc("/v2/members", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Errorf("method = %s", request.Method)
		}
		calls.Add(1)
		http.Error(response, `{"message":"rate limited"}`, http.StatusTooManyRequests)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	client.sleep = func(context.Context, time.Duration) error {
		t.Fatal("POST should not sleep for retry")
		return nil
	}
	_, err = client.CreateMember(context.Background(), CreateMemberInput{Email: "ada@example.com", FirstName: "Ada", LastName: "Lovelace"})
	if err == nil {
		t.Fatal("expected error")
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("POST calls = %d, want 1", got)
	}
}

func TestClientDoesNotRetryPATCH(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth/token", validTokenHandler)
	mux.HandleFunc("/v2/members/member-1", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPatch {
			t.Errorf("method = %s", request.Method)
		}
		calls.Add(1)
		http.Error(response, `{"message":"unavailable"}`, http.StatusBadGateway)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	client.sleep = func(context.Context, time.Duration) error {
		t.Fatal("PATCH should not sleep for retry")
		return nil
	}
	first := "Ada"
	_, err = client.UpdateMember(context.Background(), "member-1", UpdateMemberInput{FirstName: &first})
	if err == nil {
		t.Fatal("expected error")
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("PATCH calls = %d, want 1", got)
	}
}

func TestClientRetriesIdempotent5xxAndHonorsCancellation(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth/token", validTokenHandler)
	mux.HandleFunc("/v2/whoami", func(response http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		http.Error(response, `{"message":"unavailable"}`, http.StatusBadGateway)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	client.sleep = func(ctx context.Context, _ time.Duration) error {
		cancel()
		return ctx.Err()
	}
	_, err = client.Whoami(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context cancellation", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("whoami calls = %d, want 1 before cancelled backoff", got)
	}
}

func TestClientCapsRetryAfter(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth/token", validTokenHandler)
	mux.HandleFunc("/v2/whoami", func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Retry-After", "120")
		http.Error(response, "rate limited", http.StatusTooManyRequests)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	var slept time.Duration
	client.sleep = func(_ context.Context, delay time.Duration) error {
		slept = delay
		return fmt.Errorf("stop after first backoff")
	}
	_, err = client.Whoami(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if slept != 60*time.Second {
		t.Errorf("retry delay = %s, want 60s cap", slept)
	}
}

func TestWhoamiEmptyBodyIsError(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth/token", validTokenHandler)
	mux.HandleFunc("/v2/whoami", func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Whoami(context.Background())
	if err == nil || !strings.Contains(err.Error(), "empty body") {
		t.Fatalf("error = %v, want empty body", err)
	}
}

func TestRetryDelay429WithoutRetryAfterUsesMinimum(t *testing.T) {
	t.Parallel()
	response := &http.Response{StatusCode: http.StatusTooManyRequests, Header: http.Header{}}
	delay := retryDelay(response, 100*time.Millisecond, 0, time.Now())
	if delay < time.Second {
		t.Fatalf("delay = %s, want at least 1s after Cloudflare 429 floor", delay)
	}
}

func validTokenHandler(response http.ResponseWriter, _ *http.Request) {
	_ = json.NewEncoder(response).Encode(map[string]any{
		"access_token": "token",
		"token_type":   "Bearer",
		"expires_in":   3600,
	})
}
