package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*dataModelsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*dataModelsDataSource)(nil)
)

type dataModelsDataSource struct{ configuredDataSource }

type dataModelsDocModel struct {
	ID         types.String        `tfsdk:"id"`
	DataModels []dataModelDocModel `tfsdk:"data_models"`
}

func NewDataModelsDataSource() datasource.DataSource { return &dataModelsDataSource{} }

func (d *dataModelsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_models"
}
func (d *dataModelsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}
func (d *dataModelsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Lists Sigma data models.", Attributes: map[string]schema.Attribute{
		"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
		"data_models": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Data models.", NestedObject: schema.NestedAttributeObject{Attributes: dataModelDataAttributes(false)}},
	}}
}
func (d *dataModelsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	values, err := d.client.ListDataModels(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma data models", err.Error())
		return
	}
	state := dataModelsDocModel{ID: types.StringValue("data_models")}
	for i := range values {
		state.DataModels = append(state.DataModels, dataModelDoc(&values[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
