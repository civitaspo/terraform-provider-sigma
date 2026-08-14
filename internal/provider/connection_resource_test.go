package provider_test

import (
	"encoding/json"
	"net/http"
	"regexp"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

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
			"description": map[string]any{}, "poolSizes": map[string]any{},
			"timeout":      map[string]any{"default": 30},
			"friendlyName": true, "useOauth": true,
		})
	})
	mock.Mux.HandleFunc("/v2/connections/connection-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(map[string]any{
				"connectionId": "connection-1", "name": "warehouse", "type": "postgres",
				"description": map[string]any{}, "poolSizes": map[string]any{},
				"timeout":      map[string]any{"default": 30},
				"friendlyName": true, "useOauth": true,
			})
		case http.MethodPut:
			_ = json.NewEncoder(response).Encode(map[string]any{
				"connectionId": "connection-1", "name": "warehouse", "type": "postgres",
				"description": map[string]any{}, "poolSizes": map[string]any{},
				"timeout":      map[string]any{"default": 30},
				"friendlyName": true, "useOauth": true,
			})
		case http.MethodPatch:
			t.Errorf("unexpected PATCH; connection updates must use PUT")
			http.Error(response, "PATCH is deprecated", http.StatusMethodNotAllowed)
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
				resource.TestCheckResourceAttr("sigma_connection.test", "use_oauth", "true"),
				resource.TestCheckResourceAttr("sigma_connection.test", "timeout_secs", "30"),
				resource.TestCheckResourceAttr("sigma_connection.test", "timeout.default", "30"),
				resource.TestCheckNoResourceAttr("sigma_connection.test", "credentials_wo"),
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
		"friendlyName": true, "useOauth": false,
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
			if details["password"] != "rotated-secret" || details["host"] != "db.example.com" || details["type"] != "postgres" {
				t.Errorf("merged PUT details = %#v", details)
			}
			if _, ok := body["restore"]; ok {
				t.Errorf("restore unexpectedly present: %#v", body)
			}
			connection["timeoutSecs"] = float64(60)
			_ = json.NewEncoder(response).Encode(connection)
		case http.MethodPatch:
			t.Errorf("unexpected PATCH; connection updates must use PUT")
			http.Error(response, "PATCH is deprecated", http.StatusMethodNotAllowed)
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

func TestConnectionResourceRejectsCredentialsWithoutDetails(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	config := connectionProviderConfig(mock) + `
resource "sigma_connection" "test" {
  name                   = "warehouse"
  credentials_wo         = jsonencode({ password = "secret" })
  credentials_wo_version = 1
}
`
	resource.UnitTest(t, connectionTestCase([]resource.TestStep{{
		Config:      config,
		ExpectError: regexp.MustCompile(`details_json`),
	}}))
}

func TestConnectionResourceVersionRemovalFails(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	connection := map[string]any{
		"connectionId": "connection-1", "name": "warehouse", "type": "postgres",
		"description": map[string]any{}, "poolSizes": map[string]any{},
		"timeout": map[string]any{"default": 30}, "friendlyName": true, "useOauth": false,
	}
	mock.Mux.HandleFunc("/v2/connections", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(connection)
	})
	mock.Mux.HandleFunc("/v2/connections/connection-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
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
}
`
	removeConfig := connectionProviderConfig(mock) + `
resource "sigma_connection" "test" {
  name = "warehouse"
  details_json = jsonencode({
    type     = "postgres"
    host     = "db.example.com"
    database = "analytics"
    user     = "sigma"
  })
}
`
	resource.UnitTest(t, connectionTestCase([]resource.TestStep{
		{Config: createConfig},
		{
			Config:      removeConfig,
			ExpectError: regexp.MustCompile(`Cannot remove credentials_wo_version|must be supplied together`),
		},
	}))
}

func TestConnectionResourceBumpedVersionWithoutPayloadFails(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	connection := map[string]any{
		"connectionId": "connection-1", "name": "warehouse", "type": "postgres",
		"description": map[string]any{}, "poolSizes": map[string]any{},
		"timeout": map[string]any{"default": 30}, "friendlyName": true, "useOauth": false,
	}
	mock.Mux.HandleFunc("/v2/connections", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(connection)
	})
	mock.Mux.HandleFunc("/v2/connections/connection-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
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
}
`
	bumpConfig := connectionProviderConfig(mock) + `
resource "sigma_connection" "test" {
  name = "warehouse"
  details_json = jsonencode({
    type     = "postgres"
    host     = "db.example.com"
    database = "analytics"
    user     = "sigma"
  })
  credentials_wo_version = 2
}
`
	resource.UnitTest(t, connectionTestCase([]resource.TestStep{
		{Config: createConfig},
		{
			Config:      bumpConfig,
			ExpectError: regexp.MustCompile(`must be supplied together|known.*credentials_wo`),
		},
	}))
}

func TestConnectionResourceImportIncompleteUpdateFails(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	connection := map[string]any{
		"connectionId": "connection-1", "name": "warehouse", "type": "postgres",
		"description": map[string]any{}, "poolSizes": map[string]any{},
		"timeout": map[string]any{"default": 30}, "friendlyName": true, "useOauth": false,
	}
	mock.Mux.HandleFunc("/v2/connections/connection-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(connection)
		case http.MethodPut:
			t.Errorf("incomplete imported update must not PUT")
			http.Error(response, "should not PUT", http.StatusInternalServerError)
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	fullConfig := connectionProviderConfig(mock) + `
resource "sigma_connection" "test" {
  name = "warehouse"
  details_json = jsonencode({
    type     = "postgres"
    host     = "db.example.com"
    database = "analytics"
    user     = "sigma"
  })
}
`
	incompleteConfig := connectionProviderConfig(mock) + `
resource "sigma_connection" "test" {
  name = "warehouse"
}
`
	resource.UnitTest(t, connectionTestCase([]resource.TestStep{
		{
			Config:        fullConfig,
			ResourceName:  "sigma_connection.test",
			ImportState:   true,
			ImportStateId: "connection-1",
			Check:         resource.TestCheckResourceAttr("sigma_connection.test", "id", "connection-1"),
		},
		{
			Config:      incompleteConfig,
			ExpectError: regexp.MustCompile(`details_json`),
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
			"friendlyName": true, "useOauth": false,
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
				"friendlyName": true, "useOauth": false,
			})
		case http.MethodPut:
			putCalls++
			http.Error(response, "should not PUT without credentials", http.StatusInternalServerError)
		case http.MethodPatch:
			t.Errorf("unexpected PATCH; connection updates must use PUT")
			http.Error(response, "PATCH is deprecated", http.StatusMethodNotAllowed)
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
			ExpectError: regexp.MustCompile(`Cannot update without resending credentials_wo`),
		},
	}))
	if putCalls != 0 {
		t.Fatalf("unexpected PUT calls without credentials: %d", putCalls)
	}
}

func TestAccConnectionResource(t *testing.T) { runAccConnection(t) }
