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

func TestWorkspaceFileAndGrantResources(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	write := func(response http.ResponseWriter, value any) {
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(value)
	}
	workspace := map[string]any{
		"workspaceId": "workspace-1", "workspaceUrlId": "workspace-url-1", "name": "Analytics",
		"createdBy": "member-admin", "updatedBy": "member-admin", "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
	}
	file := map[string]any{
		"id": "folder-1", "urlId": "folder-url-1", "name": "Managed", "type": "folder",
		"parentId": "workspace-1", "parentUrlId": "workspace-url-1", "permission": "edit", "path": "Analytics/Managed",
		"badge": nil, "isArchived": false, "description": "Terraform managed", "ownerId": "member-admin",
		"createdBy": "member-admin", "updatedBy": "member-admin", "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
	}
	grant := func(id, inode, member, team, permission, inodeType string) map[string]any {
		return map[string]any{
			"grantId": id, "inodeId": inode, "organizationId": "org-1", "memberId": nullable(member), "teamId": nullable(team),
			"permission": permission, "inodeType": inodeType, "createdBy": "member-admin", "updatedBy": "member-admin",
			"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
		}
	}
	workspaceGrant := grant("workspace-grant-1", "workspace-1", "member-1", "", "view", "workspace")
	genericGrant := grant("grant-1", "folder-1", "", "team-1", "organize", "folder")
	workbookGrant := grant("workbook-grant-1", "workbook-1", "member-2", "", "explore", "workbook")
	reportGrant := grant("report-grant-1", "report-1", "", "team-2", "edit", "report")

	mock.Mux.HandleFunc("/v2/workspaces", func(response http.ResponseWriter, _ *http.Request) { write(response, workspace) })
	mock.Mux.HandleFunc("/v2/workspaces/workspace-1", func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodDelete {
			write(response, map[string]any{})
			return
		}
		write(response, workspace)
	})
	mock.Mux.HandleFunc("/v2/workspaces/workspace-1/grants", func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPost {
			write(response, map[string]any{})
			return
		}
		write(response, map[string]any{"entries": []any{workspaceGrant}, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/workspaces/workspace-1/grants/workspace-grant-1", func(response http.ResponseWriter, _ *http.Request) {
		write(response, map[string]any{})
	})
	mock.Mux.HandleFunc("/v2/files", func(response http.ResponseWriter, _ *http.Request) { write(response, file) })
	mock.Mux.HandleFunc("/v2/files/folder-1", func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodDelete {
			write(response, map[string]any{})
			return
		}
		write(response, file)
	})
	mock.Mux.HandleFunc("/v2/grants", func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPost {
			write(response, genericGrant)
			return
		}
		var entry any = genericGrant
		switch request.URL.Query().Get("inodeId") {
		case "workbook-1":
			entry = workbookGrant
		case "report-1":
			entry = reportGrant
		}
		write(response, map[string]any{"entries": []any{entry}, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/grants/grant-1", func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet {
			write(response, genericGrant)
			return
		}
		write(response, genericGrant)
	})
	for path, value := range map[string]any{
		"/v2/workbooks/workbook-1/grants":                  map[string]any{},
		"/v2/workbooks/workbook-1/grants/workbook-grant-1": map[string]any{},
		"/v2/reports/report-1/grants":                      map[string]any{},
		"/v2/reports/report-1/grants/report-grant-1":       map[string]any{},
	} {
		value := value
		mock.Mux.HandleFunc(path, func(response http.ResponseWriter, _ *http.Request) { write(response, value) })
	}

	config := `
provider "sigma" {
  base_url     = "` + mock.URL() + `"
  client_id     = "` + mock.ClientID + `"
  client_secret = "` + mock.ClientSecret + `"
}

resource "sigma_workspace" "test" {
  name = "Analytics"
}

resource "sigma_workspace_grant" "test" {
  inode_id   = "workspace-1"
  member_id  = "member-1"
  permission = "view"
}

resource "sigma_file" "test" {
  name        = "Managed"
  type        = "folder"
  parent_id   = "workspace-1"
  description = "Terraform managed"
}

resource "sigma_grant" "test" {
  inode_id   = "folder-1"
  team_id    = "team-1"
  permission = "organize"
}

resource "sigma_workbook_grant" "test" {
  inode_id   = "workbook-1"
  member_id  = "member-2"
  permission = "explore"
}

resource "sigma_report_grant" "test" {
  inode_id   = "report-1"
  team_id    = "team-2"
  permission = "edit"
}
`
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_workspace.test", "id", "workspace-1"),
				resource.TestCheckResourceAttr("sigma_workspace_grant.test", "id", "workspace-grant-1"),
				resource.TestCheckResourceAttr("sigma_file.test", "id", "folder-1"),
				resource.TestCheckResourceAttr("sigma_grant.test", "id", "grant-1"),
				resource.TestCheckResourceAttr("sigma_workbook_grant.test", "id", "workbook-grant-1"),
				resource.TestCheckResourceAttr("sigma_report_grant.test", "id", "report-grant-1"),
			),
		}},
	})
}

func nullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}
