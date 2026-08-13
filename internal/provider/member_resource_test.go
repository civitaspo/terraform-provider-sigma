package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	sigmaprovider "github.com/civitaspo/terraform-provider-sigma/internal/provider"
	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestMemberResourceRecreateReactivatesArchived(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	member := map[string]any{
		"memberId": "member-1", "organizationId": "org-1", "email": "ada@example.com",
		"firstName": "Ada", "lastName": "Lovelace", "memberType": "Creator", "userKind": "internal",
		"isArchived": false, "isInactive": false,
	}
	createCalls := 0
	mock.Mux.HandleFunc("/v2/members", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodPost:
			createCalls++
			if createCalls == 1 {
				_ = json.NewEncoder(response).Encode(member)
				return
			}
			http.Error(response, `{"message":"email already exists"}`, http.StatusConflict)
		case http.MethodGet:
			archived := map[string]any{}
			for key, value := range member {
				archived[key] = value
			}
			archived["isArchived"] = true
			_ = json.NewEncoder(response).Encode(map[string]any{"entries": []any{archived}, "nextPage": nil})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/members/member-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
			current := map[string]any{}
			for key, value := range member {
				current[key] = value
			}
			_ = json.NewEncoder(response).Encode(current)
		case http.MethodPatch:
			var payload map[string]any
			_ = json.NewDecoder(request.Body).Decode(&payload)
			if payload["isArchived"] != false {
				t.Errorf("isArchived = %#v, want false", payload["isArchived"])
			}
			member["isArchived"] = false
			_ = json.NewEncoder(response).Encode(member)
		case http.MethodDelete:
			member["isArchived"] = true
			_ = json.NewEncoder(response).Encode(map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})

	providerConfig := `
provider "sigma" {
  base_url      = "` + mock.URL() + `"
  client_id     = "` + mock.ClientID + `"
  client_secret = "` + mock.ClientSecret + `"
}
`
	config := providerConfig + `
resource "sigma_member" "test" {
  email      = "ada@example.com"
  first_name = "Ada"
  last_name  = "Lovelace"
}
`
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{
			{Config: config},
			{Config: providerConfig},
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sigma_member.test", "id", "member-1"),
					resource.TestCheckResourceAttr("sigma_member.test", "is_archived", "false"),
				),
			},
		},
	})
}

func TestMemberResourceUpdatePatchOmitsNullFields(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	member := map[string]any{
		"memberId": "member-1", "organizationId": "org-1", "email": "ada@example.com",
		"firstName": "Ada", "lastName": "Lovelace", "memberType": "Creator", "userKind": "internal",
		"isArchived": false, "isInactive": false,
	}
	var patchBodies []map[string]any
	mock.Mux.HandleFunc("/v2/members", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(member)
	})
	mock.Mux.HandleFunc("/v2/members/member-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(member)
		case http.MethodPatch:
			var payload map[string]any
			_ = json.NewDecoder(request.Body).Decode(&payload)
			patchBodies = append(patchBodies, payload)
			if first, ok := payload["firstName"].(string); ok {
				member["firstName"] = first
			}
			if memberType, ok := payload["memberType"].(string); ok {
				member["memberType"] = memberType
			}
			_ = json.NewEncoder(response).Encode(member)
		case http.MethodDelete:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})

	providerConfig := `
provider "sigma" {
  base_url      = "` + mock.URL() + `"
  client_id     = "` + mock.ClientID + `"
  client_secret = "` + mock.ClientSecret + `"
}
`
	createConfig := providerConfig + `
resource "sigma_member" "test" {
  email      = "ada@example.com"
  first_name = "Ada"
  last_name  = "Lovelace"
}
`
	updateConfig := providerConfig + `
resource "sigma_member" "test" {
  email       = "ada@example.com"
  first_name  = "Augusta"
  last_name   = "Lovelace"
  member_type = "Explorer"
  user_kind   = null
}
`
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{
			{Config: createConfig},
			{
				Config: updateConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sigma_member.test", "first_name", "Augusta"),
					resource.TestCheckResourceAttr("sigma_member.test", "member_type", "Explorer"),
				),
			},
		},
	})

	if len(patchBodies) != 1 {
		t.Fatalf("expected 1 PATCH, got %d", len(patchBodies))
	}
	payload := patchBodies[0]
	if payload["firstName"] != "Augusta" || payload["lastName"] != "Lovelace" || payload["email"] != "ada@example.com" {
		t.Fatalf("PATCH identity fields = %#v", payload)
	}
	if payload["memberType"] != "Explorer" {
		t.Fatalf("memberType = %#v, want Explorer", payload["memberType"])
	}
	if _, ok := payload["userKind"]; ok {
		t.Fatalf("userKind present in PATCH body, want omitted: %#v", payload)
	}
}

