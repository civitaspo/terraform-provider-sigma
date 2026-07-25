package provider_test

import (
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"testing"

	sigmaprovider "github.com/civitaspo/terraform-provider-sigma/internal/provider"
	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func identityProviderConfig(mock *testutil.MockSigma) string {
	return `
provider "sigma" {
  base_url      = "` + mock.URL() + `"
  client_id     = "` + mock.ClientID + `"
  client_secret = "` + mock.ClientSecret + `"
}
`
}

func identityTestCase(steps []resource.TestStep) resource.TestCase {
	return resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: steps,
	}
}

func writeJSON(response http.ResponseWriter, value any) {
	response.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(response).Encode(value)
}

func TestTeamResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	team := map[string]any{
		"teamId": "team-1", "name": "Analytics", "description": "Core analytics",
		"visibility": "private", "isArchived": false,
	}
	mock.Mux.HandleFunc("/v2/teams", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		var body map[string]any
		_ = json.NewDecoder(request.Body).Decode(&body)
		if body["name"] != "Analytics" || body["description"] != "Core analytics" || body["visibility"] != "private" {
			t.Errorf("create team body = %#v", body)
		}
		writeJSON(response, team)
	})
	mock.Mux.HandleFunc("/v2/teams/team-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodGet:
			writeJSON(response, team)
		case http.MethodPatch:
			var body map[string]any
			_ = json.NewDecoder(request.Body).Decode(&body)
			if body["name"] != "Analytics Updated" {
				t.Errorf("update team name = %#v", body["name"])
			}
			team["name"] = body["name"]
			team["description"] = body["description"]
			writeJSON(response, team)
		case http.MethodDelete:
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})

	createConfig := identityProviderConfig(mock) + `
resource "sigma_team" "test" {
  name        = "Analytics"
  description = "Core analytics"
  visibility  = "private"
}
`
	updateConfig := identityProviderConfig(mock) + `
resource "sigma_team" "test" {
  name        = "Analytics Updated"
  description = "Core analytics"
  visibility  = "private"
}
`
	resource.UnitTest(t, identityTestCase([]resource.TestStep{
		{
			Config: createConfig,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_team.test", "id", "team-1"),
				resource.TestCheckResourceAttr("sigma_team.test", "name", "Analytics"),
				resource.TestCheckResourceAttr("sigma_team.test", "visibility", "private"),
			),
		},
		{
			Config: updateConfig,
			Check:  resource.TestCheckResourceAttr("sigma_team.test", "name", "Analytics Updated"),
		},
		{
			ResourceName:      "sigma_team.test",
			ImportState:       true,
			ImportStateVerify: true,
		},
	}))
}

func TestTeamMemberResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	var mu sync.Mutex
	members := map[string]bool{}
	mock.Mux.HandleFunc("/v2/teams/team-1/members", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		mu.Lock()
		defer mu.Unlock()
		switch request.Method {
		case http.MethodGet:
			entries := []map[string]any{}
			for id := range members {
				entries = append(entries, map[string]any{"userId": id, "isTeamAdmin": false})
			}
			writeJSON(response, map[string]any{"entries": entries, "nextPage": nil})
		case http.MethodPatch:
			var body struct {
				Add    []string `json:"add"`
				Remove []string `json:"remove"`
			}
			_ = json.NewDecoder(request.Body).Decode(&body)
			switch {
			case len(body.Add) == 1 && body.Add[0] == "member-1" && len(body.Remove) == 0:
				// create
			case len(body.Remove) == 1 && body.Remove[0] == "member-1" && len(body.Add) == 0:
				// destroy
			default:
				t.Errorf("team member mutate body = add=%v remove=%v", body.Add, body.Remove)
			}
			for _, id := range body.Add {
				members[id] = true
			}
			for _, id := range body.Remove {
				delete(members, id)
			}
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})

	config := identityProviderConfig(mock) + `
resource "sigma_team_member" "test" {
  team_id   = "team-1"
  member_id = "member-1"
}
`
	resource.UnitTest(t, identityTestCase([]resource.TestStep{
		{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_team_member.test", "id", "team-1/member-1"),
				resource.TestCheckResourceAttr("sigma_team_member.test", "is_team_admin", "false"),
			),
		},
		{
			ResourceName:      "sigma_team_member.test",
			ImportState:       true,
			ImportStateId:     "team-1/member-1",
			ImportStateVerify: true,
		},
	}))
}

