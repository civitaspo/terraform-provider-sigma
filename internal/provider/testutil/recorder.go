package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
)

// ExpectedRequest is one inbound Sigma API call the test still expects.
type ExpectedRequest struct {
	Method   string
	Path     string
	Query    map[string]string
	JSONBody any
	Status   int
	Response any
}

type recorder struct {
	t        *testing.T
	mu       sync.Mutex
	expected []ExpectedRequest
}

// NewRecordingSigma returns a mock Sigma server that fails on unexpected
// method/path, extra or missing query parameters, semantically incorrect JSON
// bodies, and unconsumed expected requests.
func NewRecordingSigma(t *testing.T, expected ...ExpectedRequest) *MockSigma {
	t.Helper()
	rec := &recorder{t: t, expected: append([]ExpectedRequest(nil), expected...)}
	mock := &MockSigma{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		AccessToken:  "test-access-token",
	}
	mock.Server = httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/v2/auth/token" {
			mock.serveToken(response, request)
			return
		}
		if got, want := request.Header.Get("Authorization"), "Bearer "+mock.AccessToken; got != want {
			t.Errorf("Authorization header = %q, want %q", got, want)
		}
		rec.handle(response, request)
	}))
	t.Cleanup(func() {
		rec.assertConsumed()
		mock.Server.Close()
	})
	return mock
}

func (rec *recorder) handle(response http.ResponseWriter, request *http.Request) {
	rec.t.Helper()
	rec.mu.Lock()
	if len(rec.expected) == 0 {
		rec.mu.Unlock()
		rec.t.Errorf("unexpected request %s %s", request.Method, request.URL.RequestURI())
		http.Error(response, "unexpected request", http.StatusBadRequest)
		return
	}
	next := rec.expected[0]
	rec.expected = rec.expected[1:]
	rec.mu.Unlock()

	if request.Method != next.Method || request.URL.Path != next.Path {
		rec.t.Errorf("got %s %s, want %s %s", request.Method, request.URL.Path, next.Method, next.Path)
		http.Error(response, "unexpected method/path", http.StatusBadRequest)
		return
	}
	gotQuery := map[string]string{}
	for key, values := range request.URL.Query() {
		if len(values) == 1 {
			gotQuery[key] = values[0]
			continue
		}
		rec.t.Errorf("query %s = %v", key, values)
	}
	wantQuery := next.Query
	if wantQuery == nil {
		wantQuery = map[string]string{}
	}
	if !reflect.DeepEqual(gotQuery, wantQuery) {
		rec.t.Errorf("query = %#v, want %#v", gotQuery, wantQuery)
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		rec.t.Errorf("read body: %v", err)
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}
	if next.JSONBody == nil {
		if len(bytes.TrimSpace(body)) != 0 {
			rec.t.Errorf("unexpected JSON body for %s %s: %s", request.Method, request.URL.Path, body)
		}
	} else {
		var got any
		if err := json.Unmarshal(body, &got); err != nil {
			rec.t.Errorf("decode request JSON: %v body=%s", err, body)
			http.Error(response, "invalid json", http.StatusBadRequest)
			return
		}
		wantRaw, err := json.Marshal(next.JSONBody)
		if err != nil {
			rec.t.Fatalf("encode expected JSON: %v", err)
		}
		var want any
		if err := json.Unmarshal(wantRaw, &want); err != nil {
			rec.t.Fatalf("decode expected JSON: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			rec.t.Errorf("JSON body = %s, want %s", body, wantRaw)
		}
	}

	status := next.Status
	if status == 0 {
		status = http.StatusOK
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	if next.Response != nil {
		_ = json.NewEncoder(response).Encode(next.Response)
	}
}

func (rec *recorder) assertConsumed() {
	rec.t.Helper()
	rec.mu.Lock()
	defer rec.mu.Unlock()
	if len(rec.expected) != 0 {
		rec.t.Errorf("unconsumed expected requests: %#v", rec.expected)
	}
}