func memberProviderConfig(mock *testutil.MockSigma) string {
	return `
provider "sigma" {
  base_url      = "` + mock.URL() + `"
  client_id     = "` + mock.ClientID + `"
  client_secret = "` + mock.ClientSecret + `"
}
`
}

func TestMemberResourceCreateAddToTeamsAndSendInvite(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	member := map[string]any{
		"memberId": "member-1", "organizationId": "org-1", "email": "ada@example.com",
		"firstName": "Ada", "lastName": "Lovelace", "memberType": "Creator", "userKind": "internal",
		"isArchived": false, "isInactive": false, "homeFolderId": "folder-1",
	}
	var createQuery string
	var createBody map[string]any
	createCalls := 0
	mock.Mux.HandleFunc("/v2/members", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		createCalls++
		createQuery = request.URL.Query().Get("sendInvite")
		_ = json.NewDecoder(request.Body).Decode(&createBody)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(member)
	})
	mock.Mux.HandleFunc("/v2/members/member-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(member)
		case http.MethodPatch:
			var payload map[string]any
			_ = json.NewDecoder(request.Body).Decode(&payload)
			if first, ok := payload["firstName"].(string); ok {
				member["firstName"] = first
			}
			_ = json.NewEncoder(response).Encode(member)
		case http.MethodDelete:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})

	createConfig := memberProviderConfig(mock) + `
resource "sigma_member" "test" {
  email       = "ada@example.com"
  first_name  = "Ada"
  last_name   = "Lovelace"
  send_invite = false
  add_to_teams = [
    {
      team_id       = "team-1"
      is_team_admin = true
    }
  ]
}
`
	renameConfig := memberProviderConfig(mock) + `
resource "sigma_member" "test" {
  email       = "ada@example.com"
  first_name  = "Augusta"
  last_name   = "Lovelace"
  send_invite = false
  add_to_teams = [
    {
      team_id       = "team-1"
      is_team_admin = true
    }
  ]
}
`
	replaceConfig := memberProviderConfig(mock) + `
resource "sigma_member" "test" {
  email       = "ada@example.com"
  first_name  = "Augusta"
  last_name   = "Lovelace"
  send_invite = true
  add_to_teams = [
    {
      team_id       = "team-1"
      is_team_admin = true
    }
  ]
}
`
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: createConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sigma_member.test", "id", "member-1"),
					resource.TestCheckResourceAttr("sigma_member.test", "home_folder_id", "folder-1"),
					resource.TestCheckResourceAttr("sigma_member.test", "send_invite", "false"),
					resource.TestCheckResourceAttr("sigma_member.test", "add_to_teams.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("sigma_member.test", "add_to_teams.*", map[string]string{
						"team_id":       "team-1",
						"is_team_admin": "true",
					}),
				),
			},
			{
				Config: renameConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sigma_member.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sigma_member.test", "first_name", "Augusta"),
					resource.TestCheckResourceAttr("sigma_member.test", "send_invite", "false"),
					resource.TestCheckResourceAttr("sigma_member.test", "home_folder_id", "folder-1"),
				),
			},
			{
				Config: replaceConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sigma_member.test", plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.TestCheckResourceAttr("sigma_member.test", "send_invite", "true"),
			},
			{
				ResourceName:            "sigma_member.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"send_invite", "add_to_teams", "new_owner_id", "archive_documents", "archive_scheduled_exports"},
			},
		},
	})

	if createCalls != 2 {
		t.Fatalf("create calls = %d, want 2", createCalls)
	}
	if createQuery != "true" {
		t.Fatalf("last sendInvite query = %q, want true", createQuery)
	}
	teams, _ := createBody["addToTeams"].([]any)
	if len(teams) != 1 {
		t.Fatalf("last addToTeams = %#v", createBody["addToTeams"])
	}
}

