package provider_test

import (
	"encoding/json"
	"net/http"
	"regexp"
	"testing"

	sigmaprovider "github.com/civitaspo/terraform-provider-sigma/internal/provider"
	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func connectionProviderConfig(mock *testutil.MockSigma) string {
	return `
provider "sigma" {
  base_url     = "` + mock.URL() + `"
  client_id     = "` + mock.ClientID + `"
  client_secret = "` + mock.ClientSecret + `"
}
`
}

func connectionTestCase(steps []resource.TestStep) resource.TestCase {
	return resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: steps,
	}
}

func TestConnectionResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	mock.Mux.HandleFunc("/v2/connections", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		var body map[string]any
		_ = json.NewDecoder(request.Body).Decode(&body)
		details := body["details"].(map[string]any)
		if details["password"] != "secret" {
			t.Errorf("write-only password was not merged into details: %#v", details)
		}
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{
			"connectionId": "connection-1", "name": "warehouse", "type": "postgres",
			"description": map[string]any{}, "poolSizes": map[string]any{}, "timeoutSecs": 30,
			"friendlyName": true,
		})
	})
	mock.Mux.HandleFunc("/v2/connections/connection-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(map[string]any{
				"connectionId": "connection-1", "name": "warehouse", "type": "postgres",
				"description": map[string]any{}, "poolSizes": map[string]any{}, "timeoutSecs": 30,
				"friendlyName": true,
			})
		case http.MethodPut:
			_ = json.NewEncoder(response).Encode(map[string]any{
				"connectionId": "connection-1", "name": "warehouse", "type": "postgres",
				"description": map[string]any{}, "poolSizes": map[string]any{}, "timeoutSecs": 30,
				"friendlyName": true,
			})
		case http.MethodDelete:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/connections/connection-1/test", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]string{"read": "SUCCESS", "write": "FAILED"})
	})

	config := connectionProviderConfig(mock) + `
resource "sigma_connection" "test" {
  name = "warehouse"
  details_json = jsonencode({
    type     = "postgres"
    host     = "db.example.com"
    database = "analytics"
    user     = "sigma"
  })
  credentials_wo         = jsonencode({ password = "secret" })
  credentials_wo_version = 1
  timeout_secs            = 30
  use_friendly_names      = true
}
`
	resource.UnitTest(t, connectionTestCase([]resource.TestStep{
		{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_connection.test", "id", "connection-1"),
				resource.TestCheckResourceAttr("sigma_connection.test", "type", "postgres"),
			),
		},
		{
			Config: config + "\n",
			Check:  resource.TestCheckResourceAttr("sigma_connection.test", "id", "connection-1"),
		},
	}))
}

func TestConnectionResourceUpdateWithCredentialsVersionBump(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	connection := map[string]any{
		"connectionId": "connection-1", "name": "warehouse", "type": "postgres",
		"description": map[string]any{}, "poolSizes": map[string]any{}, "timeoutSecs": float64(30),
		"friendlyName": true,
	}
	var putBodies []map[string]any
	mock.Mux.HandleFunc("/v2/connections", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(connection)
	})
	mock.Mux.HandleFunc("/v2/connections/connection-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(connection)
		case http.MethodPut:
			var body map[string]any
			_ = json.NewDecoder(request.Body).Decode(&body)
			putBodies = append(putBodies, body)
			if body["name"] != "warehouse" {
				t.Errorf("name = %#v", body["name"])
			}
			if body["timeoutSecs"] != float64(60) {
				t.Errorf("timeoutSecs = %#v", body["timeoutSecs"])
			}
			details, _ := body["details"].(map[string]any)
			if details["password"] != "rotated-secret" {
				t.Errorf("credentials missing from PUT details: %#v", details)
			}
			connection["timeoutSecs"] = float64(60)
			_ = json.NewEncoder(response).Encode(connection)
		case http.MethodDelete:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/connections/connection-1/test", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]string{"read": "SUCCESS", "write": "SUCCESS"})
	})

	createConfig := connectionProviderConfig(mock) + `
resource "sigma_connection" "test" {
  name = "warehouse"
  details_json = jsonencode({
    type     = "postgres"
    host     = "db.example.com"
    database = "analytics"
    user     = "sigma"
  })
  credentials_wo         = jsonencode({ password = "secret" })
  credentials_wo_version = 1
  timeout_secs           = 30
  use_friendly_names     = true
}
`
	updateConfig := connectionProviderConfig(mock) + `
resource "sigma_connection" "test" {
  name = "warehouse"
  details_json = jsonencode({
    type     = "postgres"
    host     = "db.example.com"
    database = "analytics"
    user     = "sigma"
  })
  credentials_wo         = jsonencode({ password = "rotated-secret" })
  credentials_wo_version = 2
  timeout_secs           = 60
  use_friendly_names     = true
}
`
	resource.UnitTest(t, connectionTestCase([]resource.TestStep{
		{Config: createConfig},
		{
			Config: updateConfig,
			Check:  resource.TestCheckResourceAttr("sigma_connection.test", "timeout_secs", "60"),
		},
	}))
	if len(putBodies) != 1 {
		t.Fatalf("expected 1 PUT with credentials version bump, got %d", len(putBodies))
	}
}

