package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sync/atomic"
	"testing"

	sigmaprovider "github.com/civitaspo/terraform-provider-sigma/internal/provider"
	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func memberFixture() map[string]any {
	return map[string]any{
		"memberId": "member-1", "organizationId": "org-1", "email": "ada@example.com",
		"firstName": "Ada", "lastName": "Lovelace", "memberType": "Creator", "userKind": "internal",
		"isArchived": false, "isInactive": false, "homeFolderId": "folder-1",
	}
}

func TestMemberResourceCreateErrorDoesNotListOrPatch(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	var listCalls, patchCalls atomic.Int64
	mock.Mux.HandleFunc("/v2/members", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodPost:
			http.Error(response, `{"message":"email already exists"}`, http.StatusConflict)
		case http.MethodGet:
			listCalls.Add(1)
			writeJSON(response, map[string]any{"entries": []any{}, "nextPage": nil})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/members/", func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPatch {
			patchCalls.Add(1)
		}
		http.Error(response, "unexpected member-id request", http.StatusInternalServerError)
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{{
			Config: memberProviderConfig(mock) + `
resource "sigma_member" "test" {
  email      = "ada@example.com"
  first_name = "Ada"
  last_name  = "Lovelace"
}
`,
			ExpectError: regexp.MustCompile(`email already exists`),
		}},
	})
	if listCalls.Load() != 0 || patchCalls.Load() != 0 {
		t.Fatalf("list=%d patch=%d, want 0", listCalls.Load(), patchCalls.Load())
	}
}

func TestMemberResourceImportArchivedThenUnarchive(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	member := memberFixture()
	member["isArchived"] = true
	var patchBodies []map[string]any
	mock.Mux.HandleFunc("/v2/members", func(response http.ResponseWriter, request *http.Request) {
		http.Error(response, "create should not run for import", http.StatusInternalServerError)
	})
	mock.Mux.HandleFunc("/v2/members/member-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodGet:
			writeJSON(response, member)
		case http.MethodPatch:
			var payload map[string]any
			_ = json.NewDecoder(request.Body).Decode(&payload)
			patchBodies = append(patchBodies, payload)
			if archived, ok := payload["isArchived"].(bool); ok {
				member["isArchived"] = archived
			}
			writeJSON(response, member)
		case http.MethodDelete:
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})

	config := memberProviderConfig(mock) + `
resource "sigma_member" "test" {
  email       = "ada@example.com"
  first_name  = "Ada"
  last_name   = "Lovelace"
  is_archived = false
}
`
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config:             config,
				ResourceName:       "sigma_member.test",
				ImportState:        true,
				ImportStateId:      "member-1",
				ImportStatePersist: true,
			},
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sigma_member.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sigma_member.test", "id", "member-1"),
					resource.TestCheckResourceAttr("sigma_member.test", "is_archived", "false"),
				),
			},
		},
	})
	if len(patchBodies) != 1 {
		t.Fatalf("PATCH calls = %d, want 1; bodies=%#v", len(patchBodies), patchBodies)
	}
	if patchBodies[0]["isArchived"] != false {
		t.Fatalf("isArchived PATCH = %#v, want false", patchBodies[0]["isArchived"])
	}
}

func TestMemberResourceRejectsArchivedCreate(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	var posts atomic.Int64
	mock.Mux.HandleFunc("/v2/members", func(response http.ResponseWriter, request *http.Request) {
		posts.Add(1)
		http.Error(response, "create should not run", http.StatusInternalServerError)
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{{
			Config: memberProviderConfig(mock) + `
resource "sigma_member" "test" {
  email       = "ada@example.com"
  first_name  = "Ada"
  last_name   = "Lovelace"
  is_archived = true
}
`,
			ExpectError: regexp.MustCompile(`Cannot create an archived Sigma member`),
		}},
	})
	if posts.Load() != 0 {
		t.Fatalf("POST calls = %d, want 0", posts.Load())
	}
}

