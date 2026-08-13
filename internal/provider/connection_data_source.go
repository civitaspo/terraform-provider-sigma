package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*connectionDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*connectionDataSource)(nil)
)

type connectionDataSource struct{ configuredDataSource }

type connectionDataModel struct {
	ID               types.String  `tfsdk:"id"`
	Name             types.String  `tfsdk:"name"`
	Type             types.String  `tfsdk:"type"`
	DescriptionJSON  types.String  `tfsdk:"description_json"`
	PoolSizesJSON    types.String  `tfsdk:"pool_sizes_json"`
	TimeoutSecs      types.Float64 `tfsdk:"timeout_secs"`
	UseFriendlyNames types.Bool    `tfsdk:"use_friendly_names"`
	UseOauth         types.Bool    `tfsdk:"use_oauth"`
}

func NewConnectionDataSource() datasource.DataSource { return &connectionDataSource{} }

func (d *connectionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection"
}

func (d *connectionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}

func (d *connectionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Retrieves a Sigma connection by ID.", Attributes: connectionDataAttributes(true)}
}

func (d *connectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state connectionDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := d.client.GetConnection(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma connection", err.Error())
		return
	}
	state = connectionData(value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func connectionDataAttributes(requireID bool) map[string]schema.Attribute {
	id := schema.StringAttribute{Computed: true, MarkdownDescription: "Connection ID."}
	if requireID {
		id = schema.StringAttribute{Required: true, MarkdownDescription: "Connection ID."}
	}
	return map[string]schema.Attribute{
		"id":                 id,
		"name":               schema.StringAttribute{Computed: true, MarkdownDescription: "Connection name."},
		"type":               schema.StringAttribute{Computed: true, MarkdownDescription: "Warehouse type."},
		"description_json":   schema.StringAttribute{Computed: true, MarkdownDescription: "Description JSON returned by Sigma."},
		"pool_sizes_json":    schema.StringAttribute{Computed: true, MarkdownDescription: "Pool sizes JSON returned by Sigma."},
		"timeout_secs":       schema.Float64Attribute{Computed: true, MarkdownDescription: "Connection timeout in seconds."},
		"use_friendly_names": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether friendly names are enabled."},
		"use_oauth":          schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the connection uses OAuth, as returned by Sigma GET `useOauth`."},
	}
}

func connectionData(value *sigma.Connection) connectionDataModel {
	state := connectionDataModel{
		ID:               types.StringValue(value.ConnectionID),
		Name:             types.StringValue(value.Name),
		Type:             types.StringValue(value.Type),
		DescriptionJSON:  jsonString(value.Description),
		PoolSizesJSON:    jsonString(value.PoolSizes),
		UseFriendlyNames: types.BoolValue(value.FriendlyName),
		TimeoutSecs:      types.Float64Null(),
		UseOauth:         types.BoolNull(),
	}
	if value.TimeoutSecs != nil {
		state.TimeoutSecs = types.Float64Value(*value.TimeoutSecs)
	}
	if value.UseOauth != nil {
		state.UseOauth = types.BoolValue(*value.UseOauth)
	}
	return state
}
