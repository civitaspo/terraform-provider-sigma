package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*workspaceDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*workspaceDataSource)(nil)
)

type workspaceDataSource struct{ configuredDataSource }

type workspaceDataModel struct {
	ID        types.String `tfsdk:"id"`
	URLID     types.String `tfsdk:"url_id"`
	Name      types.String `tfsdk:"name"`
	CreatedBy types.String `tfsdk:"created_by"`
	UpdatedBy types.String `tfsdk:"updated_by"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func NewWorkspaceDataSource() datasource.DataSource { return &workspaceDataSource{} }

func (d *workspaceDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_workspace"
}

func (d *workspaceDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	d.configure(request, response)
}

func (d *workspaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a Sigma workspace by ID.",
		Attributes:          workspaceDataAttributes(true),
	}
}

func (d *workspaceDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state workspaceDataModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	value, err := d.client.GetWorkspace(ctx, state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Unable to read Sigma workspace", err.Error())
		return
	}
	state = workspaceData(value)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func workspaceDataAttributes(requireID bool) map[string]schema.Attribute {
	id := schema.StringAttribute{Computed: true, MarkdownDescription: "Workspace ID."}
	if requireID {
		id = schema.StringAttribute{Required: true, MarkdownDescription: "Workspace ID."}
	}
	return map[string]schema.Attribute{
		"id":         id,
		"url_id":     schema.StringAttribute{Computed: true, MarkdownDescription: "Base62 identifier used in Sigma workspace URLs."},
		"name":       schema.StringAttribute{Computed: true, MarkdownDescription: "Workspace name."},
		"created_by": schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the member who created the workspace."},
		"updated_by": schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the member who last updated the workspace."},
		"created_at": schema.StringAttribute{Computed: true, MarkdownDescription: "Workspace creation timestamp."},
		"updated_at": schema.StringAttribute{Computed: true, MarkdownDescription: "Workspace update timestamp."},
	}
}

func workspaceData(value *sigma.Workspace) workspaceDataModel {
	return workspaceDataModel{
		ID: types.StringValue(value.WorkspaceID), URLID: types.StringValue(value.WorkspaceURLID), Name: types.StringValue(value.Name),
		CreatedBy: types.StringValue(value.CreatedBy), UpdatedBy: types.StringValue(value.UpdatedBy),
		CreatedAt: types.StringValue(value.CreatedAt), UpdatedAt: types.StringValue(value.UpdatedAt),
	}
}
