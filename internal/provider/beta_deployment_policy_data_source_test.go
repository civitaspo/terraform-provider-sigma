package provider_test

import "testing"

func TestDeploymentPolicyDataSource(t *testing.T) {
	runSingularDataSourceCases(t, singularDataSourceCase{
		path: "/v2/deploymentPolicies/policy-1",
		config: `
data "sigma_deployment_policy" "one" {
  id = "policy-1"
}
`,
		entry: map[string]any{
			"deploymentPolicyId": "policy-1", "name": "Starter", "nameInTenant": "Starter",
			"versionTagId": "tag-1", "sourceSwapPolicies": []string{}, "copyInputTableData": true,
		},
		address:   "data.sigma_deployment_policy.one",
		checkAttr: "name",
		want:      "Starter",
	})
}

func TestAccDeploymentPolicyDataSource(t *testing.T) { requireAcceptance(t) }
