package provider_test

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestWorkbooksDataSource(t *testing.T) {
	runListDataSourceCases(t, listDataSourceCase{
		path: "/v2/workbooks",
		config: `
data "sigma_workbooks" "all" {
  exclude_tags          = true
  skip_permission_check = false
  is_archived           = false
  exclude_explorations  = true
}
`,
		wantQuery: map[string]string{
			"excludeTags":         "true",
			"skipPermissionCheck": "false",
			"isArchived":          "false",
			"excludeExplorations": "true",
		},
		entry: map[string]any{
			"workbookId": "workbook-1", "workbookUrlId": "wb-url", "name": "Revenue", "url": "/workbook/wb-url",
			"path": "Analytics/Revenue", "latestVersion": 3, "ownerId": "member-1",
			"createdBy": "member-1", "updatedBy": "member-1", "description": "Revenue workbook",
			"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-02T00:00:00Z", "isArchived": false,
			"tags": []any{},
		},
		address:   "data.sigma_workbooks.all",
		countAttr: "workbooks.#",
	})
}

func TestReportsDataSource(t *testing.T) {
	runListDataSourceCases(t, listDataSourceCase{
		path: "/v2/reports",
		config: `
data "sigma_reports" "all" {
  exclude_tags          = false
  skip_permission_check = true
  is_archived           = true
}
`,
		wantQuery: map[string]string{"excludeTags": "false", "skipPermissionCheck": "true", "isArchived": "true"},
		entry: map[string]any{
			"reportId": "report-1", "reportUrlId": "rp-url", "name": "Weekly", "url": "/report/rp-url",
			"path": "Analytics/Weekly", "latestVersion": 2, "ownerId": "member-1",
			"createdBy": "member-1", "updatedBy": "member-1",
			"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-02T00:00:00Z", "isArchived": true,
		},
		address:   "data.sigma_reports.all",
		countAttr: "reports.#",
	})
}

func TestDataModelsDataSource(t *testing.T) {
	runListDataSourceCases(t, listDataSourceCase{
		path: "/v2/dataModels",
		config: `
data "sigma_data_models" "all" {
  exclude_tags          = true
  skip_permission_check = false
}
`,
		wantQuery: map[string]string{"excludeTags": "true", "skipPermissionCheck": "false"},
		entry: map[string]any{
			"dataModelId": "data-model-1", "dataModelUrlId": "dm-url", "name": "Sales Model", "url": "/data-model/dm-url",
			"path": "Analytics/Sales Model", "latestVersion": 4, "ownerId": "member-1",
			"createdBy": "member-1", "updatedBy": "member-1",
			"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-02T00:00:00Z", "isArchived": false,
			"tags": []any{},
		},
		address:   "data.sigma_data_models.all",
		countAttr: "data_models.#",
	})
}

func TestDatasetsDataSource(t *testing.T) {
	runListDataSourceCases(t, listDataSourceCase{
		path: "/v2/datasets",
		config: `
data "sigma_datasets" "all" {
  skip_permission_check = true
}
`,
		wantQuery: map[string]string{"skipPermissionCheck": "true"},
		entry: map[string]any{
			"datasetId": "dataset-1", "name": "Orders", "description": "Order facts", "url": "/dataset/dataset-1",
			"path": "Analytics/Orders", "owner": "member-1", "referenceCount": 2, "migrationStatus": "not-migrated",
			"createdBy": "member-1", "updatedBy": "member-1",
			"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-02T00:00:00Z", "isArchived": false,
		},
		address:   "data.sigma_datasets.all",
		countAttr: "datasets.#",
	})
}

func TestTemplatesDataSource(t *testing.T) {
	runListDataSourceCases(t, listDataSourceCase{
		path: "/v2/templates",
		config: `
data "sigma_templates" "all" {
  source = "internal"
  search = "KPI"
}
`,
		wantQuery: map[string]string{"source": "internal", "search": "KPI"},
		entry: map[string]any{
			"templateId": "template-1", "templateUrlId": "tpl-url", "name": "KPI", "url": "/template/tpl-url",
			"path": "Analytics/KPI", "latestVersion": 1,
			"createdBy": "member-1", "updatedBy": "member-1",
			"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-02T00:00:00Z", "isArchived": false,
			"tags": []any{},
		},
		address:   "data.sigma_templates.all",
		countAttr: "templates.#",
	})
}

