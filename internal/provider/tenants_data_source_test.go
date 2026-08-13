package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestTenantsDataSource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	tenant := map[string]any{
		"tenantOrganizationId": "tenant-1", "parentOrganizationId": "parent-1",
		"createdBy": "user-1", "updatedBy": "user-1",
		"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
		"tenantOrganizationName": "Acme", "tenantOrganizationSlug": "acme",
	}
	mock.Mux.HandleFunc("/v2/tenants", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{"entries": []any{tenant}})
	})
	config := betaProviderConfig(mock) + `
data "sigma_tenants" "test" {}
`
	resource.UnitTest(t, betaTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.sigma_tenants.test", "tenants.#", "1"),
			resource.TestCheckResourceAttr("data.sigma_tenants.test", "tenants.0.id", "tenant-1"),
		),
	}}))
}

func TestAccTenantsDataSource(t *testing.T) { requireAcceptance(t) }
