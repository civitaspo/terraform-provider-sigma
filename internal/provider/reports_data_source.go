package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*reportsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*reportsDataSource)(nil)
)

type reportsDataSource struct{ configuredDataSource }

type reportsDocModel struct {
	ID      types.String     `tfsdk:"id"`
	Reports []reportDocModel `tfsdk:"reports"`
}

func NewReportsDataSource() datasource.DataSource { return &reportsDataSource{} }

func (d *reportsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_reports"
}
func (d *reportsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}
func (d *reportsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Lists Sigma reports.", Attributes: map[string]schema.Attribute{
		"id":      schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
		"reports": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Reports.", NestedObject: schema.NestedAttributeObject{Attributes: reportDataAttributes(false)}},
	}}
}
func (d *reportsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	values, err := d.client.ListReports(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma reports", err.Error())
		return
	}
	state := reportsDocModel{ID: types.StringValue("reports")}
	for i := range values {
		state.Reports = append(state.Reports, reportDoc(&values[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
