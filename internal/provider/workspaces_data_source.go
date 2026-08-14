package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
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
	Name       types.String         `tfsdk:"name"`
	ExactName  types.String         `tfsdk:"exact_name"`
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
		MarkdownDescription: "Lists Sigma workspaces using the paginated v2.1 endpoint." + listCollectionNotice,
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
			"name":       schema.StringAttribute{Optional: true, MarkdownDescription: "Partial name filter (`name`)."},
			"exact_name": schema.StringAttribute{Optional: true, MarkdownDescription: "Exact name filter (`exactName`)."},
			"workspaces": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Workspaces in API order.", NestedObject: schema.NestedAttributeObject{Attributes: workspaceDataAttributes(false)}},
		},
	}
}

func (d *workspacesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state workspacesDataModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	if abortUnknownInputs(&response.Diagnostics, state.Name, state.ExactName) {
		return
	}
	values, err := d.client.ListWorkspaces(ctx, sigma.ListWorkspacesOptions{
		Name:      optionalStringPtr(state.Name),
		ExactName: optionalStringPtr(state.ExactName),
	})
	if err != nil {
		response.Diagnostics.AddError("Unable to list Sigma workspaces", err.Error())
		return
	}
	state.ID = types.StringValue("workspaces")
	state.Workspaces = make([]workspaceDataModel, 0, len(values))
	for i := range values {
		state.Workspaces = append(state.Workspaces, workspaceData(&values[i]))
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
