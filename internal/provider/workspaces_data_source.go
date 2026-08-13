package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*workspacesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*workspacesDataSource)(nil)
)

type workspacesDataSource struct{ configuredDataSource }

type workspacesDataModel struct {
	ID         types.String         `tfsdk:"id"`
	Workspaces []workspaceDataModel `tfsdk:"workspaces"`
}

func NewWorkspacesDataSource() datasource.DataSource { return &workspacesDataSource{} }

func (d *workspacesDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_workspaces"
}

func (d *workspacesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	d.configure(request, response)
}

func (d *workspacesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Lists all Sigma workspaces using the paginated v2.1 endpoint.",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
			"workspaces": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Workspaces.", NestedObject: schema.NestedAttributeObject{Attributes: workspaceDataAttributes(false)}},
		},
	}
}

func (d *workspacesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	values, err := d.client.ListWorkspaces(ctx)
	if err != nil {
		response.Diagnostics.AddError("Unable to list Sigma workspaces", err.Error())
		return
	}
	state := workspacesDataModel{ID: types.StringValue("workspaces")}
	for i := range values {
		state.Workspaces = append(state.Workspaces, workspaceData(&values[i]))
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
