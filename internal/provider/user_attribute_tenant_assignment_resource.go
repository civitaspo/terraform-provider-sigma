package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*userAttributeTenantAssignmentResource)(nil)
	_ resource.ResourceWithConfigure   = (*userAttributeTenantAssignmentResource)(nil)
	_ resource.ResourceWithImportState = (*userAttributeTenantAssignmentResource)(nil)
)

type userAttributeTenantAssignmentResource struct{ configuredResource }

func NewUserAttributeTenantAssignmentResource() resource.Resource {
	return &userAttributeTenantAssignmentResource{}
}

func (r *userAttributeTenantAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_attribute_tenant_assignment"
}

func (r *userAttributeTenantAssignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}

func (r *userAttributeTenantAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = userAttributeAssignmentSchema(
		"tenant",
		"Tenant organization ID (`tenantOrganizationId`).",
		"Manages a Sigma user attribute assignment for one tenant. "+betaAPINotice,
	)
}

func (r *userAttributeTenantAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := assignmentFromPlan(ctx, req.Plan.GetAttribute, &resp.Diagnostics, "tenant_id")
	if !resp.Diagnostics.HasError() && setUserAttributeTenant(ctx, r.client, plan, &resp.Diagnostics) {
		setAssignmentState(ctx, plan, resp.State.SetAttribute, &resp.Diagnostics, "tenant_id")
	}
}

func (r *userAttributeTenantAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readUserAttributeAssignment(ctx, req, resp, "tenant_id", func(attributeID string) ([]sigma.AttributeAssignment, error) {
		return r.client.ListUserAttributeTenants(ctx, attributeID)
	}, func(assignment sigma.AttributeAssignment) string {
		return assignment.TenantOrganizationID
	})
}

func (r *userAttributeTenantAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := assignmentFromPlan(ctx, req.Plan.GetAttribute, &resp.Diagnostics, "tenant_id")
	if !resp.Diagnostics.HasError() && setUserAttributeTenant(ctx, r.client, plan, &resp.Diagnostics) {
		setAssignmentState(ctx, plan, resp.State.SetAttribute, &resp.Diagnostics, "tenant_id")
	}
}

func (r *userAttributeTenantAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var attributeID, targetID types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("user_attribute_id"), &attributeID)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("tenant_id"), &targetID)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteUserAttributeTenant(ctx, attributeID.ValueString(), targetID.ValueString()); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma user attribute assignment", err.Error())
	}
}

func (r *userAttributeTenantAssignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importUserAttributeAssignment(ctx, req, resp, "tenant_id", "tenantOrganizationId")
}

func setUserAttributeTenant(ctx context.Context, client *sigma.Client, plan *assignmentModel, diagnostics interface{ AddError(string, string) }) bool {
	if err := client.SetUserAttributeTenant(ctx, plan.UserAttributeID.ValueString(), plan.TargetID.ValueString(), plan.Value.ValueString()); err != nil {
		diagnostics.AddError("Unable to set Sigma user attribute assignment", err.Error())
		return false
	}
	return true
}
