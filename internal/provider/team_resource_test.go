package provider_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestTeamResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	team := map[string]any{
		"teamId": "team-1", "name": "Analytics", "description": "Core analytics",
		"visibility": "private", "isArchived": false, "workspaceId": "workspace-1",
	}
	mock.Mux.HandleFunc("/v2/teams", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		var body map[string]any
		_ = json.NewDecoder(request.Body).Decode(&body)
		if body["name"] != "Analytics" || body["description"] != "Core analytics" || body["visibility"] != "private" {
			t.Errorf("create team body = %#v", body)
		}
		if _, ok := body["createTeamFolder"]; ok {
			t.Errorf("createTeamFolder unexpectedly set: %#v", body["createTeamFolder"])
		}
		writeJSON(response, team)
	})
	mock.Mux.HandleFunc("/v2/teams/team-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodGet:
			writeJSON(response, team)
		case http.MethodPatch:
			var body map[string]any
			_ = json.NewDecoder(request.Body).Decode(&body)
			if body["name"] != "Analytics Updated" {
				t.Errorf("update team name = %#v", body["name"])
			}
			if _, ok := body["createTeamFolder"]; ok {
				t.Errorf("update sent createTeamFolder: %#v", body["createTeamFolder"])
			}
			team["name"] = body["name"]
			team["description"] = body["description"]
			writeJSON(response, team)
		case http.MethodDelete:
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})

	createConfig := identityProviderConfig(mock) + `
resource "sigma_team" "test" {
  name        = "Analytics"
  description = "Core analytics"
  visibility  = "private"
}
`
	updateConfig := identityProviderConfig(mock) + `
resource "sigma_team" "test" {
  name        = "Analytics Updated"
  description = "Core analytics"
  visibility  = "private"
}
`
	resource.UnitTest(t, identityTestCase([]resource.TestStep{
		{
			Config: createConfig,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_team.test", "id", "team-1"),
				resource.TestCheckResourceAttr("sigma_team.test", "name", "Analytics"),
				resource.TestCheckResourceAttr("sigma_team.test", "visibility", "private"),
				resource.TestCheckResourceAttr("sigma_team.test", "workspace_id", "workspace-1"),
				resource.TestCheckNoResourceAttr("sigma_team.test", "create_team_folder"),
			),
		},
		{
			Config: updateConfig,
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("sigma_team.test", plancheck.ResourceActionUpdate),
				},
			},
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_team.test", "name", "Analytics Updated"),
				resource.TestCheckResourceAttr("sigma_team.test", "workspace_id", "workspace-1"),
			),
		},
		{
			ResourceName:      "sigma_team.test",
			ImportState:       true,
			ImportStateVerify: true,
		},
	}))
}

func TestTeamResourceCreateTeamFolder(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	var mu sync.Mutex
	team := map[string]any{
		"teamId": "team-1", "name": "Analytics", "description": "Core analytics",
		"visibility": "private", "isArchived": false, "workspaceId": "workspace-1",
	}
	var createBodies []map[string]any

	mock.Mux.HandleFunc("/v2/teams", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		var body map[string]any
		_ = json.NewDecoder(request.Body).Decode(&body)
		mu.Lock()
		createBodies = append(createBodies, body)
		folder, _ := body["createTeamFolder"].(bool)
		if folder {
			team["workspaceId"] = "workspace-1"
		} else {
			team["workspaceId"] = nil
		}
		mu.Unlock()
		writeJSON(response, team)
	})
	mock.Mux.HandleFunc("/v2/teams/team-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		mu.Lock()
		defer mu.Unlock()
		switch request.Method {
		case http.MethodGet:
			writeJSON(response, team)
		case http.MethodPatch:
			var body map[string]any
			_ = json.NewDecoder(request.Body).Decode(&body)
			if _, ok := body["createTeamFolder"]; ok {
				t.Errorf("update sent createTeamFolder: %#v", body["createTeamFolder"])
			}
			team["name"] = body["name"]
			writeJSON(response, team)
		case http.MethodDelete:
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})

	configWith := func(name string, createFolder bool) string {
		return identityProviderConfig(mock) + `
resource "sigma_team" "test" {
  name               = "` + name + `"
  description        = "Core analytics"
  visibility         = "private"
  create_team_folder = ` + strconv.FormatBool(createFolder) + `
}
`
	}

	resource.UnitTest(t, identityTestCase([]resource.TestStep{
		{
			Config: configWith("Analytics", true),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_team.test", "id", "team-1"),
				resource.TestCheckResourceAttr("sigma_team.test", "create_team_folder", "true"),
				resource.TestCheckResourceAttr("sigma_team.test", "workspace_id", "workspace-1"),
			),
		},
		{
			Config: configWith("Analytics Renamed", true),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("sigma_team.test", plancheck.ResourceActionUpdate),
				},
			},
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_team.test", "name", "Analytics Renamed"),
				resource.TestCheckResourceAttr("sigma_team.test", "create_team_folder", "true"),
				resource.TestCheckResourceAttr("sigma_team.test", "workspace_id", "workspace-1"),
			),
		},
		{
			Config: configWith("Analytics Renamed", false),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("sigma_team.test", plancheck.ResourceActionDestroyBeforeCreate),
				},
			},
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_team.test", "create_team_folder", "false"),
				resource.TestCheckNoResourceAttr("sigma_team.test", "workspace_id"),
			),
		},
		{
			ResourceName:            "sigma_team.test",
			ImportState:             true,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: []string{"create_team_folder"},
		},
	}))

	mu.Lock()
	defer mu.Unlock()
	if len(createBodies) != 2 {
		t.Fatalf("create calls = %d, want 2; bodies=%#v", len(createBodies), createBodies)
	}
	if createBodies[0]["createTeamFolder"] != true {
		t.Errorf("first create createTeamFolder = %#v", createBodies[0]["createTeamFolder"])
	}
	if createBodies[1]["createTeamFolder"] != false {
		t.Errorf("replace create createTeamFolder = %#v", createBodies[1]["createTeamFolder"])
	}
}

func TestTeamResourcePreservesWorkspaceIDWhenGetOmitsIt(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	createResponse := map[string]any{
		"teamId": "team-1", "name": "Analytics", "description": "Core analytics",
		"visibility": "private", "isArchived": false, "workspaceId": "workspace-1",
	}
	getResponse := map[string]any{
		"teamId": "team-1", "name": "Analytics", "description": "Core analytics",
		"visibility": "private", "isArchived": false,
	}
	mock.Mux.HandleFunc("/v2/teams", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		var body map[string]any
		_ = json.NewDecoder(request.Body).Decode(&body)
		if body["createTeamFolder"] != true {
			t.Errorf("createTeamFolder = %#v", body["createTeamFolder"])
		}
		writeJSON(response, createResponse)
	})
	mock.Mux.HandleFunc("/v2/teams/team-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodGet:
			writeJSON(response, getResponse)
		case http.MethodDelete:
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})

	config := identityProviderConfig(mock) + `
resource "sigma_team" "test" {
  name               = "Analytics"
  description        = "Core analytics"
  visibility         = "private"
  create_team_folder = true
}
`
	resource.UnitTest(t, identityTestCase([]resource.TestStep{
		{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_team.test", "workspace_id", "workspace-1"),
				resource.TestCheckResourceAttr("sigma_team.test", "create_team_folder", "true"),
			),
		},
		{
			Config:             config,
			PlanOnly:           true,
			ExpectNonEmptyPlan: false,
		},
	}))
}

func TestAccTeamResource(t *testing.T) { requireAcceptance(t) }
