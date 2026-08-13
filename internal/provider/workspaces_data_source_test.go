package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestWorkspacesDataSource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	workspace := map[string]any{
		"workspaceId": "workspace-1", "workspaceUrlId": "url-1", "name": "Analytics",
		"createdBy": "member-1", "updatedBy": "member-1", "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
	}
	write := func(response http.ResponseWriter, value any) {
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(value)
	}
	mock.Mux.HandleFunc("/v2.1/workspaces", func(response http.ResponseWriter, _ *http.Request) {
		write(response, map[string]any{"entries": []any{workspace}, "nextPage": nil})
	})
	config := `
provider "sigma" {
  base_url     = "` + mock.URL() + `"
  client_id     = "` + mock.ClientID + `"
  client_secret = "` + mock.ClientSecret + `"
}

data "sigma_workspaces" "test" {}
`
	resource.UnitTest(t, providerTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.sigma_workspaces.test", "workspaces.#", "1"),
		),
	}}))
}
