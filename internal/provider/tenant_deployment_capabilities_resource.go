package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tenantDeploymentCapabilitiesResource struct{ configuredResource }

var (
	_ resource.Resource                = (*tenantDeploymentCapabilitiesResource)(nil)
	_ resource.ResourceWithConfigure   = (*tenantDeploymentCapabilitiesResource)(nil)
	_ resource.ResourceWithImportState = (*tenantDeploymentCapabilitiesResource)(nil)
)

type tenantDeploymentCapabilitiesModel struct {
	ID           types.String `tfsdk:"id"`
	TenantID     types.String `tfsdk:"tenant_id"`
	Capabilities types.Set    `tfsdk:"capabilities"`
}

func NewTenantDeploymentCapabilitiesResource() resource.Resource {
	return &tenantDeploymentCapabilitiesResource{}
}
func (r *tenantDeploymentCapabilitiesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_deployment_capabilities"
}
func (r *tenantDeploymentCapabilitiesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *tenantDeploymentCapabilitiesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Authoritatively manages which tenant organizations a Sigma tenant can deploy to. " + betaAPINotice,
		Attributes: map[string]schema.Attribute{
			"id":        schema.StringAttribute{Computed: true, MarkdownDescription: "Tenant organization ID."},
			"tenant_id": schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "Tenant organization that receives deployment capabilities."},
			"capabilities": schema.SetAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Authoritative set of tenant organization IDs this tenant can deploy to.",
			},
		},
	}
}
func (r *tenantDeploymentCapabilitiesResource) sync(ctx context.Context, plan *tenantDeploymentCapabilitiesModel, diagnostics interface{ AddError(string, string) }) bool {
	var desired []string
	plan.Capabilities.ElementsAs(ctx, &desired, false)
	current, err := r.client.ListTenantDeploymentCapabilities(ctx, plan.TenantID.ValueString())
	if err != nil {
		diagnostics.AddError("Unable to read Sigma tenant deployment capabilities", err.Error())
		return false
	}
	have := make([]string, 0, len(current))
	for _, cap := range current {
		have = append(have, cap.TenantOrganizationID)
	}
	add, remove := stringSetDiff(desired, have)
	if len(add) > 0 {
		if err := r.client.AddTenantDeploymentCapabilities(ctx, plan.TenantID.ValueString(), add); err != nil {
			diagnostics.AddError("Unable to add Sigma tenant deployment capabilities", err.Error())
			return false
		}
	}
	if len(remove) > 0 {
		if err := r.client.RemoveTenantDeploymentCapabilities(ctx, plan.TenantID.ValueString(), remove); err != nil {
			diagnostics.AddError("Unable to remove Sigma tenant deployment capabilities", err.Error())
			return false
		}
	}
	plan.ID = plan.TenantID
	return true
}
func (r *tenantDeploymentCapabilitiesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tenantDeploymentCapabilitiesModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if !resp.Diagnostics.HasError() && r.sync(ctx, &plan, &resp.Diagnostics) {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}
func (r *tenantDeploymentCapabilitiesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tenantDeploymentCapabilitiesModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	current, err := r.client.ListTenantDeploymentCapabilities(ctx, state.TenantID.ValueString())
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma tenant deployment capabilities", err.Error())
		return
	}
	ids := make([]string, 0, len(current))
	for _, cap := range current {
		ids = append(ids, cap.TenantOrganizationID)
	}
	state.Capabilities, _ = types.SetValueFrom(ctx, types.StringType, ids)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *tenantDeploymentCapabilitiesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tenantDeploymentCapabilitiesModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if !resp.Diagnostics.HasError() && r.sync(ctx, &plan, &resp.Diagnostics) {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}
func (r *tenantDeploymentCapabilitiesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tenantDeploymentCapabilitiesModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var ids []string
	state.Capabilities.ElementsAs(ctx, &ids, false)
	if len(ids) == 0 {
		return
	}
	if err := r.client.RemoveTenantDeploymentCapabilities(ctx, state.TenantID.ValueString(), ids); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to remove Sigma tenant deployment capabilities", err.Error())
	}
}
func (r *tenantDeploymentCapabilitiesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_id"), req.ID)...)
}
