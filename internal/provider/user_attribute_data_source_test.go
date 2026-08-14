package provider_test

import "testing"

func TestUserAttributeDataSource(t *testing.T) {
	runSingularDataSourceCases(t, singularDataSourceCase{
		path: "/v2/user-attributes/attribute-1",
		config: `
data "sigma_user_attribute" "one" {
  id = "attribute-1"
}
`,
		entry: map[string]any{
			"userAttributeId": "attribute-1", "name": "Region", "description": "Sales region",
			"defaultValue": map[string]string{"val": "global", "type": "string"},
			"createdBy":    "member-1", "updatedBy": "member-1",
			"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-02T00:00:00Z",
		},
		address:   "data.sigma_user_attribute.one",
		checkAttr: "name",
		want:      "Region",
	})
}

func TestAccUserAttributeDataSource(t *testing.T) { requireAcceptance(t) }
