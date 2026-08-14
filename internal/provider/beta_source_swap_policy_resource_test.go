package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestSourceSwapPolicyResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	policy := map[string]any{
		"policyId": "swap-1", "type": "deployment", "name": "Swap",
		"fromConnectionId": "conn-1",
		"swaps": map[string]any{
			"toConnection":    map[string]any{"swapType": "attribute", "userAttributeId": "attr-1"},
			"deploymentSwaps": []any{},
		},
	}
	mock.Mux.HandleFunc("/v2/sourceSwapPolicies", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		_ = json.NewEncoder(response).Encode(map[string]any{"policyId": "swap-1"})
	})
	mock.Mux.HandleFunc("/v2/sourceSwapPolicies/swap-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(policy)
		case http.MethodPatch:
			_ = json.NewEncoder(response).Encode(map[string]any{"policyId": "swap-1"})
		case http.MethodDelete:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	config := betaProviderConfig(mock) + `
resource "sigma_source_swap_policy" "test" {
  type               = "deployment"
  name               = "Swap"
  from_connection_id = "conn-1"
  swaps_json = jsonencode({
    toConnection = {
      swapType        = "attribute"
      userAttributeId = "attr-1"
    }
    deploymentSwaps = []
  })
}
`
	resource.UnitTest(t, betaTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("sigma_source_swap_policy.test", "id", "swap-1"),
			resource.TestCheckResourceAttr("sigma_source_swap_policy.test", "type", "deployment"),
		),
	}}))
}

func TestAccSourceSwapPolicyResource(t *testing.T) { requireAcceptance(t) }
