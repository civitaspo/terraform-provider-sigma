package provider_test

import "testing"

func TestFileDataSource(t *testing.T) {
	runSingularDataSourceCases(t, singularDataSourceCase{
		path: "/v2/files/file-1",
		config: `
data "sigma_file" "one" {
  id = "file-1"
}
`,
		entry: map[string]any{
			"id": "file-1", "urlId": "file-url-1", "name": "Managed", "type": "folder",
			"parentId": "workspace-1", "parentUrlId": "url-1", "permission": "edit", "path": "Analytics/Managed",
			"badge": nil, "isArchived": false, "description": "", "ownerId": nil,
			"createdBy": "member-1", "updatedBy": "member-1", "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
		},
		address:   "data.sigma_file.one",
		checkAttr: "name",
		want:      "Managed",
	})
}

func TestAccFileDataSource(t *testing.T) { requireAcceptance(t) }