func TestConnectionGrantResourceImportCompositeID(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	mock.Mux.HandleFunc("/v2/connections/connection-1/grants", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodPost:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(map[string]any{
				"entries": []map[string]any{{
					"grantId": "grant-1", "inodeId": "connection-1", "memberId": "member-1",
					"teamId": nil, "permission": "usage",
				}},
				"nextPage": nil,
			})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/connections/connection-1/grants/grant-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodDelete {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{})
	})

	config := connectionProviderConfig(mock) + `
resource "sigma_connection_grant" "test" {
  connection_id = "connection-1"
  member_id     = "member-1"
  permission    = "usage"
}
`
	resource.UnitTest(t, connectionTestCase([]resource.TestStep{
		{
			Config: config,
			Check:  resource.TestCheckResourceAttr("sigma_connection_grant.test", "id", "grant-1"),
		},
		{
			ResourceName:      "sigma_connection_grant.test",
			ImportState:       true,
			ImportStateId:     "connection-1/grant-1",
			ImportStateVerify: true,
		},
	}))
}

func TestConnectionResourceUpdateWithoutCredentialsVersionFails(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	putCalls := 0
	mock.Mux.HandleFunc("/v2/connections", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{
			"connectionId": "connection-1", "name": "warehouse", "type": "postgres",
			"description": map[string]any{}, "poolSizes": map[string]any{}, "timeoutSecs": 30,
			"friendlyName": true,
		})
	})
	mock.Mux.HandleFunc("/v2/connections/connection-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(map[string]any{
				"connectionId": "connection-1", "name": "warehouse", "type": "postgres",
				"description": map[string]any{}, "poolSizes": map[string]any{}, "timeoutSecs": 30,
				"friendlyName": true,
			})
		case http.MethodPut:
			putCalls++
			http.Error(response, "should not PUT without credentials", http.StatusInternalServerError)
		case http.MethodDelete:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/connections/connection-1/test", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]string{"read": "SUCCESS", "write": "SUCCESS"})
	})

	createConfig := connectionProviderConfig(mock) + `
resource "sigma_connection" "test" {
  name = "warehouse"
  details_json = jsonencode({
    type     = "postgres"
    host     = "db.example.com"
    database = "analytics"
    user     = "sigma"
  })
  credentials_wo         = jsonencode({ password = "secret" })
  credentials_wo_version = 1
  timeout_secs           = 30
  use_friendly_names     = true
}
`
	updateConfig := connectionProviderConfig(mock) + `
resource "sigma_connection" "test" {
  name = "warehouse"
  details_json = jsonencode({
    type     = "postgres"
    host     = "db.example.com"
    database = "analytics"
    user     = "sigma"
  })
  credentials_wo         = jsonencode({ password = "secret" })
  credentials_wo_version = 1
  timeout_secs           = 60
  use_friendly_names     = true
}
`
	resource.UnitTest(t, connectionTestCase([]resource.TestStep{
		{Config: createConfig},
		{
			Config:      updateConfig,
			ExpectError: regexp.MustCompile(`Cannot update Sigma connection without resending credentials`),
		},
	}))
	if putCalls != 0 {
		t.Fatalf("unexpected PUT calls without credentials: %d", putCalls)
	}
}

