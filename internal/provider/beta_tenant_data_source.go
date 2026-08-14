package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = (*tenantDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*tenantDataSource)(nil)
)

type tenantDataSource struct{ configuredDataSource }

func NewTenantDataSource() datasource.DataSource { return &tenantDataSource{} }

func (d *tenantDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant"
}

func (d *tenantDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}

func (d *tenantDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a Sigma tenant organization by ID. " + betaDataSourceNotice,
		Attributes:          tenantDataAttributes(true),
	}
}

func (d *tenantDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state tenantDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, idDiags := knownString(state.ID, "id")
	resp.Diagnostics.Append(idDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := d.client.GetTenant(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma tenant", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, tenantData(value))...)
}
