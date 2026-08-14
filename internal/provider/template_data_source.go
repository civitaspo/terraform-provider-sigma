package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = (*templateDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*templateDataSource)(nil)
)

type templateDataSource struct{ configuredDataSource }

func NewTemplateDataSource() datasource.DataSource { return &templateDataSource{} }

func (d *templateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template"
}

func (d *templateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}

func (d *templateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a Sigma template by ID.",
		Attributes:          templateDataAttributes(true),
	}
}

func (d *templateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state templateDocModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, idDiags := knownString(state.ID, "id")
	resp.Diagnostics.Append(idDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := d.client.GetTemplate(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma template", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, templateDoc(value))...)
}