func TestConnectionGrantResources(t *testing.T) {
	tests := []struct {
		name       string
		resource   string
		parentAttr string
		parentID   string
		basePath   string
	}{
		{name: "connection", resource: "sigma_connection_grant", parentAttr: "connection_id", parentID: "connection-1", basePath: "/v2/connections/connection-1/grants"},
		{name: "path", resource: "sigma_connection_path_grant", parentAttr: "connection_path_id", parentID: "path-1", basePath: "/v2/connections/paths/path-1/grants"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mock := testutil.NewMockSigma(t)
			mock.Mux.HandleFunc(test.basePath, func(response http.ResponseWriter, request *http.Request) {
				mock.AssertBearer(t, request)
				response.Header().Set("Content-Type", "application/json")
				switch request.Method {
				case http.MethodPost:
					_ = json.NewEncoder(response).Encode(map[string]any{})
				case http.MethodGet:
					_ = json.NewEncoder(response).Encode(map[string]any{
						"entries":  []map[string]any{{"grantId": "grant-1", "inodeId": test.parentID, "memberId": "member-1", "teamId": nil, "permission": "usage"}},
						"nextPage": nil,
					})
				default:
					http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
				}
			})
			mock.Mux.HandleFunc(test.basePath+"/grant-1", func(response http.ResponseWriter, request *http.Request) {
				mock.AssertBearer(t, request)
				if request.Method != http.MethodDelete {
					http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
					return
				}
				response.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(response).Encode(map[string]any{})
			})
			address := test.resource + ".test"
			config := connectionProviderConfig(mock) + `
resource "` + test.resource + `" "test" {
  ` + test.parentAttr + ` = "` + test.parentID + `"
  member_id = "member-1"
  permission = "usage"
}
`
			resource.UnitTest(t, connectionTestCase([]resource.TestStep{{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(address, "id", "grant-1"),
					resource.TestCheckResourceAttr(address, "permission", "usage"),
				),
			}}))
		})
	}
}

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
	}))
}

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
	}))
}

func TestConnectionDataSources(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	mock.Mux.HandleFunc("/v2/connections/connection-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{"connectionId": "connection-1", "name": "warehouse", "type": "snowflake", "description": map[string]any{}, "poolSizes": map[string]any{}, "friendlyName": false})
	})
	mock.Mux.HandleFunc("/v2/connections", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{"entries": []map[string]any{{"connectionId": "connection-1", "name": "warehouse", "type": "snowflake", "description": map[string]any{}, "poolSizes": map[string]any{}, "friendlyName": false}}, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/connection/connection-1/lookup", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{"kind": "table", "inodeId": "path-1", "url": "/connection/path-1"})
	})
	mock.Mux.HandleFunc("/v2/connections/paths", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{"entries": []map[string]any{{"connectionId": "connection-1", "urlId": "path-1", "path": []string{"DATABASE", "SCHEMA"}}}, "nextPage": nil})
	})
	config := connectionProviderConfig(mock) + `
data "sigma_connection" "one" {
  id = "connection-1"
}
data "sigma_connections" "all" {}
data "sigma_connection_paths" "lookup" {
  connection_id = "connection-1"
  path          = ["DATABASE", "SCHEMA", "TABLE"]
}
data "sigma_connection_paths" "all" {
  connection_id = "connection-1"
}
`
	resource.UnitTest(t, connectionTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.sigma_connection.one", "type", "snowflake"),
			resource.TestCheckResourceAttr("data.sigma_connections.all", "connections.#", "1"),
			resource.TestCheckResourceAttr("data.sigma_connection_paths.lookup", "inode_id", "path-1"),
			resource.TestCheckResourceAttr("data.sigma_connection_paths.all", "paths.#", "1"),
		),
	}}))
}
