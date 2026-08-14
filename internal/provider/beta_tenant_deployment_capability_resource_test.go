package provider_test

import (
	"encoding/json"
	"net/http"
	"regexp"
	"sync"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestTenantDeploymentCapabilityResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	var mu sync.Mutex
	capabilities := map[string]bool{}
	mock.Mux.HandleFunc("/v2/tenants/tenant-1/capabilities/deployments", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodGet {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		mu.Lock()
		defer mu.Unlock()
		entries := make([]any, 0, len(capabilities))
		for id := range capabilities {
			entries = append(entries, map[string]any{"tenantOrganizationId": id})
		}
		writeJSON(response, map[string]any{"entries": entries, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/tenants/tenant-1/capabilities/deployments:batchAdd", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		var body struct {
			IDs []string `json:"deployToTenantOrganizationIds"`
		}
		_ = json.NewDecoder(request.Body).Decode(&body)
		if len(body.IDs) != 1 || body.IDs[0] != "tenant-2" {
			t.Errorf("unexpected add body: %#v", body.IDs)
		}
		mu.Lock()
		capabilities["tenant-2"] = true
		mu.Unlock()
		writeJSON(response, map[string]any{})
	})
	mock.Mux.HandleFunc("/v2/tenants/tenant-1/capabilities/deployments:batchRemove", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		var body struct {
			IDs []string `json:"deployToTenantOrganizationIds"`
		}
		_ = json.NewDecoder(request.Body).Decode(&body)
		mu.Lock()
		for _, id := range body.IDs {
			delete(capabilities, id)
		}
		mu.Unlock()
		writeJSON(response, map[string]any{})
	})
	config := betaProviderConfig(mock) + `
resource "sigma_tenant_deployment_capability" "test" {
  tenant_id  = "tenant-1"
  capability = "tenant-2"
}
`
	resource.UnitTest(t, betaTestCase([]resource.TestStep{
		{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_tenant_deployment_capability.test", "id", "tenant-1/tenant-2"),
			),
		},
		{
			ResourceName:      "sigma_tenant_deployment_capability.test",
			ImportState:       true,
			ImportStateId:     "tenant-1/tenant-2",
			ImportStateVerify: true,
		},
	}))
}

func TestTenantDeploymentCapabilityResourceRead404RemovesState(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	var mu sync.Mutex
	present := true
	mock.Mux.HandleFunc("/v2/tenants/tenant-1/capabilities/deployments", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		mu.Lock()
		defer mu.Unlock()
		if !present {
			writeJSON(response, map[string]any{"entries": []any{}, "nextPage": nil})
			return
		}
		writeJSON(response, map[string]any{
			"entries":  []any{map[string]any{"tenantOrganizationId": "tenant-2"}},
			"nextPage": nil,
		})
	})
	mock.Mux.HandleFunc("/v2/tenants/tenant-1/capabilities/deployments:batchAdd", func(response http.ResponseWriter, request *http.Request) {
		writeJSON(response, map[string]any{})
	})
	mock.Mux.HandleFunc("/v2/tenants/tenant-1/capabilities/deployments:batchRemove", func(response http.ResponseWriter, request *http.Request) {
		writeJSON(response, map[string]any{})
	})
	config := betaProviderConfig(mock) + `
resource "sigma_tenant_deployment_capability" "test" {
  tenant_id  = "tenant-1"
  capability = "tenant-2"
}
`
	resource.UnitTest(t, betaTestCase([]resource.TestStep{
		{Config: config},
		{
			PreConfig:          func() { mu.Lock(); present = false; mu.Unlock() },
			RefreshState:       true,
			ExpectNonEmptyPlan: true,
		},
	}))
}

func TestTenantDeploymentCapabilityResourceDuplicateAdd(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	var mu sync.Mutex
	adds := 0
	mock.Mux.HandleFunc("/v2/tenants/tenant-1/capabilities/deployments", func(response http.ResponseWriter, request *http.Request) {
		writeJSON(response, map[string]any{
			"entries":  []any{map[string]any{"tenantOrganizationId": "tenant-2"}},
			"nextPage": nil,
		})
	})
	mock.Mux.HandleFunc("/v2/tenants/tenant-1/capabilities/deployments:batchAdd", func(response http.ResponseWriter, request *http.Request) {
		mu.Lock()
		adds++
		n := adds
		mu.Unlock()
		if n > 1 {
			http.Error(response, "capability already exists", http.StatusConflict)
			return
		}
		writeJSON(response, map[string]any{})
	})
	mock.Mux.HandleFunc("/v2/tenants/tenant-1/capabilities/deployments:batchRemove", func(response http.ResponseWriter, request *http.Request) {
		writeJSON(response, map[string]any{})
	})
	config := betaProviderConfig(mock) + `
resource "sigma_tenant_deployment_capability" "a" {
  tenant_id  = "tenant-1"
  capability = "tenant-2"
}
resource "sigma_tenant_deployment_capability" "b" {
  tenant_id  = "tenant-1"
  capability = "tenant-2"
}
`
	resource.UnitTest(t, betaTestCase([]resource.TestStep{{
		Config:      config,
		ExpectError: regexp.MustCompile(`already exists|Unable to add Sigma tenant deployment capability`),
	}}))
}

func TestAccTenantDeploymentCapabilityResource(t *testing.T) { requireAcceptance(t) }
