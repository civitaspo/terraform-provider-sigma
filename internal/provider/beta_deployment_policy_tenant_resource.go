package provider

import (
	"context"
	"fmt"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type deploymentPolicyTenantResource struct{ configuredResource }

var (
	_ resource.Resource                = (*deploymentPolicyTenantResource)(nil)
	_ resource.ResourceWithConfigure   = (*deploymentPolicyTenantResource)(nil)
	_ resource.ResourceWithImportState = (*deploymentPolicyTenantResource)(nil)
)

type deploymentPolicyTenantModel struct {
	ID                 types.String `tfsdk:"id"`
	DeploymentPolicyID types.String `tfsdk:"deployment_policy_id"`
	TenantID           types.String `tfsdk:"tenant_id"`
}

func NewDeploymentPolicyTenantResource() resource.Resource {
	return &deploymentPolicyTenantResource{}
}

func (r *deploymentPolicyTenantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment_policy_tenant"
}
func (r *deploymentPolicyTenantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *deploymentPolicyTenantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Attaches exactly one tenant organization to an existing Sigma deployment policy. This resource never creates the parent policy. " + betaAPINotice,
		Attributes: map[string]schema.Attribute{
			"id":                   schema.StringAttribute{Computed: true, MarkdownDescription: "Composite ID `deploymentPolicyId/tenantId`."},
			"deployment_policy_id": schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Existing deployment policy ID."},
			"tenant_id":            schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Tenant organization ID to attach."},
		},
	}
}
func (r *deploymentPolicyTenantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan deploymentPolicyTenantModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	policyID, policyDiags := knownString(plan.DeploymentPolicyID, "deployment_policy_id")
	tenantID, tenantDiags := knownString(plan.TenantID, "tenant_id")
	resp.Diagnostics.Append(policyDiags...)
	resp.Diagnostics.Append(tenantDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.AddDeploymentPolicyTenant(ctx, policyID, tenantID); err != nil {
		resp.Diagnostics.AddError("Unable to add Sigma deployment policy tenant", err.Error())
		return
	}
	plan.ID = types.StringValue(policyID + "/" + tenantID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *deploymentPolicyTenantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state deploymentPolicyTenantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	policyID, policyDiags := knownString(state.DeploymentPolicyID, "deployment_policy_id")
	tenantID, tenantDiags := knownString(state.TenantID, "tenant_id")
	resp.Diagnostics.Append(policyDiags...)
	resp.Diagnostics.Append(tenantDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenants, err := r.client.ListDeploymentPolicyTenants(ctx, policyID)
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma deployment policy tenants", err.Error())
		return
	}
	found := false
	for _, id := range tenants {
		if id == tenantID {
			found = true
			break
		}
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}
	state.ID = types.StringValue(policyID + "/" + tenantID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *deploymentPolicyTenantResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Unexpected update", "sigma_deployment_policy_tenant has no mutable attributes.")
}
func (r *deploymentPolicyTenantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state deploymentPolicyTenantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	policyID, policyDiags := knownString(state.DeploymentPolicyID, "deployment_policy_id")
	tenantID, tenantDiags := knownString(state.TenantID, "tenant_id")
	resp.Diagnostics.Append(policyDiags...)
	resp.Diagnostics.Append(tenantDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.RemoveDeploymentPolicyTenant(ctx, policyID, tenantID); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to remove Sigma deployment policy tenant", err.Error())
	}
}
func (r *deploymentPolicyTenantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	policyID, tenantID, ok := splitCompositeImportID(req.ID)
	if !ok {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected deploymentPolicyId/tenantId, got %q.", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("deployment_policy_id"), policyID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_id"), tenantID)...)
}
