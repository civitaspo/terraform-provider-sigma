package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestReportGrantResource(t *testing.T) {
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
	reportGrant := grant("report-grant-1", "report-1", "", "team-2", "edit", "report")
	mock.Mux.HandleFunc("/v2/grants", func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPost {
			write(response, reportGrant)
			return
		}
		write(response, map[string]any{"entries": []any{reportGrant}, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/reports/report-1/grants", func(response http.ResponseWriter, _ *http.Request) {
		write(response, map[string]any{})
	})
	mock.Mux.HandleFunc("/v2/reports/report-1/grants/report-grant-1", func(response http.ResponseWriter, _ *http.Request) {
		write(response, map[string]any{})
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
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("sigma_report_grant.test", "id", "report-grant-1"),
		),
	}}))
}

func TestAccReportGrantResource(t *testing.T) { requireAcceptance(t) }
