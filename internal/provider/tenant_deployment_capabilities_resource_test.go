package provider_test

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestTenantDeploymentCapabilitiesResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	caps := []any{map[string]any{"tenantOrganizationId": "tenant-b"}}
	mock.Mux.HandleFunc("/v2/tenants/tenant-a/capabilities/deployments", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{"entries": caps})
	})
	mock.Mux.HandleFunc("/v2/tenants/tenant-a/capabilities/deployments:batchAdd", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		body, _ := io.ReadAll(request.Body)
		if !strings.Contains(string(body), "tenant-b") {
			http.Error(response, "missing tenant", http.StatusBadRequest)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{"capabilities": caps})
	})
	mock.Mux.HandleFunc("/v2/tenants/tenant-a/capabilities/deployments:batchRemove", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{})
	})
	config := betaProviderConfig(mock) + `
resource "sigma_tenant_deployment_capabilities" "test" {
  tenant_id    = "tenant-a"
  capabilities = ["tenant-b"]
}
`
	resource.UnitTest(t, betaTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("sigma_tenant_deployment_capabilities.test", "id", "tenant-a"),
			resource.TestCheckResourceAttr("sigma_tenant_deployment_capabilities.test", "capabilities.#", "1"),
		),
	}}))
}

func TestAccTenantDeploymentCapabilitiesResource(t *testing.T) { requireAcceptance(t) }
