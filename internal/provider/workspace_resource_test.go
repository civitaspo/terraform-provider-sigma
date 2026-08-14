package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestWorkspaceResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	write := func(response http.ResponseWriter, value any) {
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(value)
	}
	workspace := map[string]any{
		"workspaceId": "workspace-1", "workspaceUrlId": "workspace-url-1", "name": "Analytics",
		"createdBy": "member-admin", "updatedBy": "member-admin", "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
	}
	mock.Mux.HandleFunc("/v2/workspaces", func(response http.ResponseWriter, _ *http.Request) { write(response, workspace) })
	mock.Mux.HandleFunc("/v2/workspaces/workspace-1", func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodDelete {
			write(response, map[string]any{})
			return
		}
		if request.Method == http.MethodPatch {
			var body map[string]any
			_ = json.NewDecoder(request.Body).Decode(&body)
			if name, ok := body["name"].(string); ok {
				workspace["name"] = name
			}
		}
		write(response, workspace)
	})
	resource.UnitTest(t, providerTestCase([]resource.TestStep{
		{
			Config: providerConfig(mock) + `
resource "sigma_workspace" "test" {
  name = "Analytics"
}
`,
			Check: resource.TestCheckResourceAttr("sigma_workspace.test", "id", "workspace-1"),
		},
		{
			ResourceName:      "sigma_workspace.test",
			ImportState:       true,
			ImportStateVerify: true,
		},
		{
			Config: providerConfig(mock) + `
resource "sigma_workspace" "test" {
  name = "Analytics Renamed"
}
`,
			Check: resource.TestCheckResourceAttr("sigma_workspace.test", "name", "Analytics Renamed"),
		},
	}))
}

func TestAccWorkspaceResource(t *testing.T) { requireAcceptance(t) }
