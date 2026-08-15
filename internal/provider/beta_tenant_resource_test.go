package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestTenantResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	tenant := map[string]any{
		"tenantOrganizationId": "tenant-1", "parentOrganizationId": "parent-1",
		"createdBy": "user-1", "updatedBy": "user-1",
		"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
		"tenantOrganizationName": "Acme", "tenantOrganizationSlug": "acme",
		"tenantCloudProvider": "aws", "tenantRegion": "us-west-2", "tenantApiUrl": "https://tenant.example",
	}
	mock.Mux.HandleFunc("/v2/tenants", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodPost:
			_ = json.NewEncoder(response).Encode(tenant)
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(map[string]any{"entries": []any{tenant}})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/tenants/tenant-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet, http.MethodPatch:
			if request.Method == http.MethodPatch {
				var body map[string]any
				_ = json.NewDecoder(request.Body).Decode(&body)
				if name, ok := body["tenantOrganizationName"].(string); ok {
					tenant["tenantOrganizationName"] = name
				}
			}
			_ = json.NewEncoder(response).Encode(tenant)
		case http.MethodDelete:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	config := betaProviderConfig(mock) + `
resource "sigma_tenant" "test" {
  tenant_organization_name = "Acme"
  tenant_organization_slug = "acme"
}
`
	resource.UnitTest(t, betaTestCase([]resource.TestStep{
		{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_tenant.test", "id", "tenant-1"),
				resource.TestCheckResourceAttr("sigma_tenant.test", "tenant_organization_slug", "acme"),
				resource.TestCheckResourceAttr("sigma_tenant.test", "tenant_cloud_provider", "aws"),
			),
		},
		{
			ResourceName:      "sigma_tenant.test",
			ImportState:       true,
			ImportStateVerify: true,
		},
		{
			Config: betaProviderConfig(mock) + `
resource "sigma_tenant" "test" {
  tenant_organization_name = "Acme Renamed"
  tenant_organization_slug = "acme"
}
`,
			Check: resource.TestCheckResourceAttr("sigma_tenant.test", "tenant_organization_name", "Acme Renamed"),
		},
	}))
}

func TestAccTenantResource(t *testing.T) {
	requireAcceptance(t)
	t.Skip("tenant organization create is skipped in production")
}
