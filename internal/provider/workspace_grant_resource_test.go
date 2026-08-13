package provider_test

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestWorkspaceGrantResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	workspaceGrant := map[string]any{
		"grantId": "workspace-grant-1", "inodeId": "workspace-1", "organizationId": "org-1",
		"memberId": "member-1", "teamId": nil, "permission": "view", "inodeType": "workspace",
		"createdBy": "member-admin", "updatedBy": "member-admin",
		"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
	}
	mock.Mux.HandleFunc("/v2/workspaces/workspace-1/grants", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method == http.MethodPost {
			writeJSON(response, map[string]any{})
			return
		}
		writeJSON(response, map[string]any{"entries": []any{workspaceGrant}, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/workspaces/workspace-1/grants/workspace-grant-1", func(response http.ResponseWriter, _ *http.Request) {
		writeJSON(response, map[string]any{})
	})
	config := providerConfig(mock) + `
resource "sigma_workspace_grant" "test" {
  inode_id   = "workspace-1"
  member_id  = "member-1"
  permission = "view"
}
`
	resource.UnitTest(t, providerTestCase([]resource.TestStep{{
		Config: config,
		Check:  resource.TestCheckResourceAttr("sigma_workspace_grant.test", "id", "workspace-grant-1"),
	}}))
}

func TestWorkspaceGrantResourceAmbiguity(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	grant := func(id string) map[string]any {
		return map[string]any{
			"grantId": id, "inodeId": "workspace-1", "organizationId": "org-1",
			"memberId": "member-1", "teamId": nil, "permission": "view", "inodeType": "workspace",
			"createdBy": "member-admin", "updatedBy": "member-admin",
			"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
		}
	}
	mock.Mux.HandleFunc("/v2/workspaces/workspace-1/grants", func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPost {
			writeJSON(response, map[string]any{})
			return
		}
		writeJSON(response, map[string]any{"entries": []any{grant("g-1"), grant("g-2")}, "nextPage": nil})
	})
	config := providerConfig(mock) + `
resource "sigma_workspace_grant" "test" {
  inode_id   = "workspace-1"
  member_id  = "member-1"
  permission = "view"
}
`
	resource.UnitTest(t, providerTestCase([]resource.TestStep{{
		Config:      config,
		ExpectError: regexp.MustCompile(`refusing to select`),
	}}))
}

func TestAccWorkspaceGrantResource(t *testing.T) { requireAcceptance(t) }
