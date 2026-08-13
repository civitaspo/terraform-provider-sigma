package provider_test

import (
	"encoding/json"
	"net/http"
	"regexp"
	"testing"

	sigmaprovider "github.com/civitaspo/terraform-provider-sigma/internal/provider"
	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func fileProviderConfig(mock *testutil.MockSigma) string {
	return `
provider "sigma" {
  base_url      = "` + mock.URL() + `"
  client_id     = "` + mock.ClientID + `"
  client_secret = "` + mock.ClientSecret + `"
}
`
}

func TestFileResourceWorkbookSourceAndRestore(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	file := map[string]any{
		"id": "workbook-1", "urlId": "workbook-url-1", "name": "Copied", "type": "workbook",
		"parentId": "folder-1", "parentUrlId": "folder-url-1", "permission": "edit", "path": "Copied",
		"badge": nil, "isArchived": false, "description": "", "ownerId": "member-admin",
		"createdBy": "member-admin", "updatedBy": "member-admin", "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
	}
	var createBody map[string]any
	var patchBodies []map[string]any
	mock.Mux.HandleFunc("/v2/files", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		_ = json.NewDecoder(request.Body).Decode(&createBody)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(file)
	})
	mock.Mux.HandleFunc("/v2/files/workbook-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(file)
		case http.MethodPatch:
			var payload map[string]any
			_ = json.NewDecoder(request.Body).Decode(&payload)
			patchBodies = append(patchBodies, payload)
			if payload["restore"] == true {
				file["isArchived"] = false
			}
			if name, ok := payload["name"].(string); ok {
				file["name"] = name
			}
			_ = json.NewEncoder(response).Encode(file)
		case http.MethodDelete:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})

	createConfig := fileProviderConfig(mock) + `
resource "sigma_file" "test" {
  type            = "workbook"
  name            = "Copied"
  parent_id       = "folder-1"
  source_inode_id = "workbook-source"
  source_version  = 3
}
`
	restoreConfig := fileProviderConfig(mock) + `
resource "sigma_file" "test" {
  type            = "workbook"
  name            = "Copied restored"
  parent_id       = "folder-1"
  source_inode_id = "workbook-source"
  source_version  = 3
  restore         = true
}
`
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: createConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sigma_file.test", "id", "workbook-1"),
					resource.TestCheckResourceAttr("sigma_file.test", "source_inode_id", "workbook-source"),
					resource.TestCheckResourceAttr("sigma_file.test", "source_version", "3"),
				),
			},
			{
				Config: restoreConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sigma_file.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sigma_file.test", "name", "Copied restored"),
					resource.TestCheckResourceAttr("sigma_file.test", "restore", "true"),
					resource.TestCheckResourceAttr("sigma_file.test", "is_archived", "false"),
				),
			},
			{
				ResourceName:            "sigma_file.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source_inode_id", "source_version", "restore"},
			},
		},
	})

	source, _ := createBody["source"].(map[string]any)
	if createBody["type"] != "workbook" || source["inodeId"] != "workbook-source" || source["version"] != float64(3) {
		t.Fatalf("create body = %#v", createBody)
	}
	if len(patchBodies) != 1 {
		t.Fatalf("PATCH calls = %d, want 1; bodies=%#v", len(patchBodies), patchBodies)
	}
	if patchBodies[0]["restore"] != true {
		t.Fatalf("restore PATCH = %#v", patchBodies[0])
	}
}

func TestFileResourceSourceRequiresWorkbook(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	config := fileProviderConfig(mock) + `
resource "sigma_file" "test" {
  type            = "folder"
  name            = "Managed"
  source_inode_id = "workbook-source"
  source_version  = 1
}
`
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{{
			Config:      config,
			ExpectError: regexp.MustCompile(`only valid when type is`),
		}},
	})
}

func TestFileResourceSourceFieldsMustBeTogether(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	config := fileProviderConfig(mock) + `
resource "sigma_file" "test" {
  type            = "workbook"
  name            = "Copied"
  source_inode_id = "workbook-source"
}
`
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{{
			Config:      config,
			ExpectError: regexp.MustCompile(`must be set together`),
		}},
	})
}
