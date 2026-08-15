package provider_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
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

func runAccWhoami(t *testing.T) {
	t.Helper()
	requireAcceptance(t)
	resource.Test(t, providerTestCase([]resource.TestStep{{
		Config: accProviderBlock() + `
data "sigma_whoami" "test" {}
data "sigma_member" "self" {
  id = data.sigma_whoami.test.user_id
}
data "sigma_members" "all" {}
`,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrSet("data.sigma_whoami.test", "user_id"),
			resource.TestCheckResourceAttrSet("data.sigma_whoami.test", "organization_id"),
			resource.TestCheckResourceAttrSet("data.sigma_member.self", "id"),
			resource.TestCheckResourceAttrSet("data.sigma_members.all", "id"),
		),
	}}))
}

func runAccReadOnlyCatalog(t *testing.T) {
	t.Helper()
	requireAcceptance(t)
	resource.Test(t, providerTestCase([]resource.TestStep{{
		Config: accProviderBlock() + `
data "sigma_workspaces" "all" {}
data "sigma_workspace" "first" {
  id = data.sigma_workspaces.all.workspaces[0].id
}
data "sigma_teams" "all" {}
data "sigma_team" "first" {
  id = data.sigma_teams.all.teams[0].id
}
data "sigma_account_types" "all" {}
data "sigma_user_attributes" "all" {}
data "sigma_connections" "all" {}
data "sigma_workbooks" "all" {}
data "sigma_reports" "all" {}
data "sigma_data_models" "all" {}
data "sigma_datasets" "all" {}
data "sigma_templates" "all" {}
data "sigma_tags" "all" {}
data "sigma_files" "all" {}
data "sigma_tenants" "all" {}
data "sigma_deployment_policies" "all" {}
`,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrSet("data.sigma_workspaces.all", "id"),
			resource.TestCheckResourceAttrSet("data.sigma_workspace.first", "id"),
			resource.TestCheckResourceAttrSet("data.sigma_teams.all", "id"),
			resource.TestCheckResourceAttrSet("data.sigma_team.first", "id"),
			resource.TestCheckResourceAttrSet("data.sigma_account_types.all", "id"),
			resource.TestCheckResourceAttrSet("data.sigma_user_attributes.all", "id"),
			resource.TestCheckResourceAttrSet("data.sigma_connections.all", "id"),
			resource.TestCheckResourceAttrSet("data.sigma_workbooks.all", "id"),
			resource.TestCheckResourceAttrSet("data.sigma_reports.all", "id"),
			resource.TestCheckResourceAttrSet("data.sigma_data_models.all", "id"),
			resource.TestCheckResourceAttrSet("data.sigma_datasets.all", "id"),
			resource.TestCheckResourceAttrSet("data.sigma_templates.all", "id"),
			resource.TestCheckResourceAttrSet("data.sigma_tags.all", "id"),
			resource.TestCheckResourceAttrSet("data.sigma_files.all", "id"),
			resource.TestCheckResourceAttrSet("data.sigma_tenants.all", "id"),
			resource.TestCheckResourceAttrSet("data.sigma_deployment_policies.all", "id"),
		),
	}}))
}

func runAccConnectionReadOnly(t *testing.T) {
	t.Helper()
	requireAcceptance(t)
	resource.Test(t, providerTestCase([]resource.TestStep{{
		Config: accProviderBlock() + `
data "sigma_connections" "all" {}
data "sigma_connection" "first" {
  id = data.sigma_connections.all.connections[0].id
}
`,
		Check: resource.TestCheckResourceAttrSet("data.sigma_connection.first", "id"),
	}}))
}

func runAccConnectionPaths(t *testing.T) {
	t.Helper()
	requireAcceptance(t)
	resource.Test(t, providerTestCase([]resource.TestStep{{
		Config: accProviderBlock() + `
data "sigma_connections" "all" {}
data "sigma_connection_paths" "all" {
  connection_id = data.sigma_connections.all.connections[0].id
}
`,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrSet("data.sigma_connection_paths.all", "id"),
			resource.TestCheckResourceAttrSet("data.sigma_connection_paths.all", "paths.#"),
		),
	}}))
}

func runAccConnectionGrants(t *testing.T) {
	t.Helper()
	requireAcceptance(t)
	path, connID, urlID, ok := peekFirstConnectionPath(t)
	if !ok {
		t.Fatal("need at least one connection path to grant")
	}
	resource.Test(t, providerTestCase([]resource.TestStep{{
		Config: accProviderBlock() + fmt.Sprintf(`
data "sigma_teams" "all" {}
data "sigma_connection_path" "first" {
  connection_id = %q
  path          = [%s]
}
resource "sigma_connection_grant" "test" {
  connection_id = %q
  team_id       = data.sigma_teams.all.teams[0].id
  permission    = "annotate"
}
resource "sigma_connection_path_grant" "test" {
  connection_path_id = %q
  team_id            = data.sigma_teams.all.teams[0].id
  permission         = "annotate"
}
`, connID, quotePath(path), connID, urlID),
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrSet("sigma_connection_grant.test", "id"),
			resource.TestCheckResourceAttrSet("sigma_connection_path_grant.test", "id"),
		),
	}}))
}

