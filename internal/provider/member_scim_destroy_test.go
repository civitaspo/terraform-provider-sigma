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
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestMemberResourceRefusesSCIMDestroy(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	member := map[string]any{
		"memberId": "member-scim", "organizationId": "org-1", "email": "scim@example.com",
		"firstName": "Scim", "lastName": "User", "memberType": "Creator", "userKind": "internal",
		"isArchived": false, "isInactive": false,
	}
	deleteCalls := 0
	mock.Mux.HandleFunc("/v2/members", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		_ = json.NewEncoder(response).Encode(member)
	})
	mock.Mux.HandleFunc("/v2/members/member-scim", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
			current := map[string]any{}
			for key, value := range member {
				current[key] = value
			}
			_ = json.NewEncoder(response).Encode(current)
		case http.MethodDelete:
			deleteCalls++
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
  email      = "scim@example.com"
  first_name = "Scim"
  last_name  = "User"
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
					resource.TestCheckResourceAttr("sigma_member.test", "id", "member-scim"),
					func(_ *terraform.State) error {
						member["isInactive"] = true
						return nil
					},
				),
			},
			{
				Config:      providerConfig,
				ExpectError: regexp.MustCompile(`Cannot deactivate SCIM-managed Sigma member`),
			},
			{
				PreConfig: func() {
					if deleteCalls != 0 {
						t.Fatalf("DeleteMember calls = %d, want 0", deleteCalls)
					}
					member["isInactive"] = false
				},
				Config: providerConfig,
			},
		},
	})
}
