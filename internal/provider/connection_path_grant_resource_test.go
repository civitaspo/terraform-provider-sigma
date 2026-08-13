package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestConnectionPathGrantResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	mock.Mux.HandleFunc("/v2/connections/paths/path-1/grants", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodPost:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(map[string]any{
				"entries":  []map[string]any{{"grantId": "grant-1", "inodeId": "path-1", "memberId": "member-1", "teamId": nil, "permission": "usage"}},
				"nextPage": nil,
			})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/connections/paths/path-1/grants/grant-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodDelete {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{})
	})
	address := "sigma_connection_path_grant.test"
	config := connectionProviderConfig(mock) + `
resource "sigma_connection_path_grant" "test" {
  connection_path_id = "path-1"
  member_id = "member-1"
  permission = "usage"
}
`
	resource.UnitTest(t, connectionTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(address, "id", "grant-1"),
			resource.TestCheckResourceAttr(address, "permission", "usage"),
		),
	}}))
}

func TestAccConnectionPathGrantResource(t *testing.T) { requireAcceptance(t) }
