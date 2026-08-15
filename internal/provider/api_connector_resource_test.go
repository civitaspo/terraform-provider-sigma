package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAPIConnectorResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	connector := map[string]any{
		"apiConnectorId": "connector-1", "name": "weather", "description": "Weather API",
		"params": map[string]any{"method": "GET", "url": "https://api.example.com/weather", "headers": []any{}, "pathParams": []any{}, "queryParams": []any{}},
		"config": map[string]any{}, "authId": "credential-1",
	}
	mock.Mux.HandleFunc("/v2/api-connectors", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(connector)
	})
	mock.Mux.HandleFunc("/v2/api-connectors/connector-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(connector)
		case http.MethodDelete:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	config := connectionProviderConfig(mock) + `
resource "sigma_api_connector" "test" {
  name        = "weather"
  description = "Weather API"
  auth_id     = "credential-1"
  params_json = jsonencode({
    method      = "GET"
    url         = "https://api.example.com/weather"
    headers     = []
    pathParams  = []
    queryParams = []
  })
}
`
	resource.UnitTest(t, connectionTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("sigma_api_connector.test", "id", "connector-1"),
			resource.TestCheckResourceAttr("sigma_api_connector.test", "auth_id", "credential-1"),
		),
	}}))
}

func TestAPIConnectorResourcePreservesConfiguredParamsOnRead(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	created := map[string]any{
		"apiConnectorId": "connector-1", "name": "weather", "description": "",
		"params": map[string]any{"method": "GET", "url": "https://example.com/tf-acc", "headers": []any{}, "pathParams": []any{}, "queryParams": []any{}, "body": ""},
		"config": map[string]any{}, "authId": "credential-1",
	}
	read := map[string]any{
		"apiConnectorId": "connector-1", "name": "weather", "description": "",
		"params": map[string]any{"method": "GET", "url": "https://example.com/tf-acc", "headers": []any{}, "pathParams": []any{}, "queryParams": []any{}, "body": "", "bodyParams": []any{}},
		"config": map[string]any{}, "authId": "credential-1",
	}
	mock.Mux.HandleFunc("/v2/api-connectors", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		writeJSON(response, created)
	})
	mock.Mux.HandleFunc("/v2/api-connectors/connector-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method == http.MethodDelete {
			writeJSON(response, map[string]any{})
			return
		}
		writeJSON(response, read)
	})
	config := connectionProviderConfig(mock) + `
resource "sigma_api_connector" "test" {
  name    = "weather"
  auth_id = "credential-1"
  params_json = jsonencode({
    method      = "GET"
    url         = "https://example.com/tf-acc"
    headers     = []
    pathParams  = []
    queryParams = []
    body        = ""
  })
}
`
	resource.UnitTest(t, connectionTestCase([]resource.TestStep{{
		Config: config,
		Check:  resource.TestCheckResourceAttr("sigma_api_connector.test", "id", "connector-1"),
	}}))
}

func TestAPIConnectorResourceUpdateOmitsParamsWithoutSecretsVersion(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	params := map[string]any{"method": "GET", "url": "https://api.example.com/weather", "headers": []any{}, "pathParams": []any{}, "queryParams": []any{}}
	connector := map[string]any{
		"apiConnectorId": "connector-1", "name": "weather", "description": "Weather API",
		"params": params, "config": map[string]any{}, "authId": "credential-1",
	}
	mock.Mux.HandleFunc("/v2/api-connectors", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(connector)
	})
	mock.Mux.HandleFunc("/v2/api-connectors/connector-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(connector)
		case http.MethodPatch:
			var body map[string]any
			_ = json.NewDecoder(request.Body).Decode(&body)
			if _, ok := body["params"]; ok {
				t.Errorf("metadata update must omit params to retain secrets; got %#v", body)
			}
			if body["description"] != "Updated" {
				t.Errorf("description = %#v", body["description"])
			}
			connector["description"] = "Updated"
			_ = json.NewEncoder(response).Encode(connector)
		case http.MethodDelete:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	createConfig := connectionProviderConfig(mock) + `
resource "sigma_api_connector" "test" {
  name        = "weather"
  description = "Weather API"
  auth_id     = "credential-1"
  params_json = jsonencode({
    method      = "GET"
    url         = "https://api.example.com/weather"
    headers     = []
    pathParams  = []
    queryParams = []
  })
  secrets_wo = jsonencode({
    queryParams = [{ name = "api_key", value = "secret", type = "string" }]
  })
  secrets_wo_version = 1
}
`
	updateConfig := connectionProviderConfig(mock) + `
resource "sigma_api_connector" "test" {
  name        = "weather"
  description = "Updated"
  auth_id     = "credential-1"
  params_json = jsonencode({
    method      = "GET"
    url         = "https://api.example.com/weather"
    headers     = []
    pathParams  = []
    queryParams = []
  })
  secrets_wo = jsonencode({
    queryParams = [{ name = "api_key", value = "secret", type = "string" }]
  })
  secrets_wo_version = 1
}
`
	resource.UnitTest(t, connectionTestCase([]resource.TestStep{
		{Config: createConfig},
		{
			Config: updateConfig,
			Check:  resource.TestCheckResourceAttr("sigma_api_connector.test", "description", "Updated"),
		},
		{
			ResourceName:            "sigma_api_connector.test",
			ImportState:             true,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: []string{"secrets_wo", "secrets_wo_version"},
		},
	}))
}

func TestAPIConnectorResourceRedactedReadKeepsConfiguredParams(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	createParams := map[string]any{"method": "GET", "url": "https://api.example.com/weather", "headers": []any{}, "pathParams": []any{}, "queryParams": []any{}}
	connector := map[string]any{
		"apiConnectorId": "connector-1", "name": "weather", "description": "Weather API",
		"params": createParams, "config": map[string]any{}, "authId": "credential-1",
	}
	mock.Mux.HandleFunc("/v2/api-connectors", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(connector)
	})
	mock.Mux.HandleFunc("/v2/api-connectors/connector-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
			redacted := map[string]any{
				"apiConnectorId": "connector-1", "name": "weather", "description": "Weather API",
				"params": map[string]any{"method": "GET", "url": "https://api.example.com/weather", "queryParams": []any{map[string]any{"name": "api_key", "value": "REDACTED"}}},
				"config": map[string]any{}, "authId": "credential-1",
			}
			_ = json.NewEncoder(response).Encode(redacted)
		case http.MethodDelete:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	config := connectionProviderConfig(mock) + `
resource "sigma_api_connector" "test" {
  name        = "weather"
  description = "Weather API"
  auth_id     = "credential-1"
  params_json = jsonencode({
    method      = "GET"
    url         = "https://api.example.com/weather"
    headers     = []
    pathParams  = []
    queryParams = []
  })
  secrets_wo = jsonencode({
    queryParams = [{ name = "api_key", value = "secret", type = "string" }]
  })
  secrets_wo_version = 1
}
`
	resource.UnitTest(t, connectionTestCase([]resource.TestStep{
		{Config: config},
		{
			Config:   config,
			PlanOnly: true,
		},
	}))
}

func TestAccAPIConnectorResource(t *testing.T) { runAccAPICredentialAndConnector(t) }
