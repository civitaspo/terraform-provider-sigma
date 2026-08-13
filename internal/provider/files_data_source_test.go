package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestFilesDataSource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	file := map[string]any{
		"id": "folder-1", "urlId": "file-url-1", "name": "Managed", "type": "folder",
		"parentId": "workspace-1", "parentUrlId": "url-1", "permission": "edit", "path": "Analytics/Managed",
		"badge": nil, "isArchived": false, "description": "", "ownerId": nil,
		"createdBy": "member-1", "updatedBy": "member-1", "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
	}
	write := func(response http.ResponseWriter, value any) {
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(value)
	}
	mock.Mux.HandleFunc("/v2/files", func(response http.ResponseWriter, request *http.Request) {
		if got := request.URL.Query().Get("parentId"); got != "workspace-1" {
			t.Errorf("parentId = %q", got)
		}
		write(response, map[string]any{"entries": []any{file}, "nextPage": nil})
	})
	config := `
provider "sigma" {
  base_url     = "` + mock.URL() + `"
  client_id     = "` + mock.ClientID + `"
  client_secret = "` + mock.ClientSecret + `"
}

data "sigma_files" "test" {
  parent_id            = "workspace-1"
  direct_children_only = true
  type_filters         = ["folder"]
}
`
	resource.UnitTest(t, providerTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.sigma_files.test", "files.0.id", "folder-1"),
		),
	}}))
}
