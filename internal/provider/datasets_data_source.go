package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*datasetsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*datasetsDataSource)(nil)
)

type datasetsDataSource struct{ configuredDataSource }

type datasetsDocModel struct {
	ID       types.String      `tfsdk:"id"`
	Datasets []datasetDocModel `tfsdk:"datasets"`
}

func NewDatasetsDataSource() datasource.DataSource { return &datasetsDataSource{} }

func (d *datasetsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datasets"
}
func (d *datasetsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}
func (d *datasetsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Sigma datasets.",
		DeprecationMessage:  datasetDeprecation,
		Attributes: map[string]schema.Attribute{
			"id":       schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
			"datasets": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Datasets.", NestedObject: schema.NestedAttributeObject{Attributes: datasetDataAttributes(false)}},
		},
	}
}
func (d *datasetsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	values, err := d.client.ListDatasets(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma datasets", err.Error())
		return
	}
	state := datasetsDocModel{ID: types.StringValue("datasets")}
	for i := range values {
		state.Datasets = append(state.Datasets, datasetDoc(&values[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
