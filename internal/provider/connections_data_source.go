package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
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
	ID              types.String          `tfsdk:"id"`
	Search          types.String          `tfsdk:"search"`
	IncludeArchived types.Bool            `tfsdk:"include_archived"`
	Connections     []connectionDataModel `tfsdk:"connections"`
}

func NewConnectionsDataSource() datasource.DataSource { return &connectionsDataSource{} }

func (d *connectionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connections"
}

func (d *connectionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}

func (d *connectionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Lists Sigma connections. Credentials are never exposed." + listCollectionNotice, Attributes: map[string]schema.Attribute{
		"id":               schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
		"search":           schema.StringAttribute{Optional: true, MarkdownDescription: "Search filter (`search`)."},
		"include_archived": schema.BoolAttribute{Optional: true, MarkdownDescription: "Whether archived connections are included (`includeArchived`). Explicit `false` is sent; null omits the parameter."},
		"connections":      schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Connections in API order.", NestedObject: schema.NestedAttributeObject{Attributes: connectionDataAttributes(false)}},
	}}
}

func (d *connectionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state connectionsDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if abortUnknownInputs(&resp.Diagnostics, state.Search, state.IncludeArchived) {
		return
	}
	values, err := d.client.ListConnections(ctx, sigma.ListConnectionsOptions{
		Search:          optionalStringPtr(state.Search),
		IncludeArchived: optionalBoolPtr(state.IncludeArchived),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma connections", err.Error())
		return
	}
	state.ID = types.StringValue("connections")
	state.Connections = make([]connectionDataModel, 0, len(values))
	for i := range values {
		item, diags := connectionData(ctx, &values[i])
		resp.Diagnostics.Append(diags...)
		state.Connections = append(state.Connections, item)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
