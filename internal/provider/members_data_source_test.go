package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMembersDataSource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	member := map[string]any{
		"memberId": "member-1", "organizationId": "org-1", "email": "ada@example.com",
		"firstName": "Ada", "lastName": "Lovelace", "memberType": "Creator", "userKind": "internal",
		"isArchived": false, "isInactive": false, "homeFolderId": "folder-1",
	}
	write := func(response http.ResponseWriter, value any) {
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(value)
	}
	mock.Mux.HandleFunc("/v2/members", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		write(response, map[string]any{"entries": []any{member}, "nextPage": nil})
	})
	config := identityProviderConfig(mock) + `
data "sigma_members" "all" {}
`
	resource.UnitTest(t, identityTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.sigma_members.all", "members.#", "1"),
			resource.TestCheckResourceAttr("data.sigma_members.all", "members.0.home_folder_id", "folder-1"),
		),
	}}))
}
