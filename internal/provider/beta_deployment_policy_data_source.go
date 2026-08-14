package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = (*deploymentPolicyDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*deploymentPolicyDataSource)(nil)
)

type deploymentPolicyDataSource struct{ configuredDataSource }

func NewDeploymentPolicyDataSource() datasource.DataSource { return &deploymentPolicyDataSource{} }

func (d *deploymentPolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment_policy"
}

func (d *deploymentPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}

func (d *deploymentPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a Sigma deployment policy by ID. " + betaDataSourceNotice,
		Attributes:          deploymentPolicyDataAttributes(true),
	}
}

func (d *deploymentPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state deploymentPolicyDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, idDiags := knownString(state.ID, "id")
	resp.Diagnostics.Append(idDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := d.client.GetDeploymentPolicy(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma deployment policy", err.Error())
		return
	}
	item, itemDiags := deploymentPolicyData(ctx, value)
	resp.Diagnostics.Append(itemDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, item)...)
}
