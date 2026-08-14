package provider_test

import (
	"encoding/json"
	"net/http"
	"regexp"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

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

func TestConnectionGrantResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	mock.Mux.HandleFunc("/v2/connections/connection-1/grants", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodPost:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(map[string]any{
				"entries":  []map[string]any{{"grantId": "grant-1", "inodeId": "connection-1", "memberId": "member-1", "teamId": nil, "permission": "usage"}},
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
	address := "sigma_connection_grant.test"
	config := connectionProviderConfig(mock) + `
resource "sigma_connection_grant" "test" {
  connection_id = "connection-1"
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
}

func TestAccConnectionGrantResource(t *testing.T) { requireAcceptance(t) }

func TestConnectionGrantResourceInvalidImportID(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	mock.Mux.HandleFunc("/v2/connections/connection-1/grants", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodPost:
			writeJSON(response, map[string]any{})
		case http.MethodGet:
			writeJSON(response, map[string]any{
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
		writeJSON(response, map[string]any{})
	})
	config := connectionProviderConfig(mock) + `
resource "sigma_connection_grant" "test" {
  connection_id = "connection-1"
  member_id     = "member-1"
  permission    = "usage"
}
`
	resource.UnitTest(t, connectionTestCase([]resource.TestStep{
		{Config: config},
		{
			ResourceName:  "sigma_connection_grant.test",
			ImportState:   true,
			ImportStateId: "only-one-segment",
			ExpectError:   regexp.MustCompile(`Invalid import ID`),
		},
	}))
}

func TestConnectionGrantResourceAmbiguousMatches(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	mock.Mux.HandleFunc("/v2/connections/connection-1/grants", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodPost:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(map[string]any{
				"entries": []map[string]any{
					{"grantId": "grant-1", "inodeId": "connection-1", "memberId": "member-1", "teamId": nil, "permission": "usage"},
					{"grantId": "grant-2", "inodeId": "connection-1", "memberId": "member-1", "teamId": nil, "permission": "usage"},
				},
				"nextPage": nil,
			})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	config := connectionProviderConfig(mock) + `
resource "sigma_connection_grant" "test" {
  connection_id = "connection-1"
  member_id     = "member-1"
  permission    = "usage"
}
`
	resource.UnitTest(t, connectionTestCase([]resource.TestStep{{
		Config:      config,
		ExpectError: regexp.MustCompile(`multiple grants matched`),
	}}))
}
