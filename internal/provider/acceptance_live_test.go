package provider_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func accProviderBlock() string {
	return `
provider "sigma" {}
`
}

func accName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func runAccTag(t *testing.T) {
	t.Helper()
	requireAcceptance(t)
	name := accName("tf-acc-tag")
	resource.Test(t, providerTestCase([]resource.TestStep{{
		Config: accProviderBlock() + `
resource "sigma_tag" "test" {
  name  = "` + name + `"
  color = "cyan"
}
`,
		Check: resource.TestCheckResourceAttrSet("sigma_tag.test", "id"),
	}}))
}

func runAccWorkspaceGrant(t *testing.T) {
	t.Helper()
	requireAcceptanceEnv(t, "SIGMA_ACC_TEAM_ID")
	teamID := os.Getenv("SIGMA_ACC_TEAM_ID")
	resource.Test(t, providerTestCase([]resource.TestStep{{
		Config: accProviderBlock() + `
data "sigma_workspaces" "all" {}
resource "sigma_workspace_grant" "test" {
  inode_id   = data.sigma_workspaces.all.workspaces[0].id
  team_id    = "` + teamID + `"
  permission = "view"
}
`,
		Check: resource.TestCheckResourceAttrSet("sigma_workspace_grant.test", "id"),
	}}))
}

func runAccWorkbookGrant(t *testing.T) {
	t.Helper()
	requireAcceptanceEnv(t, "SIGMA_ACC_WORKBOOK_ID", "SIGMA_ACC_TEAM_ID", "SIGMA_ACC_TAG_ID")
	resource.Test(t, providerTestCase([]resource.TestStep{{
		Config: accProviderBlock() + fmt.Sprintf(`
resource "sigma_workbook_grant" "test" {
  inode_id   = %q
  team_id    = %q
  tag_id     = %q
  permission = "view"
}
`, os.Getenv("SIGMA_ACC_WORKBOOK_ID"), os.Getenv("SIGMA_ACC_TEAM_ID"), os.Getenv("SIGMA_ACC_TAG_ID")),
		Check: resource.TestCheckResourceAttrSet("sigma_workbook_grant.test", "id"),
	}}))
}

func runAccReportGrant(t *testing.T) {
	t.Helper()
	requireAcceptanceEnv(t, "SIGMA_ACC_REPORT_ID", "SIGMA_ACC_TEAM_ID", "SIGMA_ACC_TAG_ID")
	resource.Test(t, providerTestCase([]resource.TestStep{{
		Config: accProviderBlock() + fmt.Sprintf(`
resource "sigma_report_grant" "test" {
  inode_id   = %q
  team_id    = %q
  tag_id     = %q
  permission = "view"
}
`, os.Getenv("SIGMA_ACC_REPORT_ID"), os.Getenv("SIGMA_ACC_TEAM_ID"), os.Getenv("SIGMA_ACC_TAG_ID")),
		Check: resource.TestCheckResourceAttrSet("sigma_report_grant.test", "id"),
	}}))
}

func runAccConnection(t *testing.T) {
	t.Helper()
	requireAcceptanceEnv(t, "SIGMA_ACC_CONNECTION_ID")
	resource.Test(t, providerTestCase([]resource.TestStep{{
		Config: accProviderBlock() + fmt.Sprintf(`
data "sigma_connection" "test" {
  connection_id = %q
}
`, os.Getenv("SIGMA_ACC_CONNECTION_ID")),
		Check: resource.TestCheckResourceAttrSet("data.sigma_connection.test", "id"),
	}}))
}

func runAccWorkbookEmbed(t *testing.T) {
	t.Helper()
	requireAcceptanceEnv(t, "SIGMA_ACC_WORKBOOK_ID")
	resource.Test(t, providerTestCase([]resource.TestStep{{
		Config: accProviderBlock() + fmt.Sprintf(`
resource "sigma_workbook_embed" "test" {
  workbook_id = %q
  embed_type  = "public"
  source_type = "workbook"
}
`, os.Getenv("SIGMA_ACC_WORKBOOK_ID")),
		Check: resource.TestCheckResourceAttr("sigma_workbook_embed.test", "embed_type", "public"),
	}}))
}

func runAccWorkbookScheduleDrift(t *testing.T) {
	t.Helper()
	requireAcceptanceEnv(t, "SIGMA_ACC_WORKBOOK_ID")
	workbookID := os.Getenv("SIGMA_ACC_WORKBOOK_ID")
	config := accProviderBlock() + fmt.Sprintf(`
resource "sigma_workbook_schedule" "test" {
  workbook_id = %q
  config_json = jsonencode({
    target = [{
      type      = "email"
      recipient = "tf-acc@example.com"
    }]
    schedule = {
      cronSpec = "0 12 * * *"
    }
    configV2 = {
      title = "tf-acc-schedule"
    }
  })
}
`, workbookID)
	resource.Test(t, providerTestCase([]resource.TestStep{
		{Config: config, Check: resource.TestCheckResourceAttrSet("sigma_workbook_schedule.test", "id")},
		{Config: config, ExpectNonEmptyPlan: false},
	}))
}
