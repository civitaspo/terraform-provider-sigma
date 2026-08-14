package provider_test

import (
	"testing"
)

func TestMembersDataSource(t *testing.T) {
	member := map[string]any{
		"memberId": "member-1", "organizationId": "org-1", "email": "ada@example.com",
		"firstName": "Ada", "lastName": "Lovelace", "memberType": "Creator", "userKind": "internal",
		"isArchived": false, "isInactive": false, "homeFolderId": "folder-1",
	}
	runListDataSourceCases(t, listDataSourceCase{
		path: "/v2/members",
		config: `
data "sigma_members" "all" {
  search           = "Ada"
  email            = "ada@example.com"
  include_archived = true
  include_inactive = false
}
`,
		wantQuery: map[string]string{
			"search":          "Ada",
			"email":           "ada@example.com",
			"includeArchived": "true",
			"includeInactive": "false",
		},
		entry:     member,
		address:   "data.sigma_members.all",
		countAttr: "members.#",
	})
}

func TestMembersDataSourceMapping(t *testing.T) {
	runListDataSourceCases(t, listDataSourceCase{
		path:      "/v2/members",
		config:    `data "sigma_members" "all" {}`,
		wantQuery: map[string]string{},
		entry: map[string]any{
			"memberId": "member-1", "organizationId": "org-1", "email": "ada@example.com",
			"firstName": "Ada", "lastName": "Lovelace", "memberType": "Creator", "userKind": "internal",
			"isArchived": false, "isInactive": false, "homeFolderId": "folder-1",
		},
		address:   "data.sigma_members.all",
		countAttr: "members.#",
	})
}

func TestTeamsDataSource(t *testing.T) {
	runListDataSourceCases(t, listDataSourceCase{
		path: "/v2.1/teams",
		config: `
data "sigma_teams" "all" {
  name        = "Analytics"
  description = "Core"
  visibility  = "private"
}
`,
		wantQuery: map[string]string{"name": "Analytics", "description": "Core", "visibility": "private"},
		entry: map[string]any{
			"teamId": "team-1", "name": "Analytics", "description": "Core", "visibility": "private", "isArchived": false, "workspaceId": "workspace-1",
		},
		address:   "data.sigma_teams.all",
		countAttr: "teams.#",
	})
}

func TestWorkspacesDataSource(t *testing.T) {
	runListDataSourceCases(t, listDataSourceCase{
		path: "/v2.1/workspaces",
		config: `
data "sigma_workspaces" "test" {
  name       = "Analytics"
  exact_name = "Analytics"
}
`,
		wantQuery: map[string]string{"name": "Analytics", "exactName": "Analytics"},
		entry: map[string]any{
			"workspaceId": "workspace-1", "workspaceUrlId": "url-1", "name": "Analytics",
			"createdBy": "member-1", "updatedBy": "member-1", "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
		},
		address:   "data.sigma_workspaces.test",
		countAttr: "workspaces.#",
	})
}

func TestAccountTypesDataSource(t *testing.T) {
	runListDataSourceCases(t, listDataSourceCase{
		path:       "/v2/accountTypes",
		cursorKey:  "pageToken",
		tokenPaged: true,
		config:     `data "sigma_account_types" "all" {}`,
		wantQuery:  map[string]string{},
		entry: map[string]any{
			"accountTypeId": "type-1", "accountTypeName": "Analyst", "description": "Analyst access", "isCustom": true,
		},
		address:   "data.sigma_account_types.all",
		countAttr: "account_types.#",
	})
}

func TestUserAttributesDataSource(t *testing.T) {
	runListDataSourceCases(t, listDataSourceCase{
		path: "/v2/user-attributes",
		config: `
data "sigma_user_attributes" "all" {
  name = "Region"
}
`,
		wantQuery: map[string]string{"name": "Region"},
		entry: map[string]any{
			"userAttributeId": "attribute-1", "name": "Region", "description": "Sales region",
			"defaultValue": map[string]string{"val": "global", "type": "string"},
		},
		address:   "data.sigma_user_attributes.all",
		countAttr: "user_attributes.#",
	})
}
