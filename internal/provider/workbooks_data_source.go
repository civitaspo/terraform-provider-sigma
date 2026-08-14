package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*workbooksDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*workbooksDataSource)(nil)
)

type workbooksDataSource struct{ configuredDataSource }

type workbooksDocModel struct {
	ID                  types.String       `tfsdk:"id"`
	ExcludeTags         types.Bool         `tfsdk:"exclude_tags"`
	SkipPermissionCheck types.Bool         `tfsdk:"skip_permission_check"`
	IsArchived          types.Bool         `tfsdk:"is_archived"`
	ExcludeExplorations types.Bool         `tfsdk:"exclude_explorations"`
	Workbooks           []workbookDocModel `tfsdk:"workbooks"`
}

func NewWorkbooksDataSource() datasource.DataSource { return &workbooksDataSource{} }

func (d *workbooksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workbooks"
}
func (d *workbooksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}
func (d *workbooksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Lists Sigma workbooks." + listCollectionNotice, Attributes: map[string]schema.Attribute{
		"id":                    schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
		"exclude_tags":          schema.BoolAttribute{Optional: true, MarkdownDescription: "Whether to exclude tags (`excludeTags`). Explicit `false` is sent; null omits the parameter."},
		"skip_permission_check": schema.BoolAttribute{Optional: true, MarkdownDescription: "Whether to skip permission checks (`skipPermissionCheck`). Explicit `false` is sent; null omits the parameter."},
		"is_archived":           schema.BoolAttribute{Optional: true, MarkdownDescription: "Archived filter (`isArchived`). Explicit `false` is sent; null omits the parameter."},
		"exclude_explorations":  schema.BoolAttribute{Optional: true, MarkdownDescription: "Whether to exclude explorations (`excludeExplorations`). Explicit `false` is sent; null omits the parameter."},
		"workbooks":             schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Workbooks in API order.", NestedObject: schema.NestedAttributeObject{Attributes: workbookDataAttributes(false)}},
	}}
}
func (d *workbooksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state workbooksDocModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if abortUnknownInputs(&resp.Diagnostics, state.ExcludeTags, state.SkipPermissionCheck, state.IsArchived, state.ExcludeExplorations) {
		return
	}
	values, err := d.client.ListWorkbooks(ctx, sigma.ListWorkbooksOptions{
		ExcludeTags:         optionalBoolPtr(state.ExcludeTags),
		SkipPermissionCheck: optionalBoolPtr(state.SkipPermissionCheck),
		IsArchived:          optionalBoolPtr(state.IsArchived),
		ExcludeExplorations: optionalBoolPtr(state.ExcludeExplorations),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma workbooks", err.Error())
		return
	}
	state.ID = types.StringValue("workbooks")
	state.Workbooks = make([]workbookDocModel, 0, len(values))
	for i := range values {
		state.Workbooks = append(state.Workbooks, workbookDoc(&values[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
