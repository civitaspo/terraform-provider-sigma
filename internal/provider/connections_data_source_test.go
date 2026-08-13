package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestConnectionsDataSource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	mock.Mux.HandleFunc("/v2/connections", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{"entries": []map[string]any{{"connectionId": "connection-1", "name": "warehouse", "type": "snowflake", "description": map[string]any{}, "poolSizes": map[string]any{}, "friendlyName": false, "useOauth": true}}, "nextPage": nil})
	})
	config := connectionProviderConfig(mock) + `
data "sigma_connections" "all" {}
`
	resource.UnitTest(t, connectionTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.sigma_connections.all", "connections.#", "1"),
			resource.TestCheckResourceAttr("data.sigma_connections.all", "connections.0.use_oauth", "true"),
		),
	}}))
}
