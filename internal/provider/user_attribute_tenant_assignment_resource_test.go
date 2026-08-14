package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestUserAttributeTenantAssignmentResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	currentValue := "tenant-conn"
	mock.Mux.HandleFunc("/v2/user-attributes/attribute-1/tenants", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodPost:
			var body map[string]any
			_ = json.NewDecoder(request.Body).Decode(&body)
			assignments := body["assignments"].([]any)
			assignment := assignments[0].(map[string]any)
			if assignment["tenantOrganizationId"] != "tenant-1" {
				t.Errorf("tenantOrganizationId = %#v", assignment["tenantOrganizationId"])
			}
			value := assignment["value"].(map[string]any)
			if value["val"] != currentValue || value["type"] != "string" {
				t.Errorf("value = %#v", assignment["value"])
			}
			writeJSON(response, map[string]any{})
		case http.MethodGet:
			writeJSON(response, map[string]any{
				"entries": []map[string]any{{
					"tenantOrganizationId": "tenant-1",
					"value":                map[string]string{"val": currentValue, "type": "string"},
				}},
				"nextPage": nil,
			})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/user-attributes/attribute-1/tenants/tenant-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodDelete {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(response, map[string]any{})
	})

	config := identityProviderConfig(mock) + `
resource "sigma_user_attribute_tenant_assignment" "test" {
  user_attribute_id = "attribute-1"
  tenant_id         = "tenant-1"
  value             = "tenant-conn"
}
`
	resource.UnitTest(t, identityTestCase([]resource.TestStep{
		{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_user_attribute_tenant_assignment.test", "id", "attribute-1/tenant-1"),
				resource.TestCheckResourceAttr("sigma_user_attribute_tenant_assignment.test", "tenant_id", "tenant-1"),
				resource.TestCheckResourceAttr("sigma_user_attribute_tenant_assignment.test", "value", "tenant-conn"),
			),
		},
		{
			ResourceName:      "sigma_user_attribute_tenant_assignment.test",
			ImportState:       true,
			ImportStateId:     "attribute-1/tenant-1",
			ImportStateVerify: true,
		},
		{
			PreConfig: func() { currentValue = "tenant-other" },
			Config: identityProviderConfig(mock) + `
resource "sigma_user_attribute_tenant_assignment" "test" {
  user_attribute_id = "attribute-1"
  tenant_id         = "tenant-1"
  value             = "tenant-other"
}
`,
			Check: resource.TestCheckResourceAttr("sigma_user_attribute_tenant_assignment.test", "value", "tenant-other"),
		},
	}))
}

func TestUserAttributeTenantAssignmentResourceRead404RemovesState(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	gone := false
	mock.Mux.HandleFunc("/v2/user-attributes/attribute-1/tenants", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodPost:
			writeJSON(response, map[string]any{})
		case http.MethodGet:
			if gone {
				writeNotFound(response)
				return
			}
			writeJSON(response, map[string]any{
				"entries": []map[string]any{{
					"tenantOrganizationId": "tenant-1",
					"value":                map[string]string{"val": "tenant-conn", "type": "string"},
				}},
				"nextPage": nil,
			})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/user-attributes/attribute-1/tenants/tenant-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodDelete {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(response, map[string]any{})
	})
	config := identityProviderConfig(mock) + `
resource "sigma_user_attribute_tenant_assignment" "test" {
  user_attribute_id = "attribute-1"
  tenant_id         = "tenant-1"
  value             = "tenant-conn"
}
`
	resource.UnitTest(t, identityTestCase([]resource.TestStep{
		{Config: config},
		{
			PreConfig:          func() { gone = true },
			RefreshState:       true,
			ExpectNonEmptyPlan: true,
		},
	}))
}

func TestAccUserAttributeTenantAssignmentResource(t *testing.T) { requireAcceptance(t) }
