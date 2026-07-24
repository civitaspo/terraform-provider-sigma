package sigma_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
)

func TestWorkspaceFileAndGrantClientMethods(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	write := func(response http.ResponseWriter, value any) {
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(value)
	}
	workspace := sigma.Workspace{WorkspaceID: "workspace-1", WorkspaceURLID: "url-1", Name: "Analytics"}
	owner := "member-1"
	file := sigma.File{ID: "file-1", URLID: "file-url-1", Name: "Managed", Type: "folder", OwnerID: &owner}
	grant := sigma.Grant{GrantID: "grant-1", InodeID: "file-1", MemberID: &owner, Permission: "view", InodeType: "folder"}

	mock.Mux.HandleFunc("/v2/workspaces", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Errorf("unexpected method %s", request.Method)
		}
		write(response, workspace)
	})
	mock.Mux.HandleFunc("/v2.1/workspaces", func(response http.ResponseWriter, _ *http.Request) {
		write(response, map[string]any{"entries": []sigma.Workspace{workspace}, "nextPage": nil})
	})
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
		write(response, map[string]any{"entries": []sigma.Grant{grant}, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/workspaces/workspace-1/grants/grant-1", func(response http.ResponseWriter, _ *http.Request) {
		write(response, map[string]any{})
	})
	mock.Mux.HandleFunc("/v2/files", func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPost {
			write(response, file)
			return
		}
		if got := request.URL.Query().Get("parentId"); got != "workspace-1" {
			t.Errorf("parentId = %q", got)
		}
		write(response, map[string]any{"entries": []sigma.File{file}, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/files/file-1", func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodDelete {
			write(response, map[string]any{})
			return
		}
		write(response, file)
	})
	mock.Mux.HandleFunc("/v2/grants", func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPost {
			write(response, grant)
			return
		}
		write(response, map[string]any{"entries": []sigma.Grant{grant}, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/grants/grant-1", func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet {
			write(response, grant)
			return
		}
		write(response, grant)
	})
	for _, kind := range []string{"workbooks", "reports"} {
		kind := kind
		mock.Mux.HandleFunc("/v2/"+kind+"/file-1/grants", func(response http.ResponseWriter, _ *http.Request) {
			write(response, map[string]any{})
		})
		mock.Mux.HandleFunc("/v2/"+kind+"/file-1/grants/grant-1", func(response http.ResponseWriter, _ *http.Request) {
			write(response, map[string]any{})
		})
	}

	client, err := sigma.NewClient(mock.URL(), mock.ClientID, mock.ClientSecret)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err = client.CreateWorkspace(ctx, sigma.CreateWorkspaceInput{Name: workspace.Name}); err != nil {
		t.Fatal(err)
	}
	if _, err = client.GetWorkspace(ctx, workspace.WorkspaceID); err != nil {
		t.Fatal(err)
	}
	if _, err = client.UpdateWorkspace(ctx, workspace.WorkspaceID, sigma.UpdateWorkspaceInput{Name: workspace.Name}); err != nil {
		t.Fatal(err)
	}
	if values, listErr := client.ListWorkspaces(ctx); listErr != nil || len(values) != 1 {
		t.Fatalf("ListWorkspaces() = %v, %v", values, listErr)
	}
	if err = client.CreateWorkspaceGrant(ctx, workspace.WorkspaceID, sigma.Grantee{MemberID: owner}, "view"); err != nil {
		t.Fatal(err)
	}
	if values, listErr := client.ListWorkspaceGrants(ctx, workspace.WorkspaceID); listErr != nil || len(values) != 1 {
		t.Fatalf("ListWorkspaceGrants() = %v, %v", values, listErr)
	}
	if err = client.DeleteWorkspaceGrant(ctx, workspace.WorkspaceID, grant.GrantID); err != nil {
		t.Fatal(err)
	}
	if err = client.DeleteWorkspace(ctx, workspace.WorkspaceID); err != nil {
		t.Fatal(err)
	}

	if _, err = client.CreateFile(ctx, sigma.CreateFileInput{Type: "folder", Name: file.Name}); err != nil {
		t.Fatal(err)
	}
	if _, err = client.GetFile(ctx, file.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = client.UpdateFile(ctx, file.ID, sigma.UpdateFileInput{Name: &file.Name}); err != nil {
		t.Fatal(err)
	}
	if values, listErr := client.ListFiles(ctx, sigma.ListFilesOptions{ParentID: workspace.WorkspaceID}); listErr != nil || len(values) != 1 {
		t.Fatalf("ListFiles() = %v, %v", values, listErr)
	}
	if err = client.DeleteFile(ctx, file.ID); err != nil {
		t.Fatal(err)
	}

	if _, err = client.CreateGrant(ctx, sigma.CreateGrantInput{Grantee: sigma.Grantee{MemberID: owner}, Permission: "view", InodeID: file.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err = client.GetGrant(ctx, grant.GrantID); err != nil {
		t.Fatal(err)
	}
	if values, listErr := client.ListGrants(ctx, file.ID); listErr != nil || len(values) != 1 {
		t.Fatalf("ListGrants() = %v, %v", values, listErr)
	}
	if err = client.DeleteGrant(ctx, grant.GrantID); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"workbooks", "reports"} {
		if err = client.CreateDocumentGrant(ctx, kind, file.ID, sigma.Grantee{MemberID: owner}, "view", ""); err != nil {
			t.Fatal(err)
		}
		if err = client.DeleteDocumentGrant(ctx, kind, file.ID, grant.GrantID); err != nil {
			t.Fatal(err)
		}
	}
}
