package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tenantResource struct{ configuredResource }

var (
	_ resource.Resource                = (*tenantResource)(nil)
	_ resource.ResourceWithConfigure   = (*tenantResource)(nil)
	_ resource.ResourceWithImportState = (*tenantResource)(nil)
)

type tenantModel struct {
	ID                     types.String `tfsdk:"id"`
	TenantOrganizationName types.String `tfsdk:"tenant_organization_name"`
	TenantOrganizationSlug types.String `tfsdk:"tenant_organization_slug"`
	CloudProvider          types.String `tfsdk:"cloud_provider"`
	ParentOrganizationID   types.String `tfsdk:"parent_organization_id"`
	CreatedBy              types.String `tfsdk:"created_by"`
	UpdatedBy              types.String `tfsdk:"updated_by"`
	CreatedAt              types.String `tfsdk:"created_at"`
	UpdatedAt              types.String `tfsdk:"updated_at"`
	SharedAt               types.String `tfsdk:"shared_at"`
	TenantCloudProvider    types.String `tfsdk:"tenant_cloud_provider"`
	TenantRegion           types.String `tfsdk:"tenant_region"`
	TenantAPIURL           types.String `tfsdk:"tenant_api_url"`
}

func NewTenantResource() resource.Resource { return &tenantResource{} }

func (r *tenantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant"
}
func (r *tenantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *tenantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Sigma tenant organization. " + betaAPINotice,
		Attributes: map[string]schema.Attribute{
			"id":                       schema.StringAttribute{Computed: true, MarkdownDescription: "Tenant organization ID."},
			"tenant_organization_name": schema.StringAttribute{Required: true, MarkdownDescription: "Display name of the tenant organization."},
			"tenant_organization_slug": schema.StringAttribute{Required: true, MarkdownDescription: "URL identifier for the tenant organization."},
			"cloud_provider": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "Cloud provider for the tenant organization. Defaults to the parent organization's provider. " +
					"Changing this forces a new resource.",
			},
			"parent_organization_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Parent organization ID."},
			"created_by":             schema.StringAttribute{Computed: true, MarkdownDescription: "Member ID that created the tenant."},
			"updated_by":             schema.StringAttribute{Computed: true, MarkdownDescription: "Member ID that last updated the tenant."},
			"created_at":             schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp."},
			"updated_at":             schema.StringAttribute{Computed: true, MarkdownDescription: "Last update timestamp."},
			"shared_at":              schema.StringAttribute{Computed: true, MarkdownDescription: "Share timestamp, when applicable."},
			"tenant_cloud_provider":  schema.StringAttribute{Computed: true, MarkdownDescription: "Cloud provider hosting the tenant."},
			"tenant_region":          schema.StringAttribute{Computed: true, MarkdownDescription: "Region hosting the tenant."},
			"tenant_api_url":         schema.StringAttribute{Computed: true, MarkdownDescription: "Tenant organization API base URL."},
		},
	}
}
func setTenant(state *tenantModel, value *sigma.Tenant) {
	state.ID = types.StringValue(value.TenantOrganizationID)
	state.TenantOrganizationName = types.StringValue(value.TenantOrganizationName)
	state.TenantOrganizationSlug = types.StringValue(value.TenantOrganizationSlug)
	state.ParentOrganizationID = types.StringValue(value.ParentOrganizationID)
	state.CreatedBy = types.StringValue(value.CreatedBy)
	state.UpdatedBy = types.StringValue(value.UpdatedBy)
	state.CreatedAt = types.StringValue(value.CreatedAt)
	state.UpdatedAt = types.StringValue(value.UpdatedAt)
	state.SharedAt = stringOrNull(value.SharedAt)
	state.TenantCloudProvider = stringOrNull(value.TenantCloudProvider)
	state.TenantRegion = stringOrNull(value.TenantRegion)
	state.TenantAPIURL = stringOrNull(value.TenantAPIURL)
}
func (r *tenantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tenantModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.CreateTenant(ctx, sigma.CreateTenantInput{
		TenantOrganizationName: plan.TenantOrganizationName.ValueString(),
		TenantOrganizationSlug: plan.TenantOrganizationSlug.ValueString(),
		CloudProvider:          optionalStringPtr(plan.CloudProvider),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma tenant", err.Error())
		return
	}
	setTenant(&plan, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *tenantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tenantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.GetTenant(ctx, state.ID.ValueString())
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma tenant", err.Error())
		return
	}
	setTenant(&state, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *tenantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tenantModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	name := plan.TenantOrganizationName.ValueString()
	slug := plan.TenantOrganizationSlug.ValueString()
	value, err := r.client.PatchTenant(ctx, plan.ID.ValueString(), sigma.PatchTenantInput{
		TenantOrganizationName: &name,
		TenantOrganizationSlug: &slug,
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Sigma tenant", err.Error())
		return
	}
	setTenant(&plan, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *tenantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tenantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteTenant(ctx, state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma tenant", err.Error())
	}
}
func (r *tenantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importPassthrough(ctx, req, resp)
}
