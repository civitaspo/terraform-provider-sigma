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

const betaAPINotice = "This resource uses a Sigma Beta API and may change without notice."

func stringSetDiff(desired, current []string) (add, remove []string) {
	have, want := map[string]bool{}, map[string]bool{}
	for _, id := range current {
		have[id] = true
	}
	for _, id := range desired {
		want[id] = true
	}
	for id := range want {
		if !have[id] {
			add = append(add, id)
		}
	}
	for id := range have {
		if !want[id] {
			remove = append(remove, id)
		}
	}
	return add, remove
}

func optionalStringPtr(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueString()
	return &v
}

func optionalBoolPtr(value types.Bool) *bool {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueBool()
	return &v
}

func stringOrNull(value *string) types.String {
	if value == nil || *value == "" {
		return types.StringNull()
	}
	return types.StringValue(*value)
}

type tenantResource struct{ configuredResource }

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

type tenantDeploymentCapabilitiesResource struct{ configuredResource }

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

type deploymentPolicyResource struct{ configuredResource }

type deploymentPolicyModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	NameInTenant       types.String `tfsdk:"name_in_tenant"`
	VersionTagID       types.String `tfsdk:"version_tag_id"`
	SourceSwapPolicies types.Set    `tfsdk:"source_swap_policies"`
	CopyInputTableData types.Bool   `tfsdk:"copy_input_table_data"`
	InodeIDs           types.Set    `tfsdk:"inode_ids"`
	TenantIDs          types.Set    `tfsdk:"tenant_ids"`
}

func NewDeploymentPolicyResource() resource.Resource { return &deploymentPolicyResource{} }

