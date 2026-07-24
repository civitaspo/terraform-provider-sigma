package sigma

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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
