package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDeploymentPoliciesDataSource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	policy := map[string]any{
		"deploymentPolicyId": "policy-1", "name": "Starter", "nameInTenant": "Starter",
		"versionTagId": "tag-1", "sourceSwapPolicies": []string{}, "copyInputTableData": true,
	}
	mock.Mux.HandleFunc("/v2/deploymentPolicies", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{"entries": []any{policy}})
	})
	config := betaProviderConfig(mock) + `
data "sigma_deployment_policies" "test" {}
`
	resource.UnitTest(t, betaTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.sigma_deployment_policies.test", "deployment_policies.#", "1"),
			resource.TestCheckResourceAttr("data.sigma_deployment_policies.test", "deployment_policies.0.id", "policy-1"),
		),
	}}))
}

func TestAccDeploymentPoliciesDataSource(t *testing.T) { requireAcceptance(t) }
