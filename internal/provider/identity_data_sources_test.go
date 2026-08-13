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

func TestIdentityDataSources(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	member := map[string]any{
		"memberId": "member-1", "organizationId": "org-1", "email": "ada@example.com",
		"firstName": "Ada", "lastName": "Lovelace", "memberType": "Creator", "userKind": "internal",
		"isArchived": false, "isInactive": false, "homeFolderId": "folder-1",
	}
	team := map[string]any{
		"teamId": "team-1", "name": "Analytics", "description": "Core", "visibility": "private", "isArchived": false, "workspaceId": "workspace-1",
	}
	accountType := map[string]any{
		"accountTypeId": "type-1", "accountTypeName": "Analyst", "description": "Analyst access", "isCustom": true,
	}
	attribute := map[string]any{
		"userAttributeId": "attribute-1", "name": "Region", "description": "Sales region",
		"defaultValue": map[string]string{"val": "global", "type": "string"},
	}
	write := func(response http.ResponseWriter, value any) {
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(value)
	}

	mock.Mux.HandleFunc("/v2/members/member-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		write(response, member)
	})
	mock.Mux.HandleFunc("/v2/members", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		write(response, map[string]any{"entries": []any{member}, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/teams/team-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		write(response, team)
	})
	mock.Mux.HandleFunc("/v2.1/teams", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		write(response, map[string]any{"entries": []any{team}, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/accountTypes", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		write(response, map[string]any{"entries": []any{accountType}})
	})
	mock.Mux.HandleFunc("/v2/user-attributes", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		write(response, map[string]any{"entries": []any{attribute}, "nextPage": nil})
	})

	config := `
provider "sigma" {
  base_url      = "` + mock.URL() + `"
  client_id     = "` + mock.ClientID + `"
  client_secret = "` + mock.ClientSecret + `"
}

data "sigma_member" "one" {
  id = "member-1"
}
data "sigma_members" "all" {}
data "sigma_team" "one" {
  id = "team-1"
}
data "sigma_teams" "all" {}
data "sigma_account_types" "all" {}
data "sigma_user_attributes" "all" {}
`
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.sigma_member.one", "email", "ada@example.com"),
				resource.TestCheckResourceAttr("data.sigma_member.one", "home_folder_id", "folder-1"),
				resource.TestCheckResourceAttr("data.sigma_members.all", "members.#", "1"),
				resource.TestCheckResourceAttr("data.sigma_members.all", "members.0.home_folder_id", "folder-1"),
				resource.TestCheckResourceAttr("data.sigma_team.one", "name", "Analytics"),
				resource.TestCheckResourceAttr("data.sigma_team.one", "workspace_id", "workspace-1"),
				resource.TestCheckResourceAttr("data.sigma_teams.all", "teams.#", "1"),
				resource.TestCheckResourceAttr("data.sigma_teams.all", "teams.0.workspace_id", "workspace-1"),
				resource.TestCheckResourceAttr("data.sigma_account_types.all", "account_types.#", "1"),
				resource.TestCheckResourceAttr("data.sigma_user_attributes.all", "user_attributes.0.name", "Region"),
			),
		}},
	})
}
