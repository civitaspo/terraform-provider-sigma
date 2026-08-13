package provider_test

import (
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

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

func TestAccTeamMembersResource(t *testing.T) { requireAcceptance(t) }
