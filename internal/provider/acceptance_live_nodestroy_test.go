package provider_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccLiveDisposableDocuments creates disposable workspace/folder/workbook/report
// fixtures, exercises grants/embed/schedules and remaining data sources against
// those fixtures, then tears down only what this test created.
func TestAccLiveDisposableDocuments(t *testing.T) {
	requireAcceptance(t)
	ctx := context.Background()
	client, err := sigma.NewClient(os.Getenv("SIGMA_BASE_URL"), os.Getenv("SIGMA_CLIENT_ID"), os.Getenv("SIGMA_CLIENT_SECRET"))
	if err != nil {
		t.Fatal(err)
	}
	cleanupTfAccWorkspaces(t, client, "tf-acc-keep-", "tf-acc-ws-")

	stamp := time.Now().UnixNano()
	workspace, err := client.CreateWorkspace(ctx, sigma.CreateWorkspaceInput{
		Name: fmt.Sprintf("tf-acc-ws-%d", stamp), NoDuplicates: true,
	})
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	folder, err := client.CreateFile(ctx, sigma.CreateFileInput{
		Type: "folder", Name: fmt.Sprintf("tf-acc-folder-%d", stamp), ParentID: workspace.WorkspaceID,
	})
	if err != nil {
		t.Fatalf("create folder: %v", err)
	}
	workbook, err := client.CreateFile(ctx, sigma.CreateFileInput{
		Type: "workbook", Name: fmt.Sprintf("tf-acc-wb-%d", stamp), ParentID: folder.ID,
	})
	if err != nil {
		t.Fatalf("create workbook via files API: %v", err)
	}
	report, err := client.CreateFile(ctx, sigma.CreateFileInput{
		Type: "report", Name: fmt.Sprintf("tf-acc-report-%d", stamp), ParentID: folder.ID,
	})
	if err != nil {
		t.Fatalf("create report via files API: %v", err)
	}
	t.Logf("fixtures workspace=%s folder=%s workbook=%s report=%s", workspace.WorkspaceID, folder.ID, workbook.ID, report.ID)
	optionalDocs := tryCreateOptionalDocuments(t, client, folder.ID)
	extraIDs := make([]string, 0, len(optionalDocs))
	for _, doc := range optionalDocs {
		extraIDs = append(extraIDs, doc.ID)
	}
	t.Cleanup(func() {
		cleanupDocumentTree(t, client, workspace.WorkspaceID, append(extraIDs, report.ID, workbook.ID, folder.ID))
	})

	teams, err := client.ListTeams(ctx, sigma.ListTeamsOptions{})
	if err != nil || len(teams) == 0 {
		t.Fatalf("list teams: %v (need an existing team as grantee)", err)
	}
	teamID := teams[0].TeamID

	config := accProviderBlock() + fmt.Sprintf(`
data "sigma_workspace" "created" { id = %q }
data "sigma_file" "folder" { id = %q }
data "sigma_file" "workbook" { id = %q }
data "sigma_file" "report" { id = %q }
data "sigma_workbook" "created" { id = %q }
data "sigma_report" "created" { id = %q }
data "sigma_workbook_materialization_schedules" "created" { workbook_id = %q }
resource "sigma_workspace_grant" "test" {
  inode_id   = %q
  team_id    = %q
  permission = "view"
}
resource "sigma_workbook_grant" "view" {
  inode_id   = %q
  team_id    = %q
  permission = "view"
}
resource "sigma_workbook_grant" "test" {
  inode_id   = %q
  team_id    = %q
  permission = "explore"
}
resource "sigma_report_grant" "test" {
  inode_id   = %q
  team_id    = %q
  permission = "edit"
}
resource "sigma_workbook_embed" "test" {
  workbook_id = %q
  embed_type  = "public"
  source_type = "workbook"
}
resource "sigma_workbook_schedule" "test" {
  workbook_id = %q
  config_json = jsonencode({
    target   = [{ teamId = %q }]
    schedule = { cronSpec = "0 12 * * *", timezone = "UTC" }
    configV2 = {
      title       = "tf-acc-schedule"
      messageBody = "tf-acc disposable workbook schedule"
      exportAttachments = [{
        formatOptions        = { type = "PDF" }
        workbookExportSource = { type = "all" }
      }]
    }
  })
}
resource "sigma_report_schedule" "test" {
  report_id = %q
  config_json = jsonencode({
    target   = [{ teamId = %q }]
    schedule = { cronSpec = "0 12 * * *", timezone = "UTC" }
    configV2 = {
      title       = "tf-acc-report-schedule"
      messageBody = "tf-acc disposable report schedule"
      exportAttachments = [{
        formatOptions = { type = "PDF" }
      }]
    }
  })
}
`, workspace.WorkspaceID, folder.ID, workbook.ID, report.ID, workbook.ID, report.ID, workbook.ID,
		workspace.WorkspaceID, teamID, workbook.ID, teamID, workbook.ID, teamID, report.ID, teamID,
		workbook.ID, workbook.ID, teamID, report.ID, teamID)

	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("data.sigma_workspace.created", "id", workspace.WorkspaceID),
		resource.TestCheckResourceAttr("data.sigma_file.workbook", "type", "workbook"),
		resource.TestCheckResourceAttr("data.sigma_file.report", "type", "report"),
		resource.TestCheckResourceAttrSet("data.sigma_workbook.created", "name"),
		resource.TestCheckResourceAttrSet("data.sigma_report.created", "name"),
		resource.TestCheckResourceAttrSet("sigma_workspace_grant.test", "id"),
		resource.TestCheckResourceAttrSet("sigma_workbook_grant.view", "id"),
		resource.TestCheckResourceAttrSet("sigma_workbook_grant.test", "id"),
		resource.TestCheckResourceAttrSet("sigma_report_grant.test", "id"),
		resource.TestCheckResourceAttr("sigma_workbook_embed.test", "embed_type", "public"),
		resource.TestCheckResourceAttrSet("sigma_workbook_schedule.test", "id"),
		resource.TestCheckResourceAttrSet("sigma_report_schedule.test", "id"),
	}
	for _, doc := range optionalDocs {
		switch doc.Type {
		case "data_model":
			config += fmt.Sprintf("\ndata \"sigma_data_model\" \"created\" { id = %q }\n", doc.ID)
			checks = append(checks, resource.TestCheckResourceAttrSet("data.sigma_data_model.created", "id"))
		case "dataset":
			config += fmt.Sprintf("\ndata \"sigma_dataset\" \"created\" { id = %q }\n", doc.ID)
			checks = append(checks, resource.TestCheckResourceAttrSet("data.sigma_dataset.created", "id"))
		case "template":
			config += fmt.Sprintf("\ndata \"sigma_template\" \"created\" { id = %q }\n", doc.ID)
			checks = append(checks, resource.TestCheckResourceAttrSet("data.sigma_template.created", "id"))
		}
	}
	if path, connID, _, ok := peekFirstConnectionPath(t); ok {
		config += fmt.Sprintf(`
data "sigma_connection_path" "first" {
  connection_id = %q
  path          = [%s]
}
`, connID, quotePath(path))
		checks = append(checks, resource.TestCheckResourceAttrSet("data.sigma_connection_path.first", "id"))
	}

	resource.Test(t, providerTestCase([]resource.TestStep{{
		Config: config,
		Check:  resource.ComposeAggregateTestCheckFunc(checks...),
	}}))
}

