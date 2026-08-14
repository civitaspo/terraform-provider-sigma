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

func TestDeploymentPolicyResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	policy := map[string]any{
		"deploymentPolicyId": "policy-1", "name": "Starter", "nameInTenant": "Starter",
		"versionTagId": "tag-1", "sourceSwapPolicies": []string{"swap-1"}, "copyInputTableData": false,
	}
	mock.Mux.HandleFunc("/v2/deploymentPolicies", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		_ = json.NewEncoder(response).Encode(map[string]any{"deploymentPolicyId": "policy-1"})
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
		t.Errorf("parent policy Create must not attach documents: %s %s", request.Method, request.URL.Path)
		http.Error(response, "unexpected attachment", http.StatusInternalServerError)
	})
	mock.Mux.HandleFunc("/v2/deploymentPolicies/policy-1/tenants", func(response http.ResponseWriter, request *http.Request) {
		t.Errorf("parent policy Create must not attach tenants: %s %s", request.Method, request.URL.Path)
		http.Error(response, "unexpected attachment", http.StatusInternalServerError)
	})
	config := betaProviderConfig(mock) + `
resource "sigma_deployment_policy" "test" {
  name                 = "Starter"
  version_tag_id       = "tag-1"
  source_swap_policies = ["swap-1"]
}
`
	resource.UnitTest(t, betaTestCase([]resource.TestStep{
		{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_deployment_policy.test", "id", "policy-1"),
				resource.TestCheckResourceAttr("sigma_deployment_policy.test", "source_swap_policies.#", "1"),
			),
		},
		{
			ResourceName:      "sigma_deployment_policy.test",
			ImportState:       true,
			ImportStateVerify: true,
		},
	}))
}

func TestDeploymentPolicyRelationFailureDoesNotOrphanParent(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	var mu sync.Mutex
	created := false
	policy := map[string]any{
		"deploymentPolicyId": "policy-1", "name": "Starter", "nameInTenant": "Starter",
		"versionTagId": "tag-1", "sourceSwapPolicies": []string{}, "copyInputTableData": false,
	}
	mock.Mux.HandleFunc("/v2/deploymentPolicies", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		mu.Lock()
		created = true
		mu.Unlock()
		writeJSON(response, map[string]any{"deploymentPolicyId": "policy-1"})
	})
	mock.Mux.HandleFunc("/v2/deploymentPolicies/policy-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodGet, http.MethodPatch:
			writeJSON(response, policy)
		case http.MethodDelete:
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/deploymentPolicies/policy-1/files", func(response http.ResponseWriter, request *http.Request) {
		http.Error(response, "inode attach failed", http.StatusInternalServerError)
	})
	config := betaProviderConfig(mock) + `
resource "sigma_deployment_policy" "test" {
  name           = "Starter"
  version_tag_id = "tag-1"
}
resource "sigma_deployment_policy_document" "test" {
  deployment_policy_id = sigma_deployment_policy.test.id
  inode_id             = "inode-1"
}
`
	resource.UnitTest(t, betaTestCase([]resource.TestStep{{
		Config:      config,
		ExpectError: regexp.MustCompile(`inode attach failed|Unable to add Sigma deployment policy document`),
	}}))
	mu.Lock()
	defer mu.Unlock()
	if !created {
		t.Fatal("expected parent deployment policy to be created before relation failure")
	}
}

func TestAccDeploymentPolicyResource(t *testing.T) { requireAcceptance(t) }
