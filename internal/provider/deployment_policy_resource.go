package provider

import (
	"context"
	"fmt"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type deploymentPolicyResource struct{ configuredResource }

var (
	_ resource.Resource                = (*deploymentPolicyResource)(nil)
	_ resource.ResourceWithConfigure   = (*deploymentPolicyResource)(nil)
	_ resource.ResourceWithImportState = (*deploymentPolicyResource)(nil)
)

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
