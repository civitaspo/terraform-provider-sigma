package provider_test

import (
	"encoding/json"
	"net/http"
	"regexp"
	"sync"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestWorkbookGrantResourceUntaggedUsesDedicatedEndpoints(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	var mu sync.Mutex
	methods := map[string][]string{}
	record := func(request *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		methods[request.URL.Path] = append(methods[request.URL.Path], request.Method)
	}
	grant := map[string]any{
		"grantId": "workbook-grant-1", "inodeId": "workbook-1", "organizationId": "org-1",
		"memberId": "member-2", "teamId": nil, "permission": "explore", "inodeType": "workbook",
		"createdBy": "member-admin", "updatedBy": "member-admin",
		"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
	}
	mock.Mux.HandleFunc("/v2/workbooks/workbook-1/grants", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		record(request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(response, map[string]any{})
	})
	mock.Mux.HandleFunc("/v2/grants", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		record(request)
		if request.Method == http.MethodPost {
			t.Errorf("untagged grant must not POST /v2/grants")
			http.Error(response, "unexpected generic create", http.StatusBadRequest)
			return
		}
		writeJSON(response, map[string]any{"entries": []any{grant}, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/workbooks/workbook-1/grants/workbook-grant-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		record(request)
		writeJSON(response, map[string]any{})
	})
	config := providerConfig(mock) + `
resource "sigma_workbook_grant" "test" {
  inode_id   = "workbook-1"
  member_id  = "member-2"
  permission = "explore"
}
`
	resource.UnitTest(t, providerTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("sigma_workbook_grant.test", "id", "workbook-grant-1"),
			resource.TestCheckNoResourceAttr("sigma_workbook_grant.test", "tag_id"),
		),
	}}))
	mu.Lock()
	defer mu.Unlock()
	if got := methods["/v2/workbooks/workbook-1/grants"]; len(got) == 0 || got[0] != http.MethodPost {
		t.Fatalf("dedicated create = %v", methods)
	}
	if got := methods["/v2/workbooks/workbook-1/grants/workbook-grant-1"]; len(got) == 0 || got[len(got)-1] != http.MethodDelete {
		t.Fatalf("dedicated delete = %v", methods)
	}
}

func TestWorkbookGrantResourceTaggedUsesGenericEndpoints(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	var mu sync.Mutex
	methods := map[string][]string{}
	record := func(request *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		methods[request.URL.Path] = append(methods[request.URL.Path], request.Method)
	}
	grant := map[string]any{
		"grantId": "tagged-grant-1", "inodeId": "workbook-1", "organizationId": "org-1",
		"memberId": "member-2", "teamId": nil, "permission": "explore", "inodeType": "workbook",
		"createdBy": "member-admin", "updatedBy": "member-admin",
		"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
	}
	mock.Mux.HandleFunc("/v2/grants", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		record(request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		var payload map[string]any
		_ = json.NewDecoder(request.Body).Decode(&payload)
		if payload["tagId"] != "tag-1" {
			t.Errorf("tagId = %#v, want tag-1", payload["tagId"])
		}
		writeJSON(response, grant)
	})
	mock.Mux.HandleFunc("/v2/grants/tagged-grant-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		record(request)
		switch request.Method {
		case http.MethodGet:
			writeJSON(response, grant)
		case http.MethodDelete:
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/workbooks/workbook-1/grants", func(response http.ResponseWriter, request *http.Request) {
		t.Errorf("tagged grant must not use dedicated workbook grant endpoints: %s %s", request.Method, request.URL.Path)
		http.Error(response, "unexpected dedicated grant call", http.StatusBadRequest)
	})
	config := providerConfig(mock) + `
resource "sigma_workbook_grant" "test" {
  inode_id   = "workbook-1"
  member_id  = "member-2"
  permission = "explore"
  tag_id     = "tag-1"
}
`
	resource.UnitTest(t, providerTestCase([]resource.TestStep{
		{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_workbook_grant.test", "id", "tagged-grant-1"),
				resource.TestCheckResourceAttr("sigma_workbook_grant.test", "tag_id", "tag-1"),
			),
		},
		{Config: config},
	}))
	mu.Lock()
	defer mu.Unlock()
	if got := methods["/v2/grants"]; len(got) == 0 || got[0] != http.MethodPost {
		t.Fatalf("generic create = %v", methods)
	}
	if got := methods["/v2/grants/tagged-grant-1"]; len(got) < 2 {
		t.Fatalf("generic get/delete = %v", methods)
	}
}

func TestWorkbookGrantResourceAmbiguity(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	member := "member-2"
	grant := func(id string) map[string]any {
		return map[string]any{
			"grantId": id, "inodeId": "workbook-1", "organizationId": "org-1",
			"memberId": member, "teamId": nil, "permission": "explore", "inodeType": "workbook",
			"createdBy": "member-admin", "updatedBy": "member-admin",
			"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
		}
	}
	mock.Mux.HandleFunc("/v2/workbooks/workbook-1/grants", func(response http.ResponseWriter, request *http.Request) {
		writeJSON(response, map[string]any{})
	})
	mock.Mux.HandleFunc("/v2/grants", func(response http.ResponseWriter, request *http.Request) {
		writeJSON(response, map[string]any{"entries": []any{grant("grant-1"), grant("grant-2")}, "nextPage": nil})
	})
	config := providerConfig(mock) + `
resource "sigma_workbook_grant" "test" {
  inode_id   = "workbook-1"
  member_id  = "member-2"
  permission = "explore"
}
`
	resource.UnitTest(t, providerTestCase([]resource.TestStep{{
		Config:      config,
		ExpectError: regexp.MustCompile(`refusing to select`),
	}}))
}

func TestAccWorkbookGrantResource(t *testing.T) { requireAcceptance(t) }