func runAccOwnedWorkspaceFolderGrant(t *testing.T) {
	t.Helper()
	requireAcceptance(t)
	name := accName("tf-acc-ws")
	folder := accName("tf-acc-folder")
	config := accProviderBlock() + `
data "sigma_teams" "all" {}
resource "sigma_workspace" "test" {
  name          = "` + name + `"
  no_duplicates = true
}
resource "sigma_folder" "test" {
  name      = "` + folder + `"
  parent_id = sigma_workspace.test.id
}
resource "sigma_workspace_grant" "test" {
  inode_id   = sigma_workspace.test.id
  team_id    = data.sigma_teams.all.teams[0].id
  permission = "view"
}
data "sigma_workspace" "created" {
  id = sigma_workspace.test.id
}
data "sigma_file" "folder" {
  id = sigma_folder.test.id
}
`
	resource.Test(t, providerTestCase([]resource.TestStep{
		{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttrSet("sigma_workspace.test", "id"),
				resource.TestCheckResourceAttrSet("sigma_folder.test", "id"),
				resource.TestCheckResourceAttrSet("sigma_workspace_grant.test", "id"),
				resource.TestCheckResourceAttrSet("data.sigma_workspace.created", "name"),
				resource.TestCheckResourceAttr("data.sigma_file.folder", "type", "folder"),
			),
		},
		{
			ResourceName:            "sigma_workspace.test",
			ImportState:             true,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: []string{"no_duplicates"},
		},
	}))
}

func runAccUserAttributeAndAssignment(t *testing.T) {
	t.Helper()
	requireAcceptance(t)
	name := accName("tf-acc-attr")
	resource.Test(t, providerTestCase([]resource.TestStep{{
		Config: accProviderBlock() + `
data "sigma_whoami" "me" {}
data "sigma_teams" "all" {}
resource "sigma_user_attribute" "test" {
  name          = "` + name + `"
  description   = "tf-acc disposable attribute"
  default_value = "unset"
}
resource "sigma_user_attribute_user_assignment" "test" {
  user_attribute_id = sigma_user_attribute.test.id
  user_id           = data.sigma_whoami.me.user_id
  value             = "tf-acc"
}
resource "sigma_user_attribute_team_assignment" "test" {
  user_attribute_id = sigma_user_attribute.test.id
  team_id           = data.sigma_teams.all.teams[0].id
  value             = "tf-acc-team"
}
data "sigma_user_attribute" "created" {
  id = sigma_user_attribute.test.id
}
`,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrSet("sigma_user_attribute.test", "id"),
			resource.TestCheckResourceAttrSet("sigma_user_attribute_user_assignment.test", "id"),
			resource.TestCheckResourceAttrSet("sigma_user_attribute_team_assignment.test", "id"),
			resource.TestCheckResourceAttr("data.sigma_user_attribute.created", "name", name),
		),
	}}))
}

func runAccAccountType(t *testing.T) {
	t.Helper()
	requireAcceptance(t)
	name := accName("tf-acc-atype")
	resource.Test(t, providerTestCase([]resource.TestStep{{
		Config: accProviderBlock() + `
data "sigma_account_types" "all" {}
resource "sigma_account_type" "test" {
  name                          = "` + name + `"
  description                   = "tf-acc disposable account type"
  permissions                   = ["comment"]
  reassign_to_account_type_id   = data.sigma_account_types.all.account_types[0].id
}
data "sigma_account_type_permissions" "test" {
  account_type_id = sigma_account_type.test.id
}
`,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrSet("sigma_account_type.test", "id"),
			resource.TestCheckResourceAttrSet("data.sigma_account_type_permissions.test", "id"),
		),
	}}))
}

func runAccAPICredentialAndConnector(t *testing.T) {
	t.Helper()
	requireAcceptance(t)
	name := accName("tf-acc-cred")
	connector := accName("tf-acc-conn")
	resource.Test(t, providerTestCase([]resource.TestStep{{
		Config: accProviderBlock() + `
resource "sigma_api_credential" "test" {
  name      = "` + name + `"
  allowlist = ["example.com"]
  credential_wo = jsonencode({
    authMethod = "bearer"
    bearer = {
      token = "tf-acc-not-a-real-token"
    }
  })
  credential_wo_version = 1
}
resource "sigma_api_connector" "test" {
  name    = "` + connector + `"
  auth_id = sigma_api_credential.test.id
  params_json = jsonencode({
    method      = "GET"
    url         = "https://example.com/tf-acc"
    headers     = []
    pathParams  = []
    queryParams = []
    body        = ""
  })
}
`,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrSet("sigma_api_credential.test", "id"),
			resource.TestCheckResourceAttrSet("sigma_api_connector.test", "id"),
		),
	}}))
}