func TestTagsDataSource(t *testing.T) {
	runListDataSourceCases(t, listDataSourceCase{
		path: "/v2/tags",
		config: `
data "sigma_tags" "all" {
  search = "prod"
}
`,
		wantQuery: map[string]string{"search": "prod"},
		entry: map[string]any{
			"versionTagId": "tag-1", "name": "prod", "color": "#00ff00", "description": "Production",
			"ownerId": "member-1", "createdBy": "member-1", "updatedBy": "member-1",
			"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-02T00:00:00Z", "isArchived": false,
		},
		address:   "data.sigma_tags.all",
		countAttr: "tags.#",
	})
}

func TestConnectionsDataSource(t *testing.T) {
	runListDataSourceCases(t, listDataSourceCase{
		path: "/v2/connections",
		config: `
data "sigma_connections" "all" {
  search           = "warehouse"
  include_archived = false
}
`,
		wantQuery: map[string]string{"search": "warehouse", "includeArchived": "false"},
		entry: map[string]any{
			"connectionId": "connection-1", "name": "warehouse", "type": "snowflake",
			"description": map[string]any{}, "poolSizes": map[string]any{}, "friendlyName": false, "useOauth": true,
			"organizationId": "org-1", "isSample": false, "isAuditLog": false, "lastActiveAt": "2026-01-01T00:00:00Z",
			"createdBy": "member-1", "updatedBy": "member-1", "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
		},
		address:   "data.sigma_connections.all",
		countAttr: "connections.#",
	})
}

func TestConnectionPathsDataSource(t *testing.T) {
	runListDataSourceCases(t, listDataSourceCase{
		path: "/v2/connections/paths",
		config: `
data "sigma_connection_paths" "all" {
  connection_id = "connection-1"
}
`,
		wantQuery: map[string]string{"connectionId": "connection-1"},
		entry:     map[string]any{"connectionId": "connection-1", "urlId": "path-1", "path": []string{"DATABASE", "SCHEMA"}},
		address:   "data.sigma_connection_paths.all",
		countAttr: "paths.#",
	})
}

func TestConnectionPathDataSource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	mock.Mux.HandleFunc("/v2/connection/connection-1/lookup", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			t.Errorf("method = %s", request.Method)
		}
		writeJSON(response, map[string]any{"kind": "table", "inodeId": "path-1", "url": "/connection/path-1"})
	})
	resource.UnitTest(t, connectionTestCase([]resource.TestStep{{
		Config: connectionProviderConfig(mock) + `
data "sigma_connection_path" "lookup" {
  connection_id = "connection-1"
  path          = ["DATABASE", "SCHEMA", "TABLE"]
}
`,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.sigma_connection_path.lookup", "id", "path-1"),
			resource.TestCheckResourceAttr("data.sigma_connection_path.lookup", "inode_id", "path-1"),
			resource.TestCheckResourceAttr("data.sigma_connection_path.lookup", "kind", "table"),
		),
	}}))
}

func TestConnectionPathDataSourceEmptyPath(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	resource.UnitTest(t, connectionTestCase([]resource.TestStep{{
		Config: connectionProviderConfig(mock) + `
data "sigma_connection_path" "lookup" {
  connection_id = "connection-1"
  path          = []
}
`,
		ExpectError: regexp.MustCompile("at least one component"),
	}}))
}

func TestTenantsDataSource(t *testing.T) {
	runListDataSourceCases(t, listDataSourceCase{
		path:       "/v2/tenants",
		cursorKey:  "pageToken",
		tokenPaged: true,
		config: `
data "sigma_tenants" "test" {
  search = "Acme"
  key    = "name"
  order  = "asc"
}
`,
		wantQuery: map[string]string{"search": "Acme", "key": "name", "order": "asc"},
		entry: map[string]any{
			"tenantOrganizationId": "tenant-1", "parentOrganizationId": "parent-1",
			"createdBy": "user-1", "updatedBy": "user-1",
			"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
			"tenantOrganizationName": "Acme", "tenantOrganizationSlug": "acme",
		},
		address:   "data.sigma_tenants.test",
		countAttr: "tenants.#",
	})
}

func TestDeploymentPoliciesDataSource(t *testing.T) {
	runListDataSourceCases(t, listDataSourceCase{
		path:       "/v2/deploymentPolicies",
		cursorKey:  "pageToken",
		tokenPaged: true,
		config:     `data "sigma_deployment_policies" "test" {}`,
		wantQuery:  map[string]string{},
		entry: map[string]any{
			"deploymentPolicyId": "policy-1", "name": "Starter", "nameInTenant": "Starter",
			"versionTagId": "tag-1", "sourceSwapPolicies": []string{}, "copyInputTableData": true,
		},
		address:   "data.sigma_deployment_policies.test",
		countAttr: "deployment_policies.#",
	})
}
