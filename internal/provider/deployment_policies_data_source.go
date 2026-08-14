package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*deploymentPoliciesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*deploymentPoliciesDataSource)(nil)
)

type deploymentPoliciesDataSource struct{ configuredDataSource }

type deploymentPolicyDataModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	NameInTenant       types.String `tfsdk:"name_in_tenant"`
	VersionTagID       types.String `tfsdk:"version_tag_id"`
	SourceSwapPolicies types.Set    `tfsdk:"source_swap_policies"`
	CopyInputTableData types.Bool   `tfsdk:"copy_input_table_data"`
}

type deploymentPoliciesDataModel struct {
	ID                 types.String                `tfsdk:"id"`
	DeploymentPolicies []deploymentPolicyDataModel `tfsdk:"deployment_policies"`
}

func NewDeploymentPoliciesDataSource() datasource.DataSource { return &deploymentPoliciesDataSource{} }

func (d *deploymentPoliciesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment_policies"
}
func (d *deploymentPoliciesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}
func (d *deploymentPoliciesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Sigma deployment policies. " + betaDataSourceNotice + listCollectionNotice,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
			"deployment_policies": schema.ListNestedAttribute{
				Computed: true, MarkdownDescription: "Deployment policies visible to the caller.",
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id":             schema.StringAttribute{Computed: true, MarkdownDescription: "Deployment policy ID."},
					"name":           schema.StringAttribute{Computed: true, MarkdownDescription: "Deployment policy name."},
					"name_in_tenant": schema.StringAttribute{Computed: true, MarkdownDescription: "Workspace name created in receiving tenants."},
					"version_tag_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Version tag ID."},
					"source_swap_policies": schema.SetAttribute{
						Computed: true, ElementType: types.StringType,
						MarkdownDescription: "Source swap policy IDs used when deploying documents.",
					},
					"copy_input_table_data": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether input table data is copied when deploying."},
				}},
			},
		},
	}
}
func (d *deploymentPoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state deploymentPoliciesDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	policies, err := d.client.ListDeploymentPolicies(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma deployment policies", err.Error())
		return
	}
	state.ID = types.StringValue("deployment_policies")
	state.DeploymentPolicies = make([]deploymentPolicyDataModel, 0, len(policies))
	for _, policy := range policies {
		item := deploymentPolicyDataModel{
			ID:                 types.StringValue(policy.DeploymentPolicyID),
			Name:               types.StringValue(policy.Name),
			NameInTenant:       types.StringValue(policy.NameInTenant),
			VersionTagID:       stringOrNull(policy.VersionTagID),
			CopyInputTableData: types.BoolValue(policy.CopyInputTableData),
		}
		if policy.SourceSwapPolicies == nil {
			policy.SourceSwapPolicies = []string{}
		}
		swaps, swapDiags := types.SetValueFrom(ctx, types.StringType, policy.SourceSwapPolicies)
		resp.Diagnostics.Append(swapDiags...)
		item.SourceSwapPolicies = swaps
		if resp.Diagnostics.HasError() {
			return
		}
		state.DeploymentPolicies = append(state.DeploymentPolicies, item)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
