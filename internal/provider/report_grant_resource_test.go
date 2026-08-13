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

func TestReportGrantResourceUntaggedUsesDedicatedEndpoints(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	var mu sync.Mutex
	methods := map[string][]string{}
	record := func(request *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		methods[request.URL.Path] = append(methods[request.URL.Path], request.Method)
	}
	grant := map[string]any{
		"grantId": "report-grant-1", "inodeId": "report-1", "organizationId": "org-1",
		"memberId": nil, "teamId": "team-2", "permission": "edit", "inodeType": "report",
		"createdBy": "member-admin", "updatedBy": "member-admin",
		"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
	}
	mock.Mux.HandleFunc("/v2/reports/report-1/grants", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		record(request)
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
	mock.Mux.HandleFunc("/v2/reports/report-1/grants/report-grant-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		record(request)
		writeJSON(response, map[string]any{})
	})
	config := providerConfig(mock) + `
resource "sigma_report_grant" "test" {
  inode_id   = "report-1"
  team_id    = "team-2"
  permission = "edit"
}
`
	resource.UnitTest(t, providerTestCase([]resource.TestStep{{
		Config: config,
		Check:  resource.TestCheckResourceAttr("sigma_report_grant.test", "id", "report-grant-1"),
	}}))
	mu.Lock()
	defer mu.Unlock()
	if got := methods["/v2/reports/report-1/grants"]; len(got) == 0 || got[0] != http.MethodPost {
		t.Fatalf("dedicated create = %v", methods)
	}
	if got := methods["/v2/reports/report-1/grants/report-grant-1"]; len(got) == 0 || got[len(got)-1] != http.MethodDelete {
		t.Fatalf("dedicated delete = %v", methods)
	}
}

func TestReportGrantResourceTaggedUsesGenericEndpoints(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	grant := map[string]any{
		"grantId": "tagged-report-grant-1", "inodeId": "report-1", "organizationId": "org-1",
		"memberId": nil, "teamId": "team-2", "permission": "edit", "inodeType": "report",
		"createdBy": "member-admin", "updatedBy": "member-admin",
		"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
	}
	mock.Mux.HandleFunc("/v2/grants", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		var payload map[string]any
		_ = json.NewDecoder(request.Body).Decode(&payload)
		if payload["tagId"] != "tag-9" {
			t.Errorf("tagId = %#v, want tag-9", payload["tagId"])
		}
		writeJSON(response, grant)
	})
	mock.Mux.HandleFunc("/v2/grants/tagged-report-grant-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodGet:
			writeJSON(response, grant)
		case http.MethodDelete:
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/reports/report-1/grants", func(http.ResponseWriter, *http.Request) {
		t.Error("tagged grant must not use dedicated report grant endpoints")
	})
	config := providerConfig(mock) + `
resource "sigma_report_grant" "test" {
  inode_id   = "report-1"
  team_id    = "team-2"
  permission = "edit"
  tag_id     = "tag-9"
}
`
	resource.UnitTest(t, providerTestCase([]resource.TestStep{
		{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_report_grant.test", "id", "tagged-report-grant-1"),
				resource.TestCheckResourceAttr("sigma_report_grant.test", "tag_id", "tag-9"),
			),
		},
		{Config: config},
	}))
}

func TestReportGrantResourceAmbiguity(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	grant := func(id string) map[string]any {
		return map[string]any{
			"grantId": id, "inodeId": "report-1", "organizationId": "org-1",
			"memberId": nil, "teamId": "team-2", "permission": "edit", "inodeType": "report",
			"createdBy": "member-admin", "updatedBy": "member-admin",
			"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
		}
	}
	mock.Mux.HandleFunc("/v2/reports/report-1/grants", func(response http.ResponseWriter, _ *http.Request) {
		writeJSON(response, map[string]any{})
	})
	mock.Mux.HandleFunc("/v2/grants", func(response http.ResponseWriter, _ *http.Request) {
		writeJSON(response, map[string]any{"entries": []any{grant("g-1"), grant("g-2")}, "nextPage": nil})
	})
	config := providerConfig(mock) + `
resource "sigma_report_grant" "test" {
  inode_id   = "report-1"
  team_id    = "team-2"
  permission = "edit"
}
`
	resource.UnitTest(t, providerTestCase([]resource.TestStep{{
		Config:      config,
		ExpectError: regexp.MustCompile(`refusing to select`),
	}}))
}

func TestAccReportGrantResource(t *testing.T) { requireAcceptance(t) }
