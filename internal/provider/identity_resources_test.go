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

	config := `
provider "sigma" {
  base_url     = "` + mock.URL() + `"
  client_id     = "` + mock.ClientID + `"
  client_secret = "` + mock.ClientSecret + `"
}

resource "sigma_user_attribute_team_assignment" "test" {
  user_attribute_id = "attribute-1"
  team_id            = "team-1"
  value              = "americas"
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
		},
	})
}
