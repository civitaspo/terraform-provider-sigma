package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestUserAttributeUserAssignmentResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	mock.Mux.HandleFunc("/v2/user-attributes/attribute-1/users", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodPost:
			var body map[string]any
			_ = json.NewDecoder(request.Body).Decode(&body)
			assignments := body["assignments"].([]any)
			assignment := assignments[0].(map[string]any)
			if assignment["userId"] != "member-1" {
				t.Errorf("userId = %#v", assignment["userId"])
			}
			value := assignment["value"].(map[string]any)
			if value["val"] != "emea" || value["type"] != "string" {
				t.Errorf("value = %#v", assignment["value"])
			}
			writeJSON(response, map[string]any{})
		case http.MethodGet:
			writeJSON(response, map[string]any{
				"entries": []map[string]any{{
					"userId": "member-1",
					"value":  map[string]string{"val": "emea", "type": "string"},
				}},
				"nextPage": nil,
			})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/user-attributes/attribute-1/users/member-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodDelete {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(response, map[string]any{})
	})

	config := identityProviderConfig(mock) + `
resource "sigma_user_attribute_user_assignment" "test" {
  user_attribute_id = "attribute-1"
  user_id           = "member-1"
  value             = "emea"
}
`
	resource.UnitTest(t, identityTestCase([]resource.TestStep{
		{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_user_attribute_user_assignment.test", "id", "attribute-1/member-1"),
				resource.TestCheckResourceAttr("sigma_user_attribute_user_assignment.test", "value", "emea"),
			),
		},
		{
			ResourceName:      "sigma_user_attribute_user_assignment.test",
			ImportState:       true,
			ImportStateId:     "attribute-1/member-1",
			ImportStateVerify: true,
		},
	}))
}

func TestUserAttributeUserAssignmentResourceRead404RemovesState(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	gone := false
	mock.Mux.HandleFunc("/v2/user-attributes/attribute-1/users", func(response http.ResponseWriter, request *http.Request) {
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
					"userId": "member-1",
					"value":  map[string]string{"val": "emea", "type": "string"},
				}},
				"nextPage": nil,
			})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/user-attributes/attribute-1/users/member-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodDelete {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(response, map[string]any{})
	})
	config := identityProviderConfig(mock) + `
resource "sigma_user_attribute_user_assignment" "test" {
  user_attribute_id = "attribute-1"
  user_id           = "member-1"
  value             = "emea"
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

func TestAccUserAttributeUserAssignmentResource(t *testing.T) { requireAcceptance(t) }