func TestMemberResourceDestroyPatchesOwnerAndArchiveOptions(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	member := map[string]any{
		"memberId": "member-1", "organizationId": "org-1", "email": "ada@example.com",
		"firstName": "Ada", "lastName": "Lovelace", "memberType": "Creator", "userKind": "internal",
		"isArchived": false, "isInactive": false, "homeFolderId": "folder-1",
	}
	var patchBodies []map[string]any
	deleteCalls := 0
	mock.Mux.HandleFunc("/v2/members", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(member)
	})
	mock.Mux.HandleFunc("/v2/members/member-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(member)
		case http.MethodPatch:
			var payload map[string]any
			_ = json.NewDecoder(request.Body).Decode(&payload)
			patchBodies = append(patchBodies, payload)
			_ = json.NewEncoder(response).Encode(member)
		case http.MethodDelete:
			deleteCalls++
			_ = json.NewEncoder(response).Encode(map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})

	config := memberProviderConfig(mock) + `
resource "sigma_member" "test" {
  email                     = "ada@example.com"
  first_name                = "Ada"
  last_name                 = "Lovelace"
  new_owner_id              = "member-admin"
  archive_documents         = true
  archive_scheduled_exports = true
}
`
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sigma_member.test", "new_owner_id", "member-admin"),
					resource.TestCheckResourceAttr("sigma_member.test", "archive_documents", "true"),
					resource.TestCheckResourceAttr("sigma_member.test", "archive_scheduled_exports", "true"),
				),
			},
			{Config: memberProviderConfig(mock)},
		},
	})

	if deleteCalls != 1 {
		t.Fatalf("DELETE calls = %d, want 1", deleteCalls)
	}
	if len(patchBodies) != 1 {
		t.Fatalf("PATCH calls = %d, want 1 (destroy only); bodies=%#v", len(patchBodies), patchBodies)
	}
	payload := patchBodies[0]
	if payload["isArchived"] != true {
		t.Errorf("isArchived = %#v", payload["isArchived"])
	}
	if payload["newOwnerId"] != "member-admin" {
		t.Errorf("newOwnerId = %#v", payload["newOwnerId"])
	}
	if payload["archiveDocuments"] != true {
		t.Errorf("archiveDocuments = %#v", payload["archiveDocuments"])
	}
	if payload["archiveScheduledExports"] != true {
		t.Errorf("archiveScheduledExports = %#v", payload["archiveScheduledExports"])
	}
}

func TestMemberResourceOmitsCreateOptionsWhenUnset(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	member := map[string]any{
		"memberId": "member-1", "organizationId": "org-1", "email": "ada@example.com",
		"firstName": "Ada", "lastName": "Lovelace", "memberType": "Creator", "userKind": "internal",
		"isArchived": false, "isInactive": false, "homeFolderId": "folder-1",
	}
	var createQuery string
	var createBody map[string]any
	patchCalls := 0
	mock.Mux.HandleFunc("/v2/members", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		createQuery = request.URL.RawQuery
		_ = json.NewDecoder(request.Body).Decode(&createBody)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(member)
	})
	mock.Mux.HandleFunc("/v2/members/member-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(member)
		case http.MethodPatch:
			patchCalls++
			_ = json.NewEncoder(response).Encode(member)
		case http.MethodDelete:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})

	config := memberProviderConfig(mock) + `
resource "sigma_member" "test" {
  email      = "ada@example.com"
  first_name = "Ada"
  last_name  = "Lovelace"
}
`
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sigma_member.test", "home_folder_id", "folder-1"),
					resource.TestCheckNoResourceAttr("sigma_member.test", "send_invite"),
					resource.TestCheckNoResourceAttr("sigma_member.test", "add_to_teams"),
				),
			},
			{Config: memberProviderConfig(mock)},
		},
	})

	if createQuery != "" {
		t.Fatalf("create query = %q, want empty", createQuery)
	}
	if _, ok := createBody["addToTeams"]; ok {
		t.Fatalf("addToTeams unexpectedly set: %#v", createBody["addToTeams"])
	}
	if patchCalls != 0 {
		t.Fatalf("PATCH calls = %d, want 0 on destroy without options", patchCalls)
	}
}

func TestAccMemberResource(t *testing.T) { requireAcceptance(t) }
