package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*connectionPathsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*connectionPathsDataSource)(nil)
)

type connectionPathsDataSource struct{ configuredDataSource }

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

func NewConnectionPathsDataSource() datasource.DataSource { return &connectionPathsDataSource{} }

func (d *connectionPathsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection_paths"
}

func (d *connectionPathsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}

func (d *connectionPathsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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

func (d *connectionPathsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
}
