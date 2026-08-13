package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*connectionsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*connectionsDataSource)(nil)
)

type connectionsDataSource struct{ configuredDataSource }

type connectionsDataModel struct {
	ID          types.String          `tfsdk:"id"`
	Connections []connectionDataModel `tfsdk:"connections"`
}

func NewConnectionsDataSource() datasource.DataSource { return &connectionsDataSource{} }

func (d *connectionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connections"
}

func (d *connectionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}

func (d *connectionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Lists Sigma connections.", Attributes: map[string]schema.Attribute{
		"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
		"connections": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Connections visible to the caller.", NestedObject: schema.NestedAttributeObject{Attributes: connectionDataAttributes(false)}},
	}}
}

func (d *connectionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	values, err := d.client.ListConnections(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma connections", err.Error())
		return
	}
	state := connectionsDataModel{ID: types.StringValue("connections")}
	for i := range values {
		state.Connections = append(state.Connections, connectionData(&values[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
