package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
		MarkdownDescription: "Manages a Sigma deployment policy. Destroy archives the policy. Attach documents with `sigma_deployment_policy_document` and tenants with `sigma_deployment_policy_tenant`. " + betaAPINotice,
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
		},
	}
}

func setDeploymentPolicy(ctx context.Context, state *deploymentPolicyModel, value *sigma.DeploymentPolicy) diag.Diagnostics {
	var diags diag.Diagnostics
	state.ID = types.StringValue(value.DeploymentPolicyID)
	state.Name = types.StringValue(value.Name)
	state.NameInTenant = types.StringValue(value.NameInTenant)
	state.VersionTagID = stringOrNull(value.VersionTagID)
	state.CopyInputTableData = types.BoolValue(value.CopyInputTableData)
	swaps := value.SourceSwapPolicies
	if swaps == nil {
		swaps = []string{}
	}
	set, setDiags := types.SetValueFrom(ctx, types.StringType, swaps)
	diags.Append(setDiags...)
	state.SourceSwapPolicies = set
	return diags
}

func sourceSwapPoliciesFromPlan(ctx context.Context, value types.Set) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() || value.IsUnknown() {
		return nil, diags
	}
	values := []string{}
	diags.Append(value.ElementsAs(ctx, &values, false)...)
	return values, diags
}

func (r *deploymentPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan deploymentPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	name, nameDiags := knownString(plan.Name, "name")
	resp.Diagnostics.Append(nameDiags...)
	sourceSwapPolicies, swapDiags := sourceSwapPoliciesFromPlan(ctx, plan.SourceSwapPolicies)
	resp.Diagnostics.Append(swapDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.CreateDeploymentPolicy(ctx, sigma.CreateDeploymentPolicyInput{
		Name:               name,
		VersionTagID:       optionalStringPtr(plan.VersionTagID),
		SourceSwapPolicies: sourceSwapPolicies,
		NameInTenant:       optionalStringPtr(plan.NameInTenant),
		CopyInputTableData: optionalBoolPtr(plan.CopyInputTableData),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma deployment policy", err.Error())
		return
	}
	resp.Diagnostics.Append(setDeploymentPolicy(ctx, &plan, value)...)
	if resp.Diagnostics.HasError() {
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
	id, idDiags := knownString(state.ID, "id")
	resp.Diagnostics.Append(idDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.GetDeploymentPolicy(ctx, id)
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma deployment policy", err.Error())
		return
	}
	resp.Diagnostics.Append(setDeploymentPolicy(ctx, &state, value)...)
	if resp.Diagnostics.HasError() {
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
	id, idDiags := knownString(plan.ID, "id")
	name, nameDiags := knownString(plan.Name, "name")
	resp.Diagnostics.Append(idDiags...)
	resp.Diagnostics.Append(nameDiags...)
	var sourceSwapPolicies *[]string
	if !plan.SourceSwapPolicies.IsNull() && !plan.SourceSwapPolicies.IsUnknown() {
		values, swapDiags := sourceSwapPoliciesFromPlan(ctx, plan.SourceSwapPolicies)
		resp.Diagnostics.Append(swapDiags...)
		sourceSwapPolicies = &values
	}
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.UpdateDeploymentPolicy(ctx, id, sigma.UpdateDeploymentPolicyInput{
		Name:               &name,
		NameInTenant:       optionalStringPtr(plan.NameInTenant),
		SourceSwapPolicies: sourceSwapPolicies,
		CopyInputTableData: optionalBoolPtr(plan.CopyInputTableData),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Sigma deployment policy", err.Error())
		return
	}
	resp.Diagnostics.Append(setDeploymentPolicy(ctx, &plan, value)...)
	if resp.Diagnostics.HasError() {
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
	id, idDiags := knownString(state.ID, "id")
	resp.Diagnostics.Append(idDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.ArchiveDeploymentPolicy(ctx, id); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to archive Sigma deployment policy", err.Error())
	}
}
func (r *deploymentPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importPassthrough(ctx, req, resp)
}
