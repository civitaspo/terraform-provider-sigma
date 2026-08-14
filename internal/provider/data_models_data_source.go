package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*dataModelsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*dataModelsDataSource)(nil)
)

type dataModelsDataSource struct{ configuredDataSource }

type dataModelsDocModel struct {
	ID                  types.String        `tfsdk:"id"`
	ExcludeTags         types.Bool          `tfsdk:"exclude_tags"`
	SkipPermissionCheck types.Bool          `tfsdk:"skip_permission_check"`
	DataModels          []dataModelDocModel `tfsdk:"data_models"`
}

func NewDataModelsDataSource() datasource.DataSource { return &dataModelsDataSource{} }

func (d *dataModelsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_models"
}
func (d *dataModelsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}
func (d *dataModelsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Lists Sigma data models." + listCollectionNotice, Attributes: map[string]schema.Attribute{
		"id":                    schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
		"exclude_tags":          schema.BoolAttribute{Optional: true, MarkdownDescription: "Whether to exclude tags (`excludeTags`). Explicit `false` is sent; null omits the parameter."},
		"skip_permission_check": schema.BoolAttribute{Optional: true, MarkdownDescription: "Whether to skip permission checks (`skipPermissionCheck`). Explicit `false` is sent; null omits the parameter."},
		"data_models":           schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Data models in API order.", NestedObject: schema.NestedAttributeObject{Attributes: dataModelDataAttributes(false)}},
	}}
}
func (d *dataModelsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state dataModelsDocModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if abortUnknownInputs(&resp.Diagnostics, state.ExcludeTags, state.SkipPermissionCheck) {
		return
	}
	values, err := d.client.ListDataModels(ctx, sigma.ListDataModelsOptions{
		ExcludeTags:         optionalBoolPtr(state.ExcludeTags),
		SkipPermissionCheck: optionalBoolPtr(state.SkipPermissionCheck),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma data models", err.Error())
		return
	}
	state.ID = types.StringValue("data_models")
	state.DataModels = make([]dataModelDocModel, 0, len(values))
	for i := range values {
		state.DataModels = append(state.DataModels, dataModelDoc(&values[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
