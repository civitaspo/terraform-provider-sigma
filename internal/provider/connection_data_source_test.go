package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestConnectionDataSource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	mock.Mux.HandleFunc("/v2/connections/connection-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{"connectionId": "connection-1", "name": "warehouse", "type": "snowflake", "description": map[string]any{}, "poolSizes": map[string]any{}, "friendlyName": false, "useOauth": true})
	})
	config := connectionProviderConfig(mock) + `
data "sigma_connection" "one" {
  id = "connection-1"
}
`
	resource.UnitTest(t, connectionTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.sigma_connection.one", "type", "snowflake"),
			resource.TestCheckResourceAttr("data.sigma_connection.one", "use_oauth", "true"),
		),
	}}))
}
