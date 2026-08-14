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

type tenantDeploymentCapabilityResource struct{ configuredResource }

var (
	_ resource.Resource                = (*tenantDeploymentCapabilityResource)(nil)
	_ resource.ResourceWithConfigure   = (*tenantDeploymentCapabilityResource)(nil)
	_ resource.ResourceWithImportState = (*tenantDeploymentCapabilityResource)(nil)
)

type tenantDeploymentCapabilityModel struct {
	ID         types.String `tfsdk:"id"`
	TenantID   types.String `tfsdk:"tenant_id"`
	Capability types.String `tfsdk:"capability"`
}

func NewTenantDeploymentCapabilityResource() resource.Resource {
	return &tenantDeploymentCapabilityResource{}
}

func (r *tenantDeploymentCapabilityResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_deployment_capability"
}
func (r *tenantDeploymentCapabilityResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *tenantDeploymentCapabilityResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Adds exactly one deploy-to tenant capability for a Sigma tenant organization. " + betaAPINotice,
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Composite ID `tenantId/capability`."},
			"tenant_id":  schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Tenant organization that receives the capability."},
			"capability": schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Tenant organization ID this tenant can deploy to."},
		},
	}
}
func (r *tenantDeploymentCapabilityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tenantDeploymentCapabilityModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, tenantDiags := knownString(plan.TenantID, "tenant_id")
	capability, capDiags := knownString(plan.Capability, "capability")
	resp.Diagnostics.Append(tenantDiags...)
	resp.Diagnostics.Append(capDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.AddTenantDeploymentCapabilities(ctx, tenantID, []string{capability}); err != nil {
		resp.Diagnostics.AddError("Unable to add Sigma tenant deployment capability", err.Error())
		return
	}
	plan.ID = types.StringValue(tenantID + "/" + capability)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *tenantDeploymentCapabilityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tenantDeploymentCapabilityModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, tenantDiags := knownString(state.TenantID, "tenant_id")
	capability, capDiags := knownString(state.Capability, "capability")
	resp.Diagnostics.Append(tenantDiags...)
	resp.Diagnostics.Append(capDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	current, err := r.client.ListTenantDeploymentCapabilities(ctx, tenantID)
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma tenant deployment capabilities", err.Error())
		return
	}
	found := false
	for _, item := range current {
		if item.TenantOrganizationID == capability {
			found = true
			break
		}
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}
	state.ID = types.StringValue(tenantID + "/" + capability)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *tenantDeploymentCapabilityResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Unexpected update", "sigma_tenant_deployment_capability has no mutable attributes.")
}
func (r *tenantDeploymentCapabilityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tenantDeploymentCapabilityModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenantID, tenantDiags := knownString(state.TenantID, "tenant_id")
	capability, capDiags := knownString(state.Capability, "capability")
	resp.Diagnostics.Append(tenantDiags...)
	resp.Diagnostics.Append(capDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.RemoveTenantDeploymentCapabilities(ctx, tenantID, []string{capability}); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to remove Sigma tenant deployment capability", err.Error())
	}
}
func (r *tenantDeploymentCapabilityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tenantID, capability, ok := splitCompositeImportID(req.ID)
	if !ok {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected tenantId/capability, got %q.", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_id"), tenantID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("capability"), capability)...)
}
