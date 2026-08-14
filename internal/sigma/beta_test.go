package sigma_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
)

func TestBetaClientContract(t *testing.T) {
	t.Parallel()
	tenant := map[string]any{
		"tenantOrganizationId": "tenant-1", "parentOrganizationId": "org-1",
		"createdBy": "m1", "updatedBy": "m1", "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
		"tenantOrganizationName": "Acme", "tenantOrganizationSlug": "acme",
	}
	policy := map[string]any{
		"deploymentPolicyId": "policy-1", "name": "Starter", "nameInTenant": "Starter",
		"sourceSwapPolicies": []any{}, "copyInputTableData": false,
	}
	swap := map[string]any{
		"policyId": "swap-1", "type": "deployment", "name": "Swap", "fromConnectionId": "conn-1",
		"swaps": map[string]any{"deploymentSwaps": []any{}},
	}
	page := func(entries any) map[string]any {
		return map[string]any{"entries": entries, "nextPageToken": nil}
	}
	mock := testutil.NewRecordingSigma(t,
		testutil.ExpectedRequest{Method: "POST", Path: "/v2/tenants", JSONBody: map[string]any{"tenantOrganizationName": "Acme", "tenantOrganizationSlug": "acme"}, Response: tenant},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/tenants/tenant-1", Response: tenant},
		testutil.ExpectedRequest{Method: "PATCH", Path: "/v2/tenants/tenant-1", JSONBody: map[string]any{"tenantOrganizationName": "Acme 2"}, Response: tenant},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/tenants", Response: page([]any{tenant})},
		testutil.ExpectedRequest{Method: "DELETE", Path: "/v2/tenants/tenant-1", Response: map[string]any{}},
		testutil.ExpectedRequest{Method: "POST", Path: "/v2/tenants/tenant-1/capabilities/deployments:batchAdd", JSONBody: map[string]any{"deployToTenantOrganizationIds": []any{"tenant-2"}}, Response: map[string]any{}},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/tenants/tenant-1/capabilities/deployments", Response: page([]any{map[string]any{"tenantOrganizationId": "tenant-2"}})},
		testutil.ExpectedRequest{Method: "POST", Path: "/v2/tenants/tenant-1/capabilities/deployments:batchRemove", JSONBody: map[string]any{"deployToTenantOrganizationIds": []any{"tenant-2"}}, Response: map[string]any{}},
		testutil.ExpectedRequest{Method: "POST", Path: "/v2/deploymentPolicies", JSONBody: map[string]any{"name": "Starter"}, Response: map[string]any{"deploymentPolicyId": "policy-1"}},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/deploymentPolicies/policy-1", Response: policy},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/deploymentPolicies/policy-1", Response: policy},
		testutil.ExpectedRequest{Method: "PATCH", Path: "/v2/deploymentPolicies/policy-1", JSONBody: map[string]any{"name": "Starter 2"}, Response: policy},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/deploymentPolicies", Response: page([]any{policy})},
		testutil.ExpectedRequest{Method: "POST", Path: "/v2/deploymentPolicies/policy-1/files", JSONBody: map[string]any{"inodeIds": []any{"inode-1"}}, Response: map[string]any{}},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/deploymentPolicies/policy-1/files", Response: page([]any{map[string]any{"inodeId": "inode-1", "deploymentPolicyId": "policy-1"}})},
		testutil.ExpectedRequest{Method: "DELETE", Path: "/v2/deploymentPolicies/policy-1/files/inode-1", Response: map[string]any{}},
		testutil.ExpectedRequest{Method: "POST", Path: "/v2/deploymentPolicies/policy-1/tenants", JSONBody: map[string]any{"tenantOrganizationId": "tenant-1"}, Response: map[string]any{}},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/deploymentPolicies/policy-1/tenants", Response: page([]any{"tenant-1"})},
		testutil.ExpectedRequest{Method: "DELETE", Path: "/v2/deploymentPolicies/policy-1/tenants/tenant-1", Response: map[string]any{}},
		testutil.ExpectedRequest{Method: "DELETE", Path: "/v2/deploymentPolicies/policy-1", Response: map[string]any{}},
		testutil.ExpectedRequest{Method: "POST", Path: "/v2/sourceSwapPolicies", JSONBody: map[string]any{"type": "deployment", "name": "Swap", "fromConnectionId": "conn-1", "swaps": map[string]any{"deploymentSwaps": []any{}}}, Response: map[string]any{"policyId": "swap-1"}},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/sourceSwapPolicies/swap-1", Response: swap},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/sourceSwapPolicies/swap-1", Response: swap},
		testutil.ExpectedRequest{Method: "PATCH", Path: "/v2/sourceSwapPolicies/swap-1", JSONBody: map[string]any{"type": "deployment", "name": "Swap 2", "fromConnectionId": "conn-1", "swaps": map[string]any{"deploymentSwaps": []any{}}}, Response: map[string]any{"policyId": "swap-1"}},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/sourceSwapPolicies/swap-1", Response: swap},
		testutil.ExpectedRequest{Method: "DELETE", Path: "/v2/sourceSwapPolicies/swap-1", Response: map[string]any{}},
	)
	client, err := sigma.NewClient(mock.URL(), mock.ClientID, mock.ClientSecret)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := client.CreateTenant(ctx, sigma.CreateTenantInput{TenantOrganizationName: "Acme", TenantOrganizationSlug: "acme"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetTenant(ctx, "tenant-1"); err != nil {
		t.Fatal(err)
	}
	name := "Acme 2"
	if _, err := client.PatchTenant(ctx, "tenant-1", sigma.PatchTenantInput{TenantOrganizationName: &name}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListTenants(ctx, sigma.ListTenantsOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteTenant(ctx, "tenant-1"); err != nil {
		t.Fatal(err)
	}
	if err := client.AddTenantDeploymentCapabilities(ctx, "tenant-1", []string{"tenant-2"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListTenantDeploymentCapabilities(ctx, "tenant-1"); err != nil {
		t.Fatal(err)
	}
	if err := client.RemoveTenantDeploymentCapabilities(ctx, "tenant-1", []string{"tenant-2"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateDeploymentPolicy(ctx, sigma.CreateDeploymentPolicyInput{Name: "Starter"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetDeploymentPolicy(ctx, "policy-1"); err != nil {
		t.Fatal(err)
	}
	policyName := "Starter 2"
	if _, err := client.UpdateDeploymentPolicy(ctx, "policy-1", sigma.UpdateDeploymentPolicyInput{Name: &policyName}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListDeploymentPolicies(ctx); err != nil {
		t.Fatal(err)
	}
	if err := client.AddDeploymentPolicyInodes(ctx, "policy-1", []string{"inode-1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListDeploymentPolicyInodes(ctx, "policy-1"); err != nil {
		t.Fatal(err)
	}
	if err := client.RemoveDeploymentPolicyInode(ctx, "policy-1", "inode-1"); err != nil {
		t.Fatal(err)
	}
	if err := client.AddDeploymentPolicyTenant(ctx, "policy-1", "tenant-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListDeploymentPolicyTenants(ctx, "policy-1"); err != nil {
		t.Fatal(err)
	}
	if err := client.RemoveDeploymentPolicyTenant(ctx, "policy-1", "tenant-1"); err != nil {
		t.Fatal(err)
	}
	if err := client.ArchiveDeploymentPolicy(ctx, "policy-1"); err != nil {
		t.Fatal(err)
	}
	swaps, _ := json.Marshal(map[string]any{"deploymentSwaps": []any{}})
	if _, err := client.CreateSourceSwapPolicy(ctx, sigma.SourceSwapPolicyInput{Type: "deployment", Name: "Swap", FromConnectionID: "conn-1", Swaps: swaps}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetSourceSwapPolicy(ctx, "swap-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.UpdateSourceSwapPolicy(ctx, "swap-1", sigma.SourceSwapPolicyInput{Type: "deployment", Name: "Swap 2", FromConnectionID: "conn-1", Swaps: swaps}); err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteSourceSwapPolicy(ctx, "swap-1"); err != nil {
		t.Fatal(err)
	}
}