func TestTeamMembersResourceAuthoritativeSync(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	var mu sync.Mutex
	members := map[string]bool{"member-keep": true, "member-drop": true}
	var patchBodies []map[string]any

	mock.Mux.HandleFunc("/v2/teams/team-1/members", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		mu.Lock()
		defer mu.Unlock()
		switch request.Method {
		case http.MethodGet:
			entries := []map[string]any{}
			for id := range members {
				entries = append(entries, map[string]any{"userId": id, "isTeamAdmin": false})
			}
			writeJSON(response, map[string]any{"entries": entries, "nextPage": nil})
		case http.MethodPatch:
			var body map[string]any
			_ = json.NewDecoder(request.Body).Decode(&body)
			patchBodies = append(patchBodies, body)
			add, _ := body["add"].([]any)
			remove, _ := body["remove"].([]any)
			for _, raw := range add {
				members[raw.(string)] = true
			}
			for _, raw := range remove {
				delete(members, raw.(string))
			}
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})

	createConfig := identityProviderConfig(mock) + `
resource "sigma_team_members" "test" {
  team_id    = "team-1"
  member_ids = ["member-keep", "member-new"]
}
`
	resource.UnitTest(t, identityTestCase([]resource.TestStep{
		{
			Config: createConfig,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_team_members.test", "id", "team-1"),
				resource.TestCheckResourceAttr("sigma_team_members.test", "member_ids.#", "2"),
				resource.TestCheckTypeSetElemAttr("sigma_team_members.test", "member_ids.*", "member-keep"),
				resource.TestCheckTypeSetElemAttr("sigma_team_members.test", "member_ids.*", "member-new"),
			),
		},
	}))

	mu.Lock()
	defer mu.Unlock()
	if len(patchBodies) < 1 {
		t.Fatal("expected at least one members PATCH for authoritative sync")
	}
	body := patchBodies[0]
	add := stringSlice(body["add"])
	remove := stringSlice(body["remove"])
	sort.Strings(add)
	sort.Strings(remove)
	if len(add) != 1 || add[0] != "member-new" {
		t.Fatalf("sync add = %v, want [member-new]", add)
	}
	if len(remove) != 1 || remove[0] != "member-drop" {
		t.Fatalf("sync remove = %v, want [member-drop]", remove)
	}
}

func stringSlice(raw any) []string {
	values, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, value.(string))
	}
	return out
}

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
		if got := request.URL.Query().Get("reassignToAccountTypeId"); got != "type-default" {
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
		{Config: identityProviderConfig(mock)},
	}))
	if !deleted {
		t.Fatal("expected account type delete")
	}
}

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

func TestUserAttributeTeamAssignmentResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	mock.Mux.HandleFunc("/v2/user-attributes/attribute-1/teams", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodPost:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(map[string]any{
				"entries": []map[string]any{{
					"teamId": "team-1",
					"value":  map[string]string{"val": "americas", "type": "string"},
				}},
				"nextPage": nil,
			})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/user-attributes/attribute-1/teams/team-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodDelete {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{})
	})

	config := identityProviderConfig(mock) + `
resource "sigma_user_attribute_team_assignment" "test" {
  user_attribute_id = "attribute-1"
  team_id            = "team-1"
  value              = "americas"
}
`
	resource.UnitTest(t, identityTestCase([]resource.TestStep{
		{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_user_attribute_team_assignment.test", "id", "attribute-1/team-1"),
				resource.TestCheckResourceAttr("sigma_user_attribute_team_assignment.test", "value", "americas"),
			),
		},
		{
			ResourceName:      "sigma_user_attribute_team_assignment.test",
			ImportState:       true,
			ImportStateId:     "attribute-1/team-1",
			ImportStateVerify: true,
		},
	}))
}
