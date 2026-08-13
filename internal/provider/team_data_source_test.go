package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestTeamDataSource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	team := map[string]any{
		"teamId": "team-1", "name": "Analytics", "description": "Core", "visibility": "private", "isArchived": false, "workspaceId": "workspace-1",
	}
	write := func(response http.ResponseWriter, value any) {
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(value)
	}
	mock.Mux.HandleFunc("/v2/teams/team-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		write(response, team)
	})
	config := identityProviderConfig(mock) + `
data "sigma_team" "one" {
  id = "team-1"
}
`
	resource.UnitTest(t, identityTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.sigma_team.one", "name", "Analytics"),
			resource.TestCheckResourceAttr("data.sigma_team.one", "workspace_id", "workspace-1"),
		),
	}}))
}
