package provider_test

import "testing"

func TestTemplateDataSource(t *testing.T) {
	runSingularDataSourceCases(t, singularDataSourceCase{
		path: "/v2/templates/template-1",
		config: `
data "sigma_template" "one" {
  id = "template-1"
}
`,
		entry: map[string]any{
			"templateId": "template-1", "templateUrlId": "tpl-url", "name": "KPI", "url": "/template/tpl-url",
			"path": "Analytics/KPI", "latestVersion": 1,
			"createdBy": "member-1", "updatedBy": "member-1",
			"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-02T00:00:00Z", "isArchived": false,
			"tags": []any{map[string]any{"versionTagId": "tag-1", "name": "prod"}},
		},
		address:   "data.sigma_template.one",
		checkAttr: "name",
		want:      "KPI",
	})
}

func TestAccTemplateDataSource(t *testing.T) { requireAcceptance(t) }
