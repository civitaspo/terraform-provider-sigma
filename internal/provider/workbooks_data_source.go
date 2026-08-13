package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*workbooksDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*workbooksDataSource)(nil)
)

type workbooksDataSource struct{ configuredDataSource }

type workbooksDocModel struct {
	ID        types.String       `tfsdk:"id"`
	Workbooks []workbookDocModel `tfsdk:"workbooks"`
}

func NewWorkbooksDataSource() datasource.DataSource { return &workbooksDataSource{} }

func (d *workbooksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workbooks"
}
func (d *workbooksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}
func (d *workbooksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Lists Sigma workbooks.", Attributes: map[string]schema.Attribute{
		"id":        schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
		"workbooks": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Workbooks.", NestedObject: schema.NestedAttributeObject{Attributes: workbookDataAttributes(false)}},
	}}
}
func (d *workbooksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	values, err := d.client.ListWorkbooks(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma workbooks", err.Error())
		return
	}
	state := workbooksDocModel{ID: types.StringValue("workbooks")}
	for i := range values {
		state.Workbooks = append(state.Workbooks, workbookDoc(&values[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
