package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*connectionPathDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*connectionPathDataSource)(nil)
)

type connectionPathDataSource struct{ configuredDataSource }

type connectionPathDataModel struct {
	ID           types.String `tfsdk:"id"`
	ConnectionID types.String `tfsdk:"connection_id"`
	Path         types.List   `tfsdk:"path"`
	Kind         types.String `tfsdk:"kind"`
	InodeID      types.String `tfsdk:"inode_id"`
	URL          types.String `tfsdk:"url"`
}

func NewConnectionPathDataSource() datasource.DataSource { return &connectionPathDataSource{} }

func (d *connectionPathDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection_path"
}

func (d *connectionPathDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}

func (d *connectionPathDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Looks up a single Sigma connection path using POST `/v2/connection/{connectionId}/lookup`.", Attributes: map[string]schema.Attribute{
		"id":            schema.StringAttribute{Computed: true, MarkdownDescription: "Returned inode ID."},
		"connection_id": schema.StringAttribute{Required: true, MarkdownDescription: "Connection ID."},
		"path":          schema.ListAttribute{Required: true, ElementType: types.StringType, MarkdownDescription: "Non-empty fully qualified path components."},
		"kind":          schema.StringAttribute{Computed: true, MarkdownDescription: "Object kind returned by lookup."},
		"inode_id":      schema.StringAttribute{Computed: true, MarkdownDescription: "Connection path inode ID."},
		"url":           schema.StringAttribute{Computed: true, MarkdownDescription: "Connection URL returned by lookup."},
	}}
}

func (d *connectionPathDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state connectionPathDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	connectionID, idDiags := knownString(state.ConnectionID, "connection_id")
	resp.Diagnostics.Append(idDiags...)
	if abortUnknownInputs(&resp.Diagnostics, state.Path) || resp.Diagnostics.HasError() {
		return
	}
	if state.Path.IsNull() {
		resp.Diagnostics.AddError("Invalid path", "path must be a known, non-empty list.")
		return
	}
	var pathValue []string
	resp.Diagnostics.Append(state.Path.ElementsAs(ctx, &pathValue, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(pathValue) == 0 {
		resp.Diagnostics.AddError("Invalid path", "path must contain at least one component.")
		return
	}
	value, err := d.client.LookupConnection(ctx, connectionID, pathValue)
	if err != nil {
		resp.Diagnostics.AddError("Unable to look up Sigma connection path", err.Error())
		return
	}
	state.ID = types.StringValue(value.InodeID)
	state.Kind = types.StringValue(value.Kind)
	state.InodeID = types.StringValue(value.InodeID)
	state.URL = types.StringValue(value.URL)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
