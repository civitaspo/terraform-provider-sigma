package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccountTypeResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	accountType := map[string]any{
		"accountTypeId": "type-1", "accountTypeName": "Analyst", "description": "Analyst access", "isCustom": true,
	}
	deleted := false
	mock.Mux.HandleFunc("/v2/accountTypes", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodPost:
			var body map[string]any
			_ = json.NewDecoder(request.Body).Decode(&body)
			if body["name"] != "Analyst" || body["description"] != "Analyst access" {
				t.Errorf("create account type body = %#v", body)
			}
			perms, _ := body["permissions"].([]any)
			if len(perms) != 1 || perms[0] != "viewWorksheet" {
				t.Errorf("permissions = %#v", body["permissions"])
			}
			response.WriteHeader(http.StatusCreated)
			writeJSON(response, accountType)
		case http.MethodGet:
			writeJSON(response, map[string]any{"entries": []any{accountType}})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/accountTypes/type-1/permissions", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		writeJSON(response, []map[string]any{{"permission": "viewWorksheet", "description": "View worksheets"}})
	})
	mock.Mux.HandleFunc("/v2/accountTypes/type-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodDelete {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		if got := request.URL.Query().Get("reassignToAccountTypeId"); got != "type-other" {
			t.Errorf("reassignToAccountTypeId = %q", got)
		}
		deleted = true
		writeJSON(response, map[string]any{})
	})

	config := identityProviderConfig(mock) + `
resource "sigma_account_type" "test" {
  name                         = "Analyst"
  description                  = "Analyst access"
  permissions                  = ["viewWorksheet"]
  reassign_to_account_type_id  = "type-default"
}
`
	resource.UnitTest(t, identityTestCase([]resource.TestStep{
		{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_account_type.test", "id", "type-1"),
				resource.TestCheckResourceAttr("sigma_account_type.test", "is_custom", "true"),
				resource.TestCheckTypeSetElemAttr("sigma_account_type.test", "permissions.*", "viewWorksheet"),
			),
		},
		{
			ResourceName:            "sigma_account_type.test",
			ImportState:             true,
			ImportStateId:           "Analyst",
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: []string{"reassign_to_account_type_id"},
		},
		{
			Config: identityProviderConfig(mock) + `
resource "sigma_account_type" "test" {
  name                         = "Analyst"
  description                  = "Analyst access"
  permissions                  = ["viewWorksheet"]
  reassign_to_account_type_id  = "type-other"
}
`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_account_type.test", "id", "type-1"),
				resource.TestCheckResourceAttr("sigma_account_type.test", "reassign_to_account_type_id", "type-other"),
			),
		},
		{Config: identityProviderConfig(mock)},
	}))
	if !deleted {
		t.Fatal("expected account type delete")
	}
}

func TestAccountTypeResourceRead404RemovesState(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	accountType := map[string]any{
		"accountTypeId": "type-1", "accountTypeName": "Analyst", "description": "Analyst access", "isCustom": true,
	}
	gone := false
	mock.Mux.HandleFunc("/v2/accountTypes", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodPost:
			response.WriteHeader(http.StatusCreated)
			writeJSON(response, accountType)
		case http.MethodGet:
			entries := []any{accountType}
			if gone {
				entries = []any{}
			}
			writeJSON(response, map[string]any{"entries": entries})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/accountTypes/type-1/permissions", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		writeJSON(response, []map[string]any{{"permission": "viewWorksheet", "description": "View worksheets"}})
	})
	mock.Mux.HandleFunc("/v2/accountTypes/type-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodDelete {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(response, map[string]any{})
	})
	config := identityProviderConfig(mock) + `
resource "sigma_account_type" "test" {
  name        = "Analyst"
  description = "Analyst access"
  permissions = ["viewWorksheet"]
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

func TestAccAccountTypeResource(t *testing.T) { runAccAccountType(t) }
