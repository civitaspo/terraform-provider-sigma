package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestTeamsDataSource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	team := map[string]any{
		"teamId": "team-1", "name": "Analytics", "description": "Core", "visibility": "private", "isArchived": false, "workspaceId": "workspace-1",
	}
	write := func(response http.ResponseWriter, value any) {
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(value)
	}
	mock.Mux.HandleFunc("/v2.1/teams", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		write(response, map[string]any{"entries": []any{team}, "nextPage": nil})
	})
	config := identityProviderConfig(mock) + `
data "sigma_teams" "all" {}
`
	resource.UnitTest(t, identityTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.sigma_teams.all", "teams.#", "1"),
			resource.TestCheckResourceAttr("data.sigma_teams.all", "teams.0.workspace_id", "workspace-1"),
		),
	}}))
}
