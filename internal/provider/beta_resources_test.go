package provider_test

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	sigmaprovider "github.com/civitaspo/terraform-provider-sigma/internal/provider"
	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func betaProviderConfig(mock *testutil.MockSigma) string {
	return `
provider "sigma" {
  base_url      = "` + mock.URL() + `"
  client_id     = "` + mock.ClientID + `"
  client_secret = "` + mock.ClientSecret + `"
}
`
}

func betaTestCase(steps []resource.TestStep) resource.TestCase {
	return resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: steps,
	}
}

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
	resource.UnitTest(t, betaTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("sigma_tenant.test", "id", "tenant-1"),
			resource.TestCheckResourceAttr("sigma_tenant.test", "tenant_organization_slug", "acme"),
			resource.TestCheckResourceAttr("sigma_tenant.test", "tenant_cloud_provider", "aws"),
		),
	}}))
}

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

func TestDeploymentPolicyResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	policy := map[string]any{
		"deploymentPolicyId": "policy-1", "name": "Starter", "nameInTenant": "Starter",
		"versionTagId": "tag-1", "sourceSwapPolicies": []string{"swap-1"}, "copyInputTableData": false,
	}
	inodes := []string{}
	tenants := []string{}
	mock.Mux.HandleFunc("/v2/deploymentPolicies", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodPost:
			_ = json.NewEncoder(response).Encode(map[string]any{"deploymentPolicyId": "policy-1"})
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(map[string]any{"entries": []any{policy}})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/deploymentPolicies/policy-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet, http.MethodPatch:
			_ = json.NewEncoder(response).Encode(policy)
		case http.MethodDelete:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/deploymentPolicies/policy-1/files", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
			entries := make([]any, 0, len(inodes))
			for _, id := range inodes {
				entries = append(entries, map[string]any{"inodeId": id, "inode": nil, "deploymentPolicyId": "policy-1"})
			}
			_ = json.NewEncoder(response).Encode(map[string]any{"entries": entries})
		case http.MethodPost:
			var body struct {
				InodeIDs []string `json:"inodeIds"`
			}
			_ = json.NewDecoder(request.Body).Decode(&body)
			inodes = append(inodes, body.InodeIDs...)
			_ = json.NewEncoder(response).Encode(map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/deploymentPolicies/policy-1/files/", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		if request.Method != http.MethodDelete {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimPrefix(request.URL.Path, "/v2/deploymentPolicies/policy-1/files/")
		next := inodes[:0]
		for _, inode := range inodes {
			if inode != id {
				next = append(next, inode)
			}
		}
		inodes = next
		_ = json.NewEncoder(response).Encode(map[string]any{})
	})
	mock.Mux.HandleFunc("/v2/deploymentPolicies/policy-1/tenants", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(map[string]any{"entries": tenants})
		case http.MethodPost:
			var body struct {
				TenantOrganizationID string `json:"tenantOrganizationId"`
			}
			_ = json.NewDecoder(request.Body).Decode(&body)
			tenants = append(tenants, body.TenantOrganizationID)
			_ = json.NewEncoder(response).Encode(map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/deploymentPolicies/policy-1/tenants/", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		if request.Method != http.MethodDelete {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimPrefix(request.URL.Path, "/v2/deploymentPolicies/policy-1/tenants/")
		next := tenants[:0]
		for _, tenant := range tenants {
			if tenant != id {
				next = append(next, tenant)
			}
		}
		tenants = next
		_ = json.NewEncoder(response).Encode(map[string]any{})
	})
	config := betaProviderConfig(mock) + `
resource "sigma_deployment_policy" "test" {
  name                 = "Starter"
  version_tag_id       = "tag-1"
  source_swap_policies = ["swap-1"]
  inode_ids            = ["inode-1"]
  tenant_ids           = ["tenant-1"]
}
`
	resource.UnitTest(t, betaTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("sigma_deployment_policy.test", "id", "policy-1"),
			resource.TestCheckResourceAttr("sigma_deployment_policy.test", "inode_ids.#", "1"),
			resource.TestCheckResourceAttr("sigma_deployment_policy.test", "tenant_ids.#", "1"),
		),
	}}))
}

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
