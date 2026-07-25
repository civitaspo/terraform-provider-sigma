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
