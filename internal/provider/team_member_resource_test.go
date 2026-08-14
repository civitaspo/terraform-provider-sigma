package provider_test

import (
	"encoding/json"
	"net/http"
	"sync"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

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

func TestTeamMemberResourceRead404RemovesState(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	var mu sync.Mutex
	members := map[string]bool{"member-1": true}
	gone := false
	mock.Mux.HandleFunc("/v2/teams/team-1/members", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		mu.Lock()
		defer mu.Unlock()
		switch request.Method {
		case http.MethodGet:
			if gone {
				writeNotFound(response)
				return
			}
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
		{Config: config},
		{
			PreConfig:          func() { mu.Lock(); gone = true; mu.Unlock() },
			RefreshState:       true,
			ExpectNonEmptyPlan: true,
		},
	}))
}

func TestAccTeamMemberResource(t *testing.T) { requireAcceptance(t) }
