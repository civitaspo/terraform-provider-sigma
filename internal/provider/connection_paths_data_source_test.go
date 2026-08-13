package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestConnectionPathsDataSource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	mock.Mux.HandleFunc("/v2/connection/connection-1/lookup", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{"kind": "table", "inodeId": "path-1", "url": "/connection/path-1"})
	})
	mock.Mux.HandleFunc("/v2/connections/paths", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{"entries": []map[string]any{{"connectionId": "connection-1", "urlId": "path-1", "path": []string{"DATABASE", "SCHEMA"}}}, "nextPage": nil})
	})
	config := connectionProviderConfig(mock) + `
data "sigma_connection_paths" "lookup" {
  connection_id = "connection-1"
  path          = ["DATABASE", "SCHEMA", "TABLE"]
}
data "sigma_connection_paths" "all" {
  connection_id = "connection-1"
}
`
	resource.UnitTest(t, connectionTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.sigma_connection_paths.lookup", "inode_id", "path-1"),
			resource.TestCheckResourceAttr("data.sigma_connection_paths.all", "paths.#", "1"),
		),
	}}))
}
