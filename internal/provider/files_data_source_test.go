package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestFilesDataSource(t *testing.T) {
	file := map[string]any{
		"id": "folder-1", "urlId": "file-url-1", "name": "Managed", "type": "folder",
		"parentId": "workspace-1", "parentUrlId": "url-1", "permission": "edit", "path": "Analytics/Managed",
		"badge": nil, "isArchived": false, "description": "", "ownerId": nil,
		"createdBy": "member-1", "updatedBy": "member-1", "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
	}
	runListDataSourceCases(t, listDataSourceCase{
		path: "/v2/files",
		config: `
data "sigma_files" "test" {
  parent_id            = "workspace-1"
  direct_children_only = true
  type_filters         = ["folder"]
  permission           = "edit"
  name                 = "Managed"
}
`,
		wantQuery: map[string]string{
			"parentId":          "workspace-1",
			"directChildFilter": "true",
			"typeFilters":       "folder",
			"permissionFilter":  "edit",
			"name":              "Managed",
		},
		entry:     file,
		address:   "data.sigma_files.test",
		countAttr: "files.#",
	})
}

func TestFilesDataSourceNoFilters(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	file := map[string]any{
		"id": "folder-1", "urlId": "file-url-1", "name": "Managed", "type": "folder",
		"parentId": "workspace-1", "parentUrlId": "url-1", "permission": "edit", "path": "Analytics/Managed",
		"badge": nil, "isArchived": false, "description": "", "ownerId": nil,
		"createdBy": "member-1", "updatedBy": "member-1", "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
	}
	mock.Mux.HandleFunc("/v2/files", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		assertExactQuery(t, request, map[string]string{}, "page", "limit", "pageSize")
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{"entries": []any{file}, "nextPage": nil})
	})
	resource.UnitTest(t, providerTestCase([]resource.TestStep{{
		Config: providerConfig(mock) + `data "sigma_files" "test" {}`,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.sigma_files.test", "files.#", "1"),
			resource.TestCheckResourceAttr("data.sigma_files.test", "files.0.id", "folder-1"),
		),
	}}))
}
