package sigma

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