func (r *deploymentPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment_policy"
}
func (r *deploymentPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *deploymentPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Sigma deployment policy. Destroy archives the policy. When set, `inode_ids` and `tenant_ids` are authoritative attachments (`[]` removes all). Omit either attribute to leave that attachment set unmanaged. " + betaAPINotice,
		Attributes: map[string]schema.Attribute{
			"id":   schema.StringAttribute{Computed: true, MarkdownDescription: "Deployment policy ID."},
			"name": schema.StringAttribute{Required: true, MarkdownDescription: "Deployment policy name."},
			"name_in_tenant": schema.StringAttribute{
				Optional: true, Computed: true,
				MarkdownDescription: "Workspace name created in receiving tenants. Defaults to the policy name when omitted.",
			},
			"version_tag_id": schema.StringAttribute{
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "Version tag ID. Cannot be changed after create; changing this forces a new resource.",
			},
			"source_swap_policies": schema.SetAttribute{
				Optional: true, Computed: true, ElementType: types.StringType,
				MarkdownDescription: "Source swap policy IDs used when deploying documents.",
			},
			"copy_input_table_data": schema.BoolAttribute{
				Optional: true, Computed: true,
				MarkdownDescription: "Whether to copy editable draft input table data when deploying.",
			},
			"inode_ids": schema.SetAttribute{
				Optional: true, ElementType: types.StringType,
				MarkdownDescription: "Authoritative set of document inode IDs attached to the policy. Omit to leave document attachments unmanaged; set to `[]` to remove all.",
			},
			"tenant_ids": schema.SetAttribute{
				Optional: true, ElementType: types.StringType,
				MarkdownDescription: "Authoritative set of tenant organization IDs attached to the policy. Omit to leave tenant attachments unmanaged; set to `[]` to remove all.",
			},
		},
	}
}
func setDeploymentPolicy(ctx context.Context, state *deploymentPolicyModel, value *sigma.DeploymentPolicy) {
	state.ID = types.StringValue(value.DeploymentPolicyID)
	state.Name = types.StringValue(value.Name)
	state.NameInTenant = types.StringValue(value.NameInTenant)
	state.VersionTagID = stringOrNull(value.VersionTagID)
	state.CopyInputTableData = types.BoolValue(value.CopyInputTableData)
	if value.SourceSwapPolicies == nil {
		value.SourceSwapPolicies = []string{}
	}
	state.SourceSwapPolicies, _ = types.SetValueFrom(ctx, types.StringType, value.SourceSwapPolicies)
}
func (r *deploymentPolicyResource) syncAttachments(ctx context.Context, id string, plan *deploymentPolicyModel, diagnostics interface{ AddError(string, string) }) bool {
	if !plan.InodeIDs.IsNull() && !plan.InodeIDs.IsUnknown() {
		desiredInodes := []string{}
		plan.InodeIDs.ElementsAs(ctx, &desiredInodes, false)
		currentFiles, err := r.client.ListDeploymentPolicyInodes(ctx, id)
		if err != nil {
			diagnostics.AddError("Unable to list Sigma deployment policy documents", err.Error())
			return false
		}
		haveInodes := make([]string, 0, len(currentFiles))
		for _, file := range currentFiles {
			haveInodes = append(haveInodes, file.InodeID)
		}
		addInodes, removeInodes := stringSetDiff(desiredInodes, haveInodes)
		if len(addInodes) > 0 {
			if err := r.client.AddDeploymentPolicyInodes(ctx, id, addInodes); err != nil {
				diagnostics.AddError("Unable to add Sigma deployment policy documents", err.Error())
				return false
			}
		}
		for _, inodeID := range removeInodes {
			if err := r.client.RemoveDeploymentPolicyInode(ctx, id, inodeID); err != nil {
				diagnostics.AddError("Unable to remove Sigma deployment policy document", err.Error())
				return false
			}
		}
		plan.InodeIDs, _ = types.SetValueFrom(ctx, types.StringType, desiredInodes)
	}

	if !plan.TenantIDs.IsNull() && !plan.TenantIDs.IsUnknown() {
		desiredTenants := []string{}
		plan.TenantIDs.ElementsAs(ctx, &desiredTenants, false)
		haveTenants, err := r.client.ListDeploymentPolicyTenants(ctx, id)
		if err != nil {
			diagnostics.AddError("Unable to list Sigma deployment policy tenants", err.Error())
			return false
		}
		addTenants, removeTenants := stringSetDiff(desiredTenants, haveTenants)
		for _, tenantID := range addTenants {
			if err := r.client.AddDeploymentPolicyTenant(ctx, id, tenantID); err != nil {
				diagnostics.AddError("Unable to add Sigma deployment policy tenant", err.Error())
				return false
			}
		}
		for _, tenantID := range removeTenants {
			if err := r.client.RemoveDeploymentPolicyTenant(ctx, id, tenantID); err != nil {
				diagnostics.AddError("Unable to remove Sigma deployment policy tenant", err.Error())
				return false
			}
		}
		plan.TenantIDs, _ = types.SetValueFrom(ctx, types.StringType, desiredTenants)
	}
	return true
}
func (r *deploymentPolicyResource) readAttachments(ctx context.Context, state *deploymentPolicyModel) error {
	if !state.InodeIDs.IsNull() {
		files, err := r.client.ListDeploymentPolicyInodes(ctx, state.ID.ValueString())
		if err != nil {
			return fmt.Errorf("list documents: %w", err)
		}
		inodes := make([]string, 0, len(files))
		for _, file := range files {
			inodes = append(inodes, file.InodeID)
		}
		state.InodeIDs, _ = types.SetValueFrom(ctx, types.StringType, inodes)
	}
	if !state.TenantIDs.IsNull() {
		tenants, err := r.client.ListDeploymentPolicyTenants(ctx, state.ID.ValueString())
		if err != nil {
			return fmt.Errorf("list tenants: %w", err)
		}
		if tenants == nil {
			tenants = []string{}
		}
		state.TenantIDs, _ = types.SetValueFrom(ctx, types.StringType, tenants)
	}
	return nil
}
func (r *deploymentPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan deploymentPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var sourceSwapPolicies []string
	if !plan.SourceSwapPolicies.IsNull() && !plan.SourceSwapPolicies.IsUnknown() {
		plan.SourceSwapPolicies.ElementsAs(ctx, &sourceSwapPolicies, false)
	}
	value, err := r.client.CreateDeploymentPolicy(ctx, sigma.CreateDeploymentPolicyInput{
		Name:               plan.Name.ValueString(),
		VersionTagID:       optionalStringPtr(plan.VersionTagID),
		SourceSwapPolicies: sourceSwapPolicies,
		NameInTenant:       optionalStringPtr(plan.NameInTenant),
		CopyInputTableData: optionalBoolPtr(plan.CopyInputTableData),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma deployment policy", err.Error())
		return
	}
	setDeploymentPolicy(ctx, &plan, value)
	if !r.syncAttachments(ctx, value.DeploymentPolicyID, &plan, &resp.Diagnostics) {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *deploymentPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state deploymentPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.GetDeploymentPolicy(ctx, state.ID.ValueString())
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma deployment policy", err.Error())
		return
	}
	setDeploymentPolicy(ctx, &state, value)
	if err := r.readAttachments(ctx, &state); err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma deployment policy attachments", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *deploymentPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan deploymentPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var sourceSwapPolicies *[]string
	if !plan.SourceSwapPolicies.IsNull() && !plan.SourceSwapPolicies.IsUnknown() {
		values := []string{}
		plan.SourceSwapPolicies.ElementsAs(ctx, &values, false)
		sourceSwapPolicies = &values
	}
	name := plan.Name.ValueString()
	in := sigma.UpdateDeploymentPolicyInput{
		Name:               &name,
		NameInTenant:       optionalStringPtr(plan.NameInTenant),
		SourceSwapPolicies: sourceSwapPolicies,
		CopyInputTableData: optionalBoolPtr(plan.CopyInputTableData),
	}
	value, err := r.client.UpdateDeploymentPolicy(ctx, plan.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Sigma deployment policy", err.Error())
		return
	}
	setDeploymentPolicy(ctx, &plan, value)
	if !r.syncAttachments(ctx, plan.ID.ValueString(), &plan, &resp.Diagnostics) {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *deploymentPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state deploymentPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.ArchiveDeploymentPolicy(ctx, state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to archive Sigma deployment policy", err.Error())
	}
}
func (r *deploymentPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importPassthrough(ctx, req, resp)
}

type sourceSwapPolicyResource struct{ configuredResource }

type sourceSwapPolicyModel struct {
	ID               types.String `tfsdk:"id"`
	Type             types.String `tfsdk:"type"`
	Name             types.String `tfsdk:"name"`
	FromConnectionID types.String `tfsdk:"from_connection_id"`
	SwapsJSON        types.String `tfsdk:"swaps_json"`
}

func NewSourceSwapPolicyResource() resource.Resource { return &sourceSwapPolicyResource{} }

func (r *sourceSwapPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_source_swap_policy"
}
func (r *sourceSwapPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *sourceSwapPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Sigma source swap policy. `swaps_json` is polymorphic by `type` (`connection` or `deployment`). " + betaAPINotice,
		Attributes: map[string]schema.Attribute{
			"id":                 schema.StringAttribute{Computed: true, MarkdownDescription: "Source swap policy ID."},
			"type":               schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Policy type: `connection` or `deployment`."},
			"name":               schema.StringAttribute{Required: true, MarkdownDescription: "Source swap policy name."},
			"from_connection_id": schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Connection ID to swap from."},
			"swaps_json":         schema.StringAttribute{Required: true, MarkdownDescription: "JSON swaps object accepted by the Sigma source swap policy API."},
		},
	}
}
func sourceSwapInput(plan *sourceSwapPolicyModel) (sigma.SourceSwapPolicyInput, error) {
	swaps, err := rawJSON(plan.SwapsJSON)
	if err != nil {
		return sigma.SourceSwapPolicyInput{}, fmt.Errorf("decode swaps_json: %w", err)
	}
	if len(swaps) == 0 {
		return sigma.SourceSwapPolicyInput{}, fmt.Errorf("swaps_json is required")
	}
	return sigma.SourceSwapPolicyInput{
		Type:             plan.Type.ValueString(),
		Name:             plan.Name.ValueString(),
		FromConnectionID: plan.FromConnectionID.ValueString(),
		Swaps:            swaps,
	}, nil
}
func setSourceSwapPolicy(state *sourceSwapPolicyModel, value *sigma.SourceSwapPolicy) {
	state.ID = types.StringValue(value.PolicyID)
	state.Type = types.StringValue(value.Type)
	state.Name = types.StringValue(value.Name)
	state.FromConnectionID = types.StringValue(value.FromConnectionID)
	if len(value.Swaps) > 0 && string(value.Swaps) != "null" {
		state.SwapsJSON = jsonString(value.Swaps)
	}
}
func (r *sourceSwapPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sourceSwapPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in, err := sourceSwapInput(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Sigma source swap policy configuration", err.Error())
		return
	}
	value, err := r.client.CreateSourceSwapPolicy(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma source swap policy", err.Error())
		return
	}
	setSourceSwapPolicy(&plan, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *sourceSwapPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sourceSwapPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.GetSourceSwapPolicy(ctx, state.ID.ValueString())
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma source swap policy", err.Error())
		return
	}
	setSourceSwapPolicy(&state, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *sourceSwapPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan sourceSwapPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in, err := sourceSwapInput(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Sigma source swap policy configuration", err.Error())
		return
	}
	value, err := r.client.UpdateSourceSwapPolicy(ctx, plan.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Sigma source swap policy", err.Error())
		return
	}
	setSourceSwapPolicy(&plan, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *sourceSwapPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sourceSwapPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteSourceSwapPolicy(ctx, state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma source swap policy", err.Error())
	}
}
func (r *sourceSwapPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importPassthrough(ctx, req, resp)
}

var (
	_ resource.Resource                = (*tenantResource)(nil)
	_ resource.ResourceWithImportState = (*tenantResource)(nil)
	_ resource.Resource                = (*tenantDeploymentCapabilitiesResource)(nil)
	_ resource.ResourceWithImportState = (*tenantDeploymentCapabilitiesResource)(nil)
	_ resource.Resource                = (*deploymentPolicyResource)(nil)
	_ resource.ResourceWithImportState = (*deploymentPolicyResource)(nil)
	_ resource.Resource                = (*sourceSwapPolicyResource)(nil)
	_ resource.ResourceWithImportState = (*sourceSwapPolicyResource)(nil)
)