type optionalDocument struct {
	Type string
	ID   string
}

func tryCreateOptionalDocuments(t *testing.T, client *sigma.Client, parentID string) []optionalDocument {
	t.Helper()
	ctx := context.Background()
	stamp := time.Now().UnixNano()
	var docs []optionalDocument
	for _, fileType := range []string{"data_model", "dataset", "template"} {
		file, err := client.CreateFile(ctx, sigma.CreateFileInput{
			Type: fileType, Name: fmt.Sprintf("tf-acc-%s-%d", fileType, stamp), ParentID: parentID,
		})
		if err != nil {
			t.Logf("CreateFile type=%s is not available: %v", fileType, err)
			continue
		}
		t.Logf("created disposable %s %s", fileType, file.ID)
		docs = append(docs, optionalDocument{Type: fileType, ID: file.ID})
	}
	return docs
}

func cleanupTfAccWorkspaces(t *testing.T, client *sigma.Client, prefixes ...string) {
	t.Helper()
	ctx := context.Background()
	workspaces, err := client.ListWorkspaces(ctx, sigma.ListWorkspacesOptions{})
	if err != nil {
		t.Logf("list workspaces for leftover cleanup: %v", err)
		return
	}
	for _, workspace := range workspaces {
		keep := false
		for _, prefix := range prefixes {
			if strings.HasPrefix(workspace.Name, prefix) {
				keep = true
				break
			}
		}
		if !keep {
			continue
		}
		t.Logf("removing leftover test workspace %s %s", workspace.WorkspaceID, workspace.Name)
		deleteFilesRecursive(t, client, workspace.WorkspaceID)
		if err := client.DeleteWorkspace(ctx, workspace.WorkspaceID); err != nil {
			t.Logf("delete leftover workspace %s: %v", workspace.Name, err)
		}
	}
}

