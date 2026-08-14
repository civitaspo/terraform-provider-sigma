package provider_test

import "testing"

func TestTenantDataSource(t *testing.T) {
	runSingularDataSourceCases(t, singularDataSourceCase{
		path: "/v2/tenants/tenant-1",
		config: `
data "sigma_tenant" "one" {
  id = "tenant-1"
}
`,
		entry: map[string]any{
			"tenantOrganizationId": "tenant-1", "parentOrganizationId": "parent-1",
			"createdBy": "user-1", "updatedBy": "user-1",
			"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
			"tenantOrganizationName": "Acme", "tenantOrganizationSlug": "acme",
		},
		address:   "data.sigma_tenant.one",
		checkAttr: "tenant_organization_name",
		want:      "Acme",
	})
}

func TestAccTenantDataSource(t *testing.T) { runAccReadOnlyCatalog(t) }
