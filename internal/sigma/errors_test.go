package sigma

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPIErrorAndIsNotFound(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth/token", validTokenHandler)
	mux.HandleFunc("/v2/whoami", func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("X-Request-Id", "request-123")
		response.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(response).Encode(map[string]string{
			"code":    "not_found",
			"message": "identity not found",
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Whoami(context.Background())
	if !IsNotFound(err) {
		t.Fatalf("IsNotFound(%v) = false", err)
	}
	var apiError *APIError
	if !errors.As(err, &apiError) {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiError.StatusCode != http.StatusNotFound ||
		apiError.Code != "not_found" ||
		apiError.Message != "identity not found" ||
		apiError.RequestID != "request-123" {
		t.Errorf("API error = %#v", apiError)
	}
}

func TestIsNotFoundRejectsHTMLProxy404(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth/token", validTokenHandler)
	mux.HandleFunc("/v2/whoami", func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusNotFound)
		_, _ = response.Write([]byte("<!DOCTYPE html><html><body>Not Found</body></html>"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Whoami(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if IsNotFound(err) {
		t.Fatalf("IsNotFound(%v) = true for HTML proxy body", err)
	}
}

func TestIsNotFoundRejectsUnknownErrorCode(t *testing.T) {
	t.Parallel()

	err := &APIError{StatusCode: http.StatusNotFound, Code: "route_not_configured", Message: "no upstream"}
	if IsNotFound(err) {
		t.Fatalf("IsNotFound(%v) = true for non-resource code", err)
	}
}

func TestIsNotFoundAcceptsBare404Message(t *testing.T) {
	t.Parallel()

	// Sigma often returns 404 without a code; provider-synthesized lookups do the same.
	err := &APIError{StatusCode: http.StatusNotFound, Message: "member not found"}
	if !IsNotFound(err) {
		t.Fatalf("IsNotFound(%v) = false", err)
	}
}

func TestAPIErrorIncludesRequestID(t *testing.T) {
	t.Parallel()

	err := &APIError{StatusCode: http.StatusNotFound, Code: "not_found", Message: "identity not found", RequestID: "request-123"}
	got := err.Error()
	if !strings.Contains(got, "404") || !strings.Contains(got, "not_found") || !strings.Contains(got, "request_id=request-123") || !strings.Contains(got, "identity not found") {
		t.Fatalf("Error() = %q", got)
	}
}
