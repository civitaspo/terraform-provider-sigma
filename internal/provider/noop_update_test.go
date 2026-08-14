package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestNoopAndUnexpectedUpdates(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	req := resource.UpdateRequest{}
	resp := &resource.UpdateResponse{}

	(&workspaceGrantResource{}).Update(ctx, req, resp)
	(&workbookGrantResource{}).Update(ctx, req, resp)
	(&reportGrantResource{}).Update(ctx, req, resp)
	(&connectionGrantResource{}).Update(ctx, req, resp)
	(&connectionPathGrantResource{}).Update(ctx, req, resp)
	(&teamMemberResource{}).Update(ctx, req, resp)
	(&userAttributeResource{}).Update(ctx, req, resp)
	(&workbookEmbedResource{}).Update(ctx, req, resp)

	errorResp := &resource.UpdateResponse{}
	(&tenantDeploymentCapabilityResource{}).Update(ctx, req, errorResp)
	(&deploymentPolicyDocumentResource{}).Update(ctx, req, errorResp)
	(&deploymentPolicyTenantResource{}).Update(ctx, req, errorResp)
	if !errorResp.Diagnostics.HasError() {
		t.Fatal("expected unexpected-update diagnostics")
	}
}
