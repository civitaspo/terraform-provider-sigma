package sigma

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListAll(t *testing.T) {
	t.Parallel()

	type entry struct {
		ID string `json:"id"`
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth/token", validTokenHandler)
	mux.HandleFunc("/v2.1/items", func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Query().Get("page") {
		case "":
			next := "cursor-2"
			_ = json.NewEncoder(response).Encode(pageEnvelope[entry]{
				Entries:  []entry{{ID: "one"}},
				NextPage: &next,
			})
		case "cursor-2":
			_ = json.NewEncoder(response).Encode(pageEnvelope[entry]{
				Entries: []entry{{ID: "two"}},
			})
		default:
			http.Error(response, "unexpected page", http.StatusBadRequest)
		}
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	entries, err := ListAll[entry](context.Background(), client, "/v2.1/items?limit=1")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].ID != "one" || entries[1].ID != "two" {
		t.Errorf("entries = %#v", entries)
	}
}

func TestListAllAcceptsBareArray(t *testing.T) {
	t.Parallel()

	type entry struct {
		ID string `json:"id"`
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth/token", validTokenHandler)
	mux.HandleFunc("/v2.1/items", func(response http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(response).Encode([]entry{{ID: "legacy"}})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	entries, err := ListAll[entry](context.Background(), client, "/v2.1/items")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].ID != "legacy" {
		t.Errorf("entries = %#v", entries)
	}
}

func TestListAllDetectsCursorCycle(t *testing.T) {
	t.Parallel()

	type entry struct {
		ID string `json:"id"`
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth/token", validTokenHandler)
	mux.HandleFunc("/v2.1/items", func(response http.ResponseWriter, _ *http.Request) {
		next := "cursor-loop"
		_ = json.NewEncoder(response).Encode(pageEnvelope[entry]{
			Entries:  []entry{{ID: "one"}},
			NextPage: &next,
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	_, err = ListAll[entry](context.Background(), client, "/v2.1/items")
	if err == nil || !strings.Contains(err.Error(), "pagination cycle") {
		t.Fatalf("error = %v, want pagination cycle", err)
	}
}

func TestListAllRejectsUnexpectedEnvelope(t *testing.T) {
	t.Parallel()

	type entry struct {
		ID string `json:"id"`
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth/token", validTokenHandler)
	mux.HandleFunc("/v2.1/items", func(response http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(response).Encode(map[string]any{"data": []entry{{ID: "x"}}})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	_, err = ListAll[entry](context.Background(), client, "/v2.1/items")
	if err == nil || !strings.Contains(err.Error(), "unexpected Sigma list response envelope") {
		t.Fatalf("error = %v, want unexpected envelope", err)
	}
}
