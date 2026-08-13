package provider_test

import (
	"encoding/json"
	"net/http"
	"regexp"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestFolderResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	file := folderFixture("Managed", "Terraform managed")
	var createBody map[string]any
	mock.Mux.HandleFunc("/v2/files", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		_ = json.NewDecoder(request.Body).Decode(&createBody)
		writeJSON(response, file)
	})
	mock.Mux.HandleFunc("/v2/files/folder-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodGet:
			writeJSON(response, file)
		case http.MethodDelete:
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	config := providerConfig(mock) + `
resource "sigma_folder" "test" {
  name        = "Managed"
  parent_id   = "workspace-1"
  description = "Terraform managed"
}
`
	resource.UnitTest(t, providerTestCase([]resource.TestStep{
		{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_folder.test", "id", "folder-1"),
				resource.TestCheckResourceAttr("sigma_folder.test", "name", "Managed"),
				resource.TestCheckResourceAttr("sigma_folder.test", "parent_id", "workspace-1"),
				resource.TestCheckResourceAttr("sigma_folder.test", "url_id", "folder-url-1"),
				resource.TestCheckResourceAttr("sigma_folder.test", "path", "Analytics/Managed"),
				resource.TestCheckResourceAttr("sigma_folder.test", "is_archived", "false"),
			),
		},
		{
			ResourceName:      "sigma_folder.test",
			ImportState:       true,
			ImportStateVerify: true,
		},
	}))
	if createBody["type"] != "folder" || createBody["name"] != "Managed" || createBody["parentId"] != "workspace-1" {
		t.Fatalf("create body = %#v", createBody)
	}
	if _, ok := createBody["source"]; ok {
		t.Fatalf("folder create unexpectedly included source: %#v", createBody)
	}
}

func TestFolderResourceUpdateSendsChangedFieldsOnly(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	file := folderFixture("Managed", "Terraform managed")
	var patchBodies []map[string]any
	mock.Mux.HandleFunc("/v2/files", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		writeJSON(response, file)
	})
	mock.Mux.HandleFunc("/v2/files/folder-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodGet:
			writeJSON(response, file)
		case http.MethodPatch:
			var payload map[string]any
			_ = json.NewDecoder(request.Body).Decode(&payload)
			patchBodies = append(patchBodies, payload)
			if name, ok := payload["name"].(string); ok {
				file["name"] = name
			}
			writeJSON(response, file)
		case http.MethodDelete:
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	createConfig := providerConfig(mock) + `
resource "sigma_folder" "test" {
  name        = "Managed"
  parent_id   = "workspace-1"
  description = "Terraform managed"
}
`
	updateConfig := providerConfig(mock) + `
resource "sigma_folder" "test" {
  name        = "Renamed"
  parent_id   = "workspace-1"
  description = "Terraform managed"
}
`
	resource.UnitTest(t, providerTestCase([]resource.TestStep{
		{Config: createConfig},
		{
			Config: updateConfig,
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("sigma_folder.test", plancheck.ResourceActionUpdate),
				},
			},
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_folder.test", "name", "Renamed"),
			),
		},
	}))
	if len(patchBodies) != 1 {
		t.Fatalf("PATCH calls = %d, want 1; bodies=%#v", len(patchBodies), patchBodies)
	}
	if patchBodies[0]["name"] != "Renamed" {
		t.Fatalf("name PATCH = %#v", patchBodies[0])
	}
	if _, ok := patchBodies[0]["description"]; ok {
		t.Fatalf("description unexpectedly present on name-only PATCH: %#v", patchBodies[0])
	}
	if _, ok := patchBodies[0]["restore"]; ok {
		t.Fatalf("restore unexpectedly present: %#v", patchBodies[0])
	}
}

func TestFolderResourceRejectsNonFolder(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	workbook := folderFixture("Copied", "")
	workbook["id"] = "workbook-1"
	workbook["type"] = "workbook"
	mock.Mux.HandleFunc("/v2/files", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		writeJSON(response, workbook)
	})
	config := providerConfig(mock) + `
resource "sigma_folder" "test" {
  name = "Copied"
}
`
	resource.UnitTest(t, providerTestCase([]resource.TestStep{{
		Config:      config,
		ExpectError: regexp.MustCompile(`not a folder`),
	}}))
}

func TestFolderResourceImportRejectsNonFolder(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	folder := folderFixture("Managed", "Terraform managed")
	workbook := folderFixture("Copied", "")
	workbook["id"] = "workbook-1"
	workbook["type"] = "workbook"
	mock.Mux.HandleFunc("/v2/files", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		writeJSON(response, folder)
	})
	mock.Mux.HandleFunc("/v2/files/folder-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method == http.MethodDelete {
			writeJSON(response, map[string]any{})
			return
		}
		writeJSON(response, folder)
	})
	mock.Mux.HandleFunc("/v2/files/workbook-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		writeJSON(response, workbook)
	})
	config := providerConfig(mock) + `
resource "sigma_folder" "test" {
  name        = "Managed"
  parent_id   = "workspace-1"
  description = "Terraform managed"
}
`
	resource.UnitTest(t, providerTestCase([]resource.TestStep{
		{Config: config},
		{
			Config:        config,
			ResourceName:  "sigma_folder.test",
			ImportState:   true,
			ImportStateId: "workbook-1",
			ExpectError:   regexp.MustCompile(`not a folder`),
		},
	}))
}

func TestFolderResourceRejectsRemovedFileArguments(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	config := providerConfig(mock) + `
resource "sigma_folder" "test" {
  name            = "Managed"
  type            = "folder"
  source_inode_id = "workbook-source"
  restore         = true
}
`
	resource.UnitTest(t, providerTestCase([]resource.TestStep{{
		Config:      config,
		ExpectError: regexp.MustCompile(`Unsupported argument`),
	}}))
}

func TestAccFolderResource(t *testing.T) { requireAcceptance(t) }

func folderFixture(name, description string) map[string]any {
	return map[string]any{
		"id": "folder-1", "urlId": "folder-url-1", "name": name, "type": "folder",
		"parentId": "workspace-1", "parentUrlId": "workspace-url-1", "permission": "edit", "path": "Analytics/" + name,
		"badge": nil, "isArchived": false, "description": description, "ownerId": "member-admin",
		"parentSourceUrlId": "", "createdBy": "member-admin", "updatedBy": "member-admin",
		"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
	}
}
