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
