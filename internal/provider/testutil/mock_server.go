package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// MockSigma is an httptest-backed Sigma API server.
type MockSigma struct {
	Server       *httptest.Server
	Mux          *http.ServeMux
	ClientID     string
	ClientSecret string
	AccessToken  string
	TokenCalls   atomic.Int64
}

// NewMockSigma creates a mock with a built-in client-credentials token endpoint.
func NewMockSigma(t *testing.T) *MockSigma {
	t.Helper()
	mock := &MockSigma{
		Mux:          http.NewServeMux(),
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		AccessToken:  "test-access-token",
	}
	mock.Mux.HandleFunc("/v2/auth/token", mock.serveToken)
	mock.Server = httptest.NewServer(mock.Mux)
	t.Cleanup(mock.Server.Close)
	return mock
}

// URL returns the mock API base URL.
func (mock *MockSigma) URL() string {
	return mock.Server.URL
}

// AssertBearer verifies the Authorization header.
func (mock *MockSigma) AssertBearer(t *testing.T, request *http.Request) {
	t.Helper()
	if got, want := request.Header.Get("Authorization"), "Bearer "+mock.AccessToken; got != want {
		t.Errorf("Authorization header = %q, want %q", got, want)
	}
}

func (mock *MockSigma) serveToken(response http.ResponseWriter, request *http.Request) {
	mock.TokenCalls.Add(1)
	if request.Method != http.MethodPost {
		http.Error(response, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if got := request.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
		http.Error(response, "unexpected content type", http.StatusBadRequest)
		return
	}
	if err := request.ParseForm(); err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}
	if request.Form.Get("grant_type") != "client_credentials" ||
		request.Form.Get("client_id") != mock.ClientID ||
		request.Form.Get("client_secret") != mock.ClientSecret {
		http.Error(response, "invalid client credentials form", http.StatusBadRequest)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(response).Encode(map[string]any{
		"access_token": mock.AccessToken,
		"token_type":   "Bearer",
		"expires_in":   3600,
	})
}
