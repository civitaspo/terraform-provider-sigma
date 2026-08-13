package provider_test

import (
	"encoding/json"
	"net/http"
	"strings"
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

func TestDeploymentPolicyResourceOmitsAttachmentSyncWhenNull(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	policy := map[string]any{
		"deploymentPolicyId": "policy-1", "name": "Starter", "nameInTenant": "Starter",
		"versionTagId": "tag-1", "sourceSwapPolicies": []string{}, "copyInputTableData": false,
	}
	inodes := []string{"inode-existing"}
	tenants := []string{"tenant-existing"}
	deleteCalls := 0
	mock.Mux.HandleFunc("/v2/deploymentPolicies", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
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
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		entries := make([]any, 0, len(inodes))
		for _, id := range inodes {
			entries = append(entries, map[string]any{"inodeId": id})
		}
		_ = json.NewEncoder(response).Encode(map[string]any{"entries": entries})
	})
	mock.Mux.HandleFunc("/v2/deploymentPolicies/policy-1/files/", func(response http.ResponseWriter, request *http.Request) {
		deleteCalls++
		http.Error(response, "should not delete unmanaged attachments", http.StatusInternalServerError)
	})
	mock.Mux.HandleFunc("/v2/deploymentPolicies/policy-1/tenants", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{"entries": tenants})
	})
	mock.Mux.HandleFunc("/v2/deploymentPolicies/policy-1/tenants/", func(response http.ResponseWriter, request *http.Request) {
		deleteCalls++
		http.Error(response, "should not delete unmanaged attachments", http.StatusInternalServerError)
	})
	config := betaProviderConfig(mock) + `
resource "sigma_deployment_policy" "test" {
  name           = "Starter"
  version_tag_id = "tag-1"
}
`
	resource.UnitTest(t, betaTestCase([]resource.TestStep{{
		Config: config,
		Check:  resource.TestCheckResourceAttr("sigma_deployment_policy.test", "id", "policy-1"),
	}}))
	if deleteCalls != 0 {
		t.Fatalf("unexpected attachment delete calls: %d", deleteCalls)
	}
}

func TestAccDeploymentPolicyResource(t *testing.T) { requireAcceptance(t) }