func TestMemberResourceRejectsSendInviteChangeWithoutDelete(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	member := memberFixture()
	var allowDelete atomic.Bool
	var deletes atomic.Int64
	mock.Mux.HandleFunc("/v2/members", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(response, member)
	})
	mock.Mux.HandleFunc("/v2/members/member-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodGet:
			writeJSON(response, member)
		case http.MethodDelete:
			deletes.Add(1)
			if !allowDelete.Load() {
				t.Errorf("DELETE during send_invite change")
			}
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: memberProviderConfig(mock) + `
resource "sigma_member" "test" {
  email       = "ada@example.com"
  first_name  = "Ada"
  last_name   = "Lovelace"
  send_invite = false
}
`,
			},
			{
				Config: memberProviderConfig(mock) + `
resource "sigma_member" "test" {
  email       = "ada@example.com"
  first_name  = "Ada"
  last_name   = "Lovelace"
  send_invite = true
}
`,
				ExpectError: regexp.MustCompile(`Cannot change send_invite`),
			},
			{
				PreConfig: func() { allowDelete.Store(true) },
				Config:    memberProviderConfig(mock),
			},
		},
	})
	if deletes.Load() == 0 {
		t.Fatal("expected cleanup DELETE after the failed send_invite plan")
	}
}

func TestMemberResourceRejectsAddToTeams(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{{
			Config: memberProviderConfig(mock) + `
resource "sigma_member" "test" {
  email      = "ada@example.com"
  first_name = "Ada"
  last_name  = "Lovelace"
  add_to_teams = [
    {
      team_id = "team-1"
    }
  ]
}
`,
			ExpectError: regexp.MustCompile(`Unsupported argument`),
		}},
	})
}

