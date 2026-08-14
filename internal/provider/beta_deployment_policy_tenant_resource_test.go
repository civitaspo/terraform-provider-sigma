package provider_test

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDeploymentPolicyTenantResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	var mu sync.Mutex
	tenants := map[string]bool{}
	mock.Mux.HandleFunc("/v2/deploymentPolicies/policy-1/tenants", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		mu.Lock()
		defer mu.Unlock()
		switch request.Method {
		case http.MethodGet:
			entries := make([]any, 0, len(tenants))
			for id := range tenants {
				entries = append(entries, id)
			}
			writeJSON(response, map[string]any{"entries": entries, "nextPage": nil})
		case http.MethodPost:
			var body struct {
				TenantOrganizationID string `json:"tenantOrganizationId"`
			}
			_ = json.NewDecoder(request.Body).Decode(&body)
			if body.TenantOrganizationID != "tenant-1" {
				t.Errorf("unexpected tenant add: %s", body.TenantOrganizationID)
			}
			tenants["tenant-1"] = true
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/deploymentPolicies/policy-1/tenants/", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodDelete {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimPrefix(request.URL.Path, "/v2/deploymentPolicies/policy-1/tenants/")
		mu.Lock()
		delete(tenants, id)
		mu.Unlock()
		writeJSON(response, map[string]any{})
	})
	config := betaProviderConfig(mock) + `
resource "sigma_deployment_policy_tenant" "test" {
  deployment_policy_id = "policy-1"
  tenant_id            = "tenant-1"
}
`
	resource.UnitTest(t, betaTestCase([]resource.TestStep{
		{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_deployment_policy_tenant.test", "id", "policy-1/tenant-1"),
			),
		},
		{
			ResourceName:      "sigma_deployment_policy_tenant.test",
			ImportState:       true,
			ImportStateId:     "policy-1/tenant-1",
			ImportStateVerify: true,
		},
	}))
}

func TestDeploymentPolicyTenantResourceReadMissingRemovesState(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	var mu sync.Mutex
	present := true
	mock.Mux.HandleFunc("/v2/deploymentPolicies/policy-1/tenants", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodPost:
			writeJSON(response, map[string]any{})
		case http.MethodGet:
			mu.Lock()
			ok := present
			mu.Unlock()
			if !ok {
				writeJSON(response, map[string]any{"entries": []any{}, "nextPage": nil})
				return
			}
			writeJSON(response, map[string]any{"entries": []any{"tenant-1"}, "nextPage": nil})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/deploymentPolicies/policy-1/tenants/", func(response http.ResponseWriter, request *http.Request) {
		writeJSON(response, map[string]any{})
	})
	config := betaProviderConfig(mock) + `
resource "sigma_deployment_policy_tenant" "test" {
  deployment_policy_id = "policy-1"
  tenant_id            = "tenant-1"
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

func TestDeploymentPolicyTenantResourceDuplicateAdd(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	var mu sync.Mutex
	adds := 0
	mock.Mux.HandleFunc("/v2/deploymentPolicies/policy-1/tenants", func(response http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodPost:
			mu.Lock()
			adds++
			n := adds
			mu.Unlock()
			if n > 1 {
				http.Error(response, "tenant already attached", http.StatusConflict)
				return
			}
			writeJSON(response, map[string]any{})
		case http.MethodGet:
			writeJSON(response, map[string]any{"entries": []any{"tenant-1"}, "nextPage": nil})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/deploymentPolicies/policy-1/tenants/", func(response http.ResponseWriter, request *http.Request) {
		writeJSON(response, map[string]any{})
	})
	config := betaProviderConfig(mock) + `
resource "sigma_deployment_policy_tenant" "a" {
  deployment_policy_id = "policy-1"
  tenant_id            = "tenant-1"
}
resource "sigma_deployment_policy_tenant" "b" {
  deployment_policy_id = "policy-1"
  tenant_id            = "tenant-1"
}
`
	resource.UnitTest(t, betaTestCase([]resource.TestStep{{
		Config:      config,
		ExpectError: regexp.MustCompile(`already attached|Unable to add Sigma deployment policy tenant`),
	}}))
}

func TestAccDeploymentPolicyTenantResource(t *testing.T) {
	requireAcceptance(t)
	t.Skip("deployment policy tenant attach is skipped without a disposable policy")
}
