package sigma

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListAllByPageSecondPage(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth/token", validTokenHandler)
	mux.HandleFunc("/v2/members", func(response http.ResponseWriter, request *http.Request) {
		if got := request.Header.Get("User-Agent"); !strings.Contains(got, "terraform-provider-sigma") {
			t.Errorf("User-Agent = %q", got)
		}
		switch request.URL.Query().Get("page") {
		case "":
			next := "cursor-2"
			_ = json.NewEncoder(response).Encode(pageEnvelope[Member]{
				Entries:  []Member{{MemberID: "one"}},
				NextPage: &next,
			})
		case "cursor-2":
			_ = json.NewEncoder(response).Encode(pageEnvelope[Member]{
				Entries: []Member{{MemberID: "two"}},
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
	entries, err := client.ListMembers(context.Background(), ListMembersOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].MemberID != "one" || entries[1].MemberID != "two" {
		t.Errorf("entries = %#v", entries)
	}
}

func TestListAllByPageTokenSecondPage(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth/token", validTokenHandler)
	mux.HandleFunc("/v2/accountTypes", func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Query().Get("pageToken") {
		case "":
			next := "token-2"
			_ = json.NewEncoder(response).Encode(pageTokenEnvelope[AccountType]{
				Entries:       []AccountType{{AccountTypeID: "one"}},
				NextPageToken: &next,
			})
		case "token-2":
			_ = json.NewEncoder(response).Encode(pageTokenEnvelope[AccountType]{
				Entries: []AccountType{{AccountTypeID: "two"}},
			})
		default:
			http.Error(response, "unexpected pageToken", http.StatusBadRequest)
		}
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	entries, err := client.ListAccountTypes(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].AccountTypeID != "one" || entries[1].AccountTypeID != "two" {
		t.Errorf("entries = %#v", entries)
	}
}

func TestListAllByPageDetectsCursorCycle(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth/token", validTokenHandler)
	mux.HandleFunc("/v2/members", func(response http.ResponseWriter, _ *http.Request) {
		next := "cursor-loop"
		_ = json.NewEncoder(response).Encode(pageEnvelope[Member]{
			Entries:  []Member{{MemberID: "one"}},
			NextPage: &next,
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.ListMembers(context.Background(), ListMembersOptions{})
	if err == nil || !strings.Contains(err.Error(), "pagination cycle") {
		t.Fatalf("error = %v, want pagination cycle", err)
	}
}

func TestListAllByPageTokenDetectsCursorCycle(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth/token", validTokenHandler)
	mux.HandleFunc("/v2/accountTypes", func(response http.ResponseWriter, _ *http.Request) {
		next := "token-loop"
		_ = json.NewEncoder(response).Encode(pageTokenEnvelope[AccountType]{
			Entries:       []AccountType{{AccountTypeID: "one"}},
			NextPageToken: &next,
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.ListAccountTypes(context.Background())
	if err == nil || !strings.Contains(err.Error(), "nextPageToken") {
		t.Fatalf("error = %v, want nextPageToken cycle", err)
	}
}

func TestListAllByPageCap(t *testing.T) {
	t.Parallel()

	calls := 0
	_, err := listAllCursors(context.Background(), func(context.Context, *string) ([]int, *string, error) {
		calls++
		next := fmt.Sprintf("cursor-%d", calls)
		return []int{calls}, &next, nil
	}, "nextPage", 2)
	if err == nil || !strings.Contains(err.Error(), "exceeded 2 pages") {
		t.Fatalf("error = %v, want page cap", err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestListAllByPageEmptySlice(t *testing.T) {
	t.Parallel()

	entries, err := listAllByPage(context.Background(), func(context.Context, *string) ([]string, *string, error) {
		return nil, nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if entries == nil || len(entries) != 0 {
		t.Fatalf("entries = %#v, want initialized empty slice", entries)
	}
}

func TestListAllAcceptsBareArray(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth/token", validTokenHandler)
	mux.HandleFunc("/v2/members", func(response http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(response).Encode([]Member{{MemberID: "legacy"}})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	entries, err := client.ListMembers(context.Background(), ListMembersOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].MemberID != "legacy" {
		t.Errorf("entries = %#v", entries)
	}
}

func TestListAllRejectsUnexpectedEnvelope(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth/token", validTokenHandler)
	mux.HandleFunc("/v2/members", func(response http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(response).Encode(map[string]any{"data": []Member{{MemberID: "x"}}})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, "client-id", "client-secret")
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.ListMembers(context.Background(), ListMembersOptions{})
	if err == nil || !strings.Contains(err.Error(), "unexpected Sigma list response envelope") {
		t.Fatalf("error = %v, want unexpected envelope", err)
	}
}