func TestMemberResourceUpdatePatchOmitsUnchangedFields(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	member := memberFixture()
	var patchBodies []map[string]any
	mock.Mux.HandleFunc("/v2/members", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(response, member)
	})
	mock.Mux.HandleFunc("/v2/members/member-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodGet:
			writeJSON(response, member)
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
			writeJSON(response, member)
		case http.MethodDelete:
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{
			{Config: memberProviderConfig(mock) + `
resource "sigma_member" "test" {
  email      = "ada@example.com"
  first_name = "Ada"
  last_name  = "Lovelace"
}
`},
			{
				Config: memberProviderConfig(mock) + `
resource "sigma_member" "test" {
  email       = "ada@example.com"
  first_name  = "Augusta"
  last_name   = "Lovelace"
  member_type = "Explorer"
}
`,
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
	if payload["firstName"] != "Augusta" {
		t.Fatalf("firstName = %#v", payload["firstName"])
	}
	if payload["memberType"] != "Explorer" {
		t.Fatalf("memberType = %#v", payload["memberType"])
	}
	for _, key := range []string{"lastName", "email", "userKind", "isArchived"} {
		if _, ok := payload[key]; ok {
			t.Fatalf("%s present in PATCH body, want omitted: %#v", key, payload)
		}
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

func TestMemberResourceCreateSendInvite(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	member := memberFixture()
	var createQuery string
	mock.Mux.HandleFunc("/v2/members", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		createQuery = request.URL.Query().Get("sendInvite")
		writeJSON(response, member)
	})
	mock.Mux.HandleFunc("/v2/members/member-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodGet:
			writeJSON(response, member)
		case http.MethodDelete:
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: memberProviderConfig(mock) + `
resource "sigma_member" "test" {
  email       = "ada@example.com"
  first_name  = "Ada"
  last_name   = "Lovelace"
  send_invite = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sigma_member.test", "id", "member-1"),
					resource.TestCheckResourceAttr("sigma_member.test", "home_folder_id", "folder-1"),
					resource.TestCheckResourceAttr("sigma_member.test", "send_invite", "false"),
				),
			},
			{
				ResourceName:            "sigma_member.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"send_invite", "new_owner_id", "archive_documents", "archive_scheduled_exports"},
			},
		},
	})
	if createQuery != "false" {
		t.Fatalf("sendInvite query = %q, want false", createQuery)
	}
}

func TestMemberResourceDestroyPatchesOwner(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	member := memberFixture()
	var patchBodies []map[string]any
	var deleteCalls atomic.Int64
	mock.Mux.HandleFunc("/v2/members", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(response, member)
	})
	mock.Mux.HandleFunc("/v2/members/member-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodGet:
			writeJSON(response, member)
		case http.MethodPatch:
			var payload map[string]any
			_ = json.NewDecoder(request.Body).Decode(&payload)
			patchBodies = append(patchBodies, payload)
			writeJSON(response, member)
		case http.MethodDelete:
			deleteCalls.Add(1)
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: memberProviderConfig(mock) + `
resource "sigma_member" "test" {
  email                     = "ada@example.com"
  first_name                = "Ada"
  last_name                 = "Lovelace"
  new_owner_id              = "member-admin"
  archive_scheduled_exports = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sigma_member.test", "new_owner_id", "member-admin"),
					resource.TestCheckResourceAttr("sigma_member.test", "archive_scheduled_exports", "true"),
				),
			},
			{Config: memberProviderConfig(mock)},
		},
	})

	if deleteCalls.Load() != 1 {
		t.Fatalf("DELETE calls = %d, want 1", deleteCalls.Load())
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
	if payload["archiveScheduledExports"] != true {
		t.Errorf("archiveScheduledExports = %#v", payload["archiveScheduledExports"])
	}
	if _, ok := payload["archiveDocuments"]; ok {
		t.Errorf("archiveDocuments unexpectedly present: %#v", payload)
	}
}

func TestMemberResourceDestroyPolicyValidation(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	var posts atomic.Int64
	mock.Mux.HandleFunc("/v2/members", func(response http.ResponseWriter, request *http.Request) {
		posts.Add(1)
		http.Error(response, "create should not run", http.StatusInternalServerError)
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{{
			Config: memberProviderConfig(mock) + `
resource "sigma_member" "test" {
  email             = "ada@example.com"
  first_name        = "Ada"
  last_name         = "Lovelace"
  new_owner_id      = "member-admin"
  archive_documents = true
}
`,
			ExpectError: regexp.MustCompile(`Invalid member destroy policy`),
		}},
	})
	if posts.Load() != 0 {
		t.Fatalf("POST calls = %d, want 0", posts.Load())
	}
}

func TestMemberResourcePreservesDestroyPolicyAcrossRead(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	member := memberFixture()
	mock.Mux.HandleFunc("/v2/members", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(response, member)
	})
	mock.Mux.HandleFunc("/v2/members/member-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodGet:
			writeJSON(response, member)
		case http.MethodPatch:
			writeJSON(response, member)
		case http.MethodDelete:
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	config := memberProviderConfig(mock) + `
resource "sigma_member" "test" {
  email             = "ada@example.com"
  first_name        = "Ada"
  last_name         = "Lovelace"
  archive_documents = true
}
`
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.TestCheckResourceAttr("sigma_member.test", "archive_documents", "true"),
			},
			{
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestMemberResourceOmitsCreateOptionsWhenUnset(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	member := memberFixture()
	var createQuery string
	var createBody map[string]any
	var patchCalls atomic.Int64
	mock.Mux.HandleFunc("/v2/members", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		createQuery = request.URL.RawQuery
		_ = json.NewDecoder(request.Body).Decode(&createBody)
		writeJSON(response, member)
	})
	mock.Mux.HandleFunc("/v2/members/member-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodGet:
			writeJSON(response, member)
		case http.MethodPatch:
			patchCalls.Add(1)
			writeJSON(response, member)
		case http.MethodDelete:
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: memberProviderConfig(mock) + `
resource "sigma_member" "test" {
  email      = "ada@example.com"
  first_name = "Ada"
  last_name  = "Lovelace"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sigma_member.test", "home_folder_id", "folder-1"),
					resource.TestCheckNoResourceAttr("sigma_member.test", "send_invite"),
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
	if patchCalls.Load() != 0 {
		t.Fatalf("PATCH calls = %d, want 0 on destroy without options", patchCalls.Load())
	}
}

func TestMemberResourceRead404RemovesState(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	member := memberFixture()
	var gone atomic.Bool
	mock.Mux.HandleFunc("/v2/members", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(response, member)
	})
	mock.Mux.HandleFunc("/v2/members/member-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodGet:
			if gone.Load() {
				writeNotFound(response)
				return
			}
			writeJSON(response, member)
		case http.MethodDelete:
			writeJSON(response, map[string]any{})
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
			{Config: config},
			{
				PreConfig:          func() { gone.Store(true) },
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
				Check: func(state *terraform.State) error {
					if _, ok := state.RootModule().Resources["sigma_member.test"]; ok {
						return fmt.Errorf("sigma_member.test still in state after 404")
					}
					return nil
				},
			},
		},
	})
}

func TestAccMemberResource(t *testing.T) {
	requireAcceptance(t)
	t.Skip("member create/archive is skipped in production to avoid leftover archived users")
}
