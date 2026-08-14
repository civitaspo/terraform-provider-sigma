package provider_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func loadOpenAPIFixture(t *testing.T, name string) map[string]any {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(file), "testdata", "openapi", name))
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func TestFolderResourceContract(t *testing.T) {
	file := loadOpenAPIFixture(t, "files_folder.json")
	file["id"] = file["inodeId"]
	mock := testutil.NewRecordingSigma(t,
		testutil.ExpectedRequest{
			Method: "POST", Path: "/v2/files",
			JSONBody: map[string]any{
				"type": "folder", "name": "Managed", "parentId": "workspace-1", "description": "Terraform managed",
			},
			Response: file,
		},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/files/folder-1", Response: file},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/files/folder-1", Response: file},
		testutil.ExpectedRequest{Method: "DELETE", Path: "/v2/files/folder-1", Response: map[string]any{}},
	)
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
				resource.TestCheckResourceAttr("sigma_folder.test", "url_id", "folder-url-1"),
			),
		},
		{
			ResourceName:      "sigma_folder.test",
			ImportState:       true,
			ImportStateVerify: true,
		},
	}))
}

func TestFolderResourceUnknownParentMakesZeroRequests(t *testing.T) {
	mock := testutil.NewRecordingSigma(t)
	config := providerConfig(mock) + `
resource "sigma_folder" "test" {
  name      = "Managed"
  parent_id = timestamp()
}
`
	resource.UnitTest(t, providerTestCase([]resource.TestStep{{
		Config:             config,
		PlanOnly:           true,
		ExpectNonEmptyPlan: true,
	}}))
}
