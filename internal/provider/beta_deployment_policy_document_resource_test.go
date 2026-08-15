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

func TestDeploymentPolicyDocumentResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	var mu sync.Mutex
	inodes := map[string]bool{}
	mock.Mux.HandleFunc("/v2/deploymentPolicies/policy-1/files", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		mu.Lock()
		defer mu.Unlock()
		switch request.Method {
		case http.MethodGet:
			entries := make([]any, 0, len(inodes))
			for id := range inodes {
				entries = append(entries, map[string]any{"inodeId": id, "deploymentPolicyId": "policy-1"})
			}
			writeJSON(response, map[string]any{"entries": entries, "nextPage": nil})
		case http.MethodPost:
			var body struct {
				InodeIDs []string `json:"inodeIds"`
			}
			_ = json.NewDecoder(request.Body).Decode(&body)
			if len(body.InodeIDs) != 1 || body.InodeIDs[0] != "inode-1" {
				t.Errorf("unexpected inode add: %#v", body.InodeIDs)
			}
			inodes["inode-1"] = true
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/deploymentPolicies/policy-1/files/", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodDelete {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimPrefix(request.URL.Path, "/v2/deploymentPolicies/policy-1/files/")
		mu.Lock()
		delete(inodes, id)
		mu.Unlock()
		writeJSON(response, map[string]any{})
	})
	config := betaProviderConfig(mock) + `
resource "sigma_deployment_policy_document" "test" {
  deployment_policy_id = "policy-1"
  inode_id             = "inode-1"
}
`
	resource.UnitTest(t, betaTestCase([]resource.TestStep{
		{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_deployment_policy_document.test", "id", "policy-1/inode-1"),
			),
		},
		{
			ResourceName:      "sigma_deployment_policy_document.test",
			ImportState:       true,
			ImportStateId:     "policy-1/inode-1",
			ImportStateVerify: true,
		},
	}))
}

func TestDeploymentPolicyDocumentResourceReadMissingRemovesState(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	var mu sync.Mutex
	present := true
	mock.Mux.HandleFunc("/v2/deploymentPolicies/policy-1/files", func(response http.ResponseWriter, request *http.Request) {
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
			writeJSON(response, map[string]any{
				"entries":  []any{map[string]any{"inodeId": "inode-1", "deploymentPolicyId": "policy-1"}},
				"nextPage": nil,
			})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/deploymentPolicies/policy-1/files/", func(response http.ResponseWriter, request *http.Request) {
		writeJSON(response, map[string]any{})
	})
	config := betaProviderConfig(mock) + `
resource "sigma_deployment_policy_document" "test" {
  deployment_policy_id = "policy-1"
  inode_id             = "inode-1"
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

func TestDeploymentPolicyDocumentResourceDuplicateAdd(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	var mu sync.Mutex
	adds := 0
	mock.Mux.HandleFunc("/v2/deploymentPolicies/policy-1/files", func(response http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodPost:
			mu.Lock()
			adds++
			n := adds
			mu.Unlock()
			if n > 1 {
				http.Error(response, "inode already attached", http.StatusConflict)
				return
			}
			writeJSON(response, map[string]any{})
		case http.MethodGet:
			writeJSON(response, map[string]any{
				"entries":  []any{map[string]any{"inodeId": "inode-1", "deploymentPolicyId": "policy-1"}},
				"nextPage": nil,
			})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/deploymentPolicies/policy-1/files/", func(response http.ResponseWriter, request *http.Request) {
		writeJSON(response, map[string]any{})
	})
	config := betaProviderConfig(mock) + `
resource "sigma_deployment_policy_document" "a" {
  deployment_policy_id = "policy-1"
  inode_id             = "inode-1"
}
resource "sigma_deployment_policy_document" "b" {
  deployment_policy_id = "policy-1"
  inode_id             = "inode-1"
}
`
	resource.UnitTest(t, betaTestCase([]resource.TestStep{{
		Config:      config,
		ExpectError: regexp.MustCompile(`already attached|Unable to add Sigma deployment policy document`),
	}}))
}

func TestAccDeploymentPolicyDocumentResource(t *testing.T) {
	runAccDeploymentPolicyDocument(t)
}