func deleteFilesRecursive(t *testing.T, client *sigma.Client, parentID string) {
	t.Helper()
	ctx := context.Background()
	direct := true
	files, err := client.ListFiles(ctx, sigma.ListFilesOptions{ParentID: &parentID, DirectChildrenOnly: &direct})
	if err != nil {
		t.Logf("list files under %s: %v", parentID, err)
		return
	}
	for _, file := range files {
		if file.Type == "folder" {
			deleteFilesRecursive(t, client, file.ID)
		}
		if err := client.DeleteFile(ctx, file.ID); err != nil {
			t.Logf("delete file %s: %v", file.Name, err)
		}
	}
}

func cleanupDocumentTree(t *testing.T, client *sigma.Client, workspaceID string, fileIDs []string) {
	t.Helper()
	ctx := context.Background()
	if grants, err := client.ListWorkspaceGrants(ctx, workspaceID); err == nil {
		for _, grant := range grants {
			_ = client.DeleteWorkspaceGrant(ctx, workspaceID, grant.GrantID)
		}
	}
	for _, id := range fileIDs {
		if err := client.DeleteFile(ctx, id); err != nil {
			t.Logf("delete fixture %s: %v", id, err)
		}
	}
	if err := client.DeleteWorkspace(ctx, workspaceID); err != nil {
		t.Logf("delete fixture workspace %s: %v", workspaceID, err)
	}
}

func peekFirstConnectionPath(t *testing.T) (path []string, connectionID, urlID string, ok bool) {
	t.Helper()
	base := strings.TrimRight(os.Getenv("SIGMA_BASE_URL"), "/")
	form := url.Values{"grant_type": {"client_credentials"}, "client_id": {os.Getenv("SIGMA_CLIENT_ID")}, "client_secret": {os.Getenv("SIGMA_CLIENT_SECRET")}}
	tokenResp, err := http.PostForm(base+"/v2/auth/token", form)
	if err != nil {
		t.Logf("skip connection_path: token: %v", err)
		return nil, "", "", false
	}
	defer func() { _ = tokenResp.Body.Close() }()
	var tokenBody struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenBody); err != nil || tokenBody.AccessToken == "" {
		t.Log("skip connection_path: token decode")
		return nil, "", "", false
	}
	get := func(path string) (map[string]any, error) {
		req, err := http.NewRequest(http.MethodGet, base+path, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+tokenBody.AccessToken)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer func() { _ = resp.Body.Close() }()
		var body map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			return nil, err
		}
		return body, nil
	}
	conns, err := get("/v2/connections")
	if err != nil {
		t.Logf("skip connection_path: list connections: %v", err)
		return nil, "", "", false
	}
	entries, _ := conns["entries"].([]any)
	if len(entries) == 0 {
		return nil, "", "", false
	}
	first, _ := entries[0].(map[string]any)
	connectionID, _ = first["connectionId"].(string)
	if connectionID == "" {
		return nil, "", "", false
	}
	paths, err := get("/v2/connections/paths?" + url.Values{"connectionId": {connectionID}}.Encode())
	if err != nil {
		t.Logf("skip connection_path: first page: %v", err)
		return nil, "", "", false
	}
	pathEntries, _ := paths["entries"].([]any)
	if len(pathEntries) == 0 {
		return nil, "", "", false
	}
	item, _ := pathEntries[0].(map[string]any)
	urlID, _ = item["urlId"].(string)
	raw, _ := item["path"].([]any)
	for _, part := range raw {
		if s, ok := part.(string); ok {
			path = append(path, s)
		}
	}
	return path, connectionID, urlID, len(path) > 0 && urlID != ""
}

func quotePath(parts []string) string {
	quoted := make([]string, 0, len(parts))
	for _, part := range parts {
		quoted = append(quoted, fmt.Sprintf("%q", part))
	}
	return strings.Join(quoted, ", ")
}
