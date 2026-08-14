package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAPICredentialResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	credential := map[string]any{
		"apiCredentialId": "credential-1", "name": "weather", "description": "Weather credential",
		"authMethod": "apiKey", "allowlist": []string{"api.example.com"},
		"credential": map[string]any{"authMethod": "apiKey", "apiKey": map[string]any{"key": "X-API-Key", "isQueryParam": false}},
	}
	mock.Mux.HandleFunc("/v2/api-credentials", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		var body map[string]any
		_ = json.NewDecoder(request.Body).Decode(&body)
		if body["credential"].(map[string]any)["apiKey"].(map[string]any)["value"] != "secret" {
			t.Errorf("credential secret missing from request")
		}
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(credential)
	})
	mock.Mux.HandleFunc("/v2/api-credentials/credential-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(credential)
		case http.MethodPatch:
			var body map[string]any
			_ = json.NewDecoder(request.Body).Decode(&body)
			if _, ok := body["credential"]; ok {
				t.Errorf("metadata update must omit credential; got %#v", body)
			}
			if body["description"] != "Updated credential" {
				t.Errorf("description = %#v", body["description"])
			}
			credential["description"] = "Updated credential"
			_ = json.NewEncoder(response).Encode(credential)
		case http.MethodDelete:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	createConfig := connectionProviderConfig(mock) + `
resource "sigma_api_credential" "test" {
  name        = "weather"
  description = "Weather credential"
  allowlist   = ["api.example.com"]
  credential_wo = jsonencode({
    authMethod = "apiKey"
    apiKey = {
      key          = "X-API-Key"
      value        = "secret"
      isQueryParam = false
    }
  })
  credential_wo_version = 1
}
`
	updateConfig := connectionProviderConfig(mock) + `
resource "sigma_api_credential" "test" {
  name        = "weather"
  description = "Updated credential"
  allowlist   = ["api.example.com"]
  credential_wo = jsonencode({
    authMethod = "apiKey"
    apiKey = {
      key          = "X-API-Key"
      value        = "secret"
      isQueryParam = false
    }
  })
  credential_wo_version = 1
}
`
	resource.UnitTest(t, connectionTestCase([]resource.TestStep{
		{
			Config: createConfig,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_api_credential.test", "id", "credential-1"),
				resource.TestCheckResourceAttr("sigma_api_credential.test", "auth_method", "apiKey"),
			),
		},
		{
			Config: updateConfig,
			Check:  resource.TestCheckResourceAttr("sigma_api_credential.test", "description", "Updated credential"),
		},
		{
			ResourceName:            "sigma_api_credential.test",
			ImportState:             true,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: []string{"credential_wo", "credential_wo_version"},
		},
	}))
}

func TestAPICredentialResourceOmitsEmptyDescription(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	credential := map[string]any{
		"apiCredentialId": "credential-1", "name": "weather", "description": "",
		"authMethod": "bearer", "allowlist": []string{"example.com"},
		"credential": map[string]any{"authMethod": "bearer"},
	}
	mock.Mux.HandleFunc("/v2/api-credentials", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		writeJSON(response, credential)
	})
	mock.Mux.HandleFunc("/v2/api-credentials/credential-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method == http.MethodDelete {
			writeJSON(response, map[string]any{})
			return
		}
		writeJSON(response, credential)
	})
	config := connectionProviderConfig(mock) + `
resource "sigma_api_credential" "test" {
  name      = "weather"
  allowlist = ["example.com"]
  credential_wo = jsonencode({
    authMethod = "bearer"
    bearer     = { token = "secret" }
  })
  credential_wo_version = 1
}
`
	resource.UnitTest(t, connectionTestCase([]resource.TestStep{{
		Config: config,
		Check:  resource.TestCheckNoResourceAttr("sigma_api_credential.test", "description"),
	}}))
}

func TestAccAPICredentialResource(t *testing.T) { runAccAPICredentialAndConnector(t) }
