package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestGrantResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	write := func(response http.ResponseWriter, value any) {
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(value)
	}
	grant := func(id, inode, member, team, permission, inodeType string) map[string]any {
		return map[string]any{
			"grantId": id, "inodeId": inode, "organizationId": "org-1", "memberId": nullable(member), "teamId": nullable(team),
			"permission": permission, "inodeType": inodeType, "createdBy": "member-admin", "updatedBy": "member-admin",
			"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
		}
	}
	genericGrant := grant("grant-1", "folder-1", "", "team-1", "organize", "folder")
	mock.Mux.HandleFunc("/v2/grants", func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPost {
			write(response, genericGrant)
			return
		}
		write(response, map[string]any{"entries": []any{genericGrant}, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/grants/grant-1", func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet {
			write(response, genericGrant)
			return
		}
		write(response, genericGrant)
	})
	config := providerConfig(mock) + `
resource "sigma_grant" "test" {
  inode_id   = "folder-1"
  team_id    = "team-1"
  permission = "organize"
}
`
	resource.UnitTest(t, providerTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("sigma_grant.test", "id", "grant-1"),
		),
	}}))
}

func TestAccGrantResource(t *testing.T) { requireAcceptance(t) }
