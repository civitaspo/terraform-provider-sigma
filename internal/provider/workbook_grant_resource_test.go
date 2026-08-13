package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestWorkbookGrantResource(t *testing.T) {
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
	workbookGrant := grant("workbook-grant-1", "workbook-1", "member-2", "", "explore", "workbook")
	mock.Mux.HandleFunc("/v2/grants", func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPost {
			write(response, workbookGrant)
			return
		}
		write(response, map[string]any{"entries": []any{workbookGrant}, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/workbooks/workbook-1/grants", func(response http.ResponseWriter, _ *http.Request) {
		write(response, map[string]any{})
	})
	mock.Mux.HandleFunc("/v2/workbooks/workbook-1/grants/workbook-grant-1", func(response http.ResponseWriter, _ *http.Request) {
		write(response, map[string]any{})
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
		),
	}}))
}

func TestAccWorkbookGrantResource(t *testing.T) { requireAcceptance(t) }
