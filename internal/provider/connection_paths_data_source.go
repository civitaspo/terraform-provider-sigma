package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*connectionPathsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*connectionPathsDataSource)(nil)
)

type connectionPathsDataSource struct{ configuredDataSource }

type connectionPathListItemModel struct {
	ConnectionID types.String `tfsdk:"connection_id"`
	URLID        types.String `tfsdk:"url_id"`
	Path         types.List   `tfsdk:"path"`
}

type connectionPathsDataModel struct {
	ID           types.String                  `tfsdk:"id"`
	ConnectionID types.String                  `tfsdk:"connection_id"`
	Paths        []connectionPathListItemModel `tfsdk:"paths"`
}

func NewConnectionPathsDataSource() datasource.DataSource { return &connectionPathsDataSource{} }

func (d *connectionPathsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection_paths"
}

func (d *connectionPathsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}

func (d *connectionPathsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Lists Sigma connection paths using GET `/v2/connections/paths`. Warehouse catalogs can contain tens of thousands of paths; the provider requests 500 rows per page and paces pagination to avoid Cloudflare 429s. " + listCollectionNotice, Attributes: map[string]schema.Attribute{
		"id":            schema.StringAttribute{Computed: true, MarkdownDescription: "Stable list identifier. Includes the optional connection filter when set."},
		"connection_id": schema.StringAttribute{Optional: true, MarkdownDescription: "Optional connection filter (`connectionId`)."},
		"paths": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Connection paths in API order.", NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
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
	if abortUnknownInputs(&resp.Diagnostics, state.ConnectionID) {
		return
	}
	values, err := d.client.ListConnectionPaths(ctx, sigma.ListConnectionPathsOptions{ConnectionID: optionalStringPtr(state.ConnectionID)})
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma connection paths", err.Error())
		return
	}
	id := "connection-paths"
	if ptr := optionalStringPtr(state.ConnectionID); ptr != nil {
		id += ":" + *ptr
	}
	state.ID = types.StringValue(id)
	state.Paths = make([]connectionPathListItemModel, 0, len(values))
	for _, value := range values {
		pathValue, diagnostics := types.ListValueFrom(ctx, types.StringType, value.Path)
		resp.Diagnostics.Append(diagnostics...)
		state.Paths = append(state.Paths, connectionPathListItemModel{ConnectionID: types.StringValue(value.ConnectionID), URLID: types.StringValue(value.URLID), Path: pathValue})
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