func TestAccWhoamiAndMemberDataSources(t *testing.T) { runAccWhoami(t) }

func TestAccReadOnlyCatalogDataSources(t *testing.T) { runAccReadOnlyCatalog(t) }

func TestAccConnectionReadOnlyDataSources(t *testing.T) { runAccConnectionReadOnly(t) }

func runAccDeploymentPolicyDocument(t *testing.T) {
	t.Helper()
	requireAcceptance(t)
	ctx := context.Background()
	client, err := sigma.NewClient(os.Getenv("SIGMA_BASE_URL"), os.Getenv("SIGMA_CLIENT_ID"), os.Getenv("SIGMA_CLIENT_SECRET"))
	if err != nil {
		t.Fatal(err)
	}
	stamp := time.Now().UnixNano()
	workspace, err := client.CreateWorkspace(ctx, sigma.CreateWorkspaceInput{
		Name: fmt.Sprintf("tf-acc-ws-%d", stamp), NoDuplicates: true,
	})
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	workbook, err := client.CreateFile(ctx, sigma.CreateFileInput{
		Type: "workbook", Name: fmt.Sprintf("tf-acc-wb-%d", stamp), ParentID: workspace.WorkspaceID,
	})
	if err != nil {
		t.Fatalf("create workbook: %v", err)
	}
	t.Cleanup(func() {
		cleanupDocumentTree(t, client, workspace.WorkspaceID, []string{workbook.ID})
	})
	tag := accName("tf-acc-dtag")
	name := accName("tf-acc-dpol")
	resource.Test(t, providerTestCase([]resource.TestStep{{
		Config: accProviderBlock() + fmt.Sprintf(`
resource "sigma_tag" "test" {
  name  = %q
  color = "cyan"
}
resource "sigma_deployment_policy" "test" {
  name           = %q
  version_tag_id = sigma_tag.test.id
}
resource "sigma_deployment_policy_document" "test" {
  deployment_policy_id = sigma_deployment_policy.test.id
  inode_id             = %q
}
`, tag, name, workbook.ID),
		Check: resource.TestCheckResourceAttrSet("sigma_deployment_policy_document.test", "id"),
	}}))
}

func runAccDeploymentPolicy(t *testing.T) {
	t.Helper()
	requireAcceptance(t)
	tag := accName("tf-acc-dtag")
	name := accName("tf-acc-dpol")
	resource.Test(t, providerTestCase([]resource.TestStep{{
		Config: accProviderBlock() + `
resource "sigma_tag" "test" {
  name  = "` + tag + `"
  color = "cyan"
}
resource "sigma_deployment_policy" "test" {
  name           = "` + name + `"
  version_tag_id = sigma_tag.test.id
}
data "sigma_deployment_policy" "created" {
  id = sigma_deployment_policy.test.id
}
`,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrSet("sigma_deployment_policy.test", "id"),
			resource.TestCheckResourceAttr("data.sigma_deployment_policy.created", "name", name),
		),
	}}))
}

func runAccSourceSwapPolicy(t *testing.T) {
	t.Helper()
	requireAcceptance(t)
	name := accName("tf-acc-swap")
	attr := accName("tf-acc-swap-attr")
	resource.Test(t, providerTestCase([]resource.TestStep{{
		Config: accProviderBlock() + `
data "sigma_connections" "all" {}
resource "sigma_user_attribute" "test" {
  name          = "` + attr + `"
  default_value = "unset"
}
resource "sigma_source_swap_policy" "test" {
  type               = "deployment"
  name               = "` + name + `"
  from_connection_id = data.sigma_connections.all.connections[0].id
  swaps_json = jsonencode({
    toConnection = {
      swapType        = "attribute"
      userAttributeId = sigma_user_attribute.test.id
    }
    deploymentSwaps = []
  })
}
`,
		Check: resource.TestCheckResourceAttrSet("sigma_source_swap_policy.test", "id"),
	}}))
}

func runAccTranslationVariant(t *testing.T) {
	t.Helper()
	requireAcceptance(t)
	variant := accName("tf-acc")
	resource.Test(t, providerTestCase([]resource.TestStep{{
		Config: accProviderBlock() + `
resource "sigma_translation" "test" {
  lng         = "en"
  lng_variant = "` + variant + `"
  translations = {
    "tf-acc.hello" = "hello"
  }
}
`,
		Check: resource.TestCheckResourceAttrSet("sigma_translation.test", "id"),
	}}))
}
