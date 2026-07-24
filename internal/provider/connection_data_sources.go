package provider

import (
	"context"
	"fmt"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type connectionDataSource struct {
	client *sigma.Client
	kind   string
}

func NewConnectionDataSource() datasource.DataSource {
	return &connectionDataSource{kind: "connection"}
}
func NewConnectionsDataSource() datasource.DataSource {
	return &connectionDataSource{kind: "connections"}
}
func NewConnectionPathsDataSource() datasource.DataSource {
	return &connectionDataSource{kind: "connection_paths"}
}
func (d *connectionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.kind
}
func (d *connectionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sigma.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected data source configuration type", "The Sigma provider returned an unexpected client type.")
		return
	}
	d.client = client
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
	}
}
func (d *connectionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	switch d.kind {
	case "connection":
		resp.Schema = schema.Schema{MarkdownDescription: "Retrieves a Sigma connection by ID.", Attributes: connectionDataAttributes(true)}
	case "connections":
		resp.Schema = schema.Schema{MarkdownDescription: "Lists Sigma connections.", Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
			"connections": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Connections visible to the caller.", NestedObject: schema.NestedAttributeObject{Attributes: connectionDataAttributes(false)}},
		}}
	case "connection_paths":
		resp.Schema = schema.Schema{MarkdownDescription: "Lists Sigma connection paths, or looks up one fully qualified path when `path` is configured.", Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
			"connection_id": schema.StringAttribute{Optional: true, MarkdownDescription: "Optional connection filter; required with `path`."},
			"path":          schema.ListAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "Fully qualified path components to look up."},
			"kind":          schema.StringAttribute{Computed: true, MarkdownDescription: "Object kind returned by lookup."},
			"inode_id":      schema.StringAttribute{Computed: true, MarkdownDescription: "Connection path ID returned by lookup."},
			"url":           schema.StringAttribute{Computed: true, MarkdownDescription: "Connection URL returned by lookup."},
			"paths": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Connection paths returned by list.", NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
				"connection_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Connection ID."},
				"url_id":        schema.StringAttribute{Computed: true, MarkdownDescription: "Connection path URL ID."},
				"path":          schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Fully qualified path components."},
			}}},
		}}
	}
}

type connectionDataModel struct {
	ID               types.String  `tfsdk:"id"`
	Name             types.String  `tfsdk:"name"`
	Type             types.String  `tfsdk:"type"`
	DescriptionJSON  types.String  `tfsdk:"description_json"`
	PoolSizesJSON    types.String  `tfsdk:"pool_sizes_json"`
	TimeoutSecs      types.Float64 `tfsdk:"timeout_secs"`
	UseFriendlyNames types.Bool    `tfsdk:"use_friendly_names"`
}
type connectionsDataModel struct {
	ID          types.String          `tfsdk:"id"`
	Connections []connectionDataModel `tfsdk:"connections"`
}
type connectionPathDataModel struct {
	ConnectionID types.String `tfsdk:"connection_id"`
	URLID        types.String `tfsdk:"url_id"`
	Path         types.List   `tfsdk:"path"`
}
type connectionPathsDataModel struct {
	ID           types.String              `tfsdk:"id"`
	ConnectionID types.String              `tfsdk:"connection_id"`
	Path         types.List                `tfsdk:"path"`
	Kind         types.String              `tfsdk:"kind"`
	InodeID      types.String              `tfsdk:"inode_id"`
	URL          types.String              `tfsdk:"url"`
	Paths        []connectionPathDataModel `tfsdk:"paths"`
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
	}
	if value.TimeoutSecs != nil {
		state.TimeoutSecs = types.Float64Value(*value.TimeoutSecs)
	}
	return state
}
func (d *connectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	switch d.kind {
	case "connection":
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
	case "connections":
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
	case "connection_paths":
		var state connectionPathsDataModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if !state.Path.IsNull() && !state.Path.IsUnknown() {
			if state.ConnectionID.IsNull() || state.ConnectionID.ValueString() == "" {
				resp.Diagnostics.AddError("Invalid connection path lookup", "connection_id is required when path is configured.")
				return
			}
			var pathValue []string
			resp.Diagnostics.Append(state.Path.ElementsAs(ctx, &pathValue, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			value, err := d.client.LookupConnection(ctx, state.ConnectionID.ValueString(), pathValue)
			if err != nil {
				resp.Diagnostics.AddError("Unable to look up Sigma connection path", err.Error())
				return
			}
			state.ID = types.StringValue(value.InodeID)
			state.Kind = types.StringValue(value.Kind)
			state.InodeID = types.StringValue(value.InodeID)
			state.URL = types.StringValue(value.URL)
			state.Paths = []connectionPathDataModel{}
		} else {
			values, err := d.client.ListConnectionPaths(ctx, state.ConnectionID.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Unable to list Sigma connection paths", err.Error())
				return
			}
			id := "connection-paths"
			if !state.ConnectionID.IsNull() && state.ConnectionID.ValueString() != "" {
				id += ":" + state.ConnectionID.ValueString()
			}
			state.ID = types.StringValue(id)
			state.Kind, state.InodeID, state.URL = types.StringNull(), types.StringNull(), types.StringNull()
			for _, value := range values {
				pathValue, diagnostics := types.ListValueFrom(ctx, types.StringType, value.Path)
				resp.Diagnostics.Append(diagnostics...)
				state.Paths = append(state.Paths, connectionPathDataModel{ConnectionID: types.StringValue(value.ConnectionID), URLID: types.StringValue(value.URLID), Path: pathValue})
			}
		}
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	default:
		resp.Diagnostics.AddError("Unsupported Sigma connection data source", fmt.Sprintf("Unknown kind %q.", d.kind))
	}
}
