package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestUserAttributeResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	attribute := map[string]any{
		"userAttributeId": "attribute-1", "name": "Region", "description": "Sales region",
		"defaultValue": map[string]string{"val": "global", "type": "string"},
	}
	mock.Mux.HandleFunc("/v2/user-attributes", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		var body map[string]any
		_ = json.NewDecoder(request.Body).Decode(&body)
		if body["name"] != "Region" || body["description"] != "Sales region" {
			t.Errorf("create user attribute body = %#v", body)
		}
		def := body["defaultValue"].(map[string]any)
		if def["val"] != "global" || def["type"] != "string" {
			t.Errorf("defaultValue = %#v", body["defaultValue"])
		}
		writeJSON(response, attribute)
	})
	mock.Mux.HandleFunc("/v2/user-attributes/attribute-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodGet:
			writeJSON(response, attribute)
		case http.MethodDelete:
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})

	config := identityProviderConfig(mock) + `
resource "sigma_user_attribute" "test" {
  name          = "Region"
  description   = "Sales region"
  default_value = "global"
}
`
	resource.UnitTest(t, identityTestCase([]resource.TestStep{
		{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_user_attribute.test", "id", "attribute-1"),
				resource.TestCheckResourceAttr("sigma_user_attribute.test", "default_value", "global"),
			),
		},
		{
			ResourceName:      "sigma_user_attribute.test",
			ImportState:       true,
			ImportStateVerify: true,
		},
	}))
}

func TestAccUserAttributeResource(t *testing.T) { requireAcceptance(t) }
