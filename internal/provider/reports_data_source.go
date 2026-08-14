package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*reportsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*reportsDataSource)(nil)
)

type reportsDataSource struct{ configuredDataSource }

type reportsDocModel struct {
	ID                  types.String     `tfsdk:"id"`
	ExcludeTags         types.Bool       `tfsdk:"exclude_tags"`
	SkipPermissionCheck types.Bool       `tfsdk:"skip_permission_check"`
	IsArchived          types.Bool       `tfsdk:"is_archived"`
	Reports             []reportDocModel `tfsdk:"reports"`
}

func NewReportsDataSource() datasource.DataSource { return &reportsDataSource{} }

func (d *reportsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_reports"
}
func (d *reportsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}
func (d *reportsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Lists Sigma reports." + listCollectionNotice, Attributes: map[string]schema.Attribute{
		"id":                    schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
		"exclude_tags":          schema.BoolAttribute{Optional: true, MarkdownDescription: "Whether to exclude tags (`excludeTags`). Explicit `false` is sent; null omits the parameter."},
		"skip_permission_check": schema.BoolAttribute{Optional: true, MarkdownDescription: "Whether to skip permission checks (`skipPermissionCheck`). Explicit `false` is sent; null omits the parameter."},
		"is_archived":           schema.BoolAttribute{Optional: true, MarkdownDescription: "Archived filter (`isArchived`). Explicit `false` is sent; null omits the parameter."},
		"reports":               schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Reports in API order.", NestedObject: schema.NestedAttributeObject{Attributes: reportDataAttributes(false)}},
	}}
}
func (d *reportsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state reportsDocModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if abortUnknownInputs(&resp.Diagnostics, state.ExcludeTags, state.SkipPermissionCheck, state.IsArchived) {
		return
	}
	values, err := d.client.ListReports(ctx, sigma.ListReportsOptions{
		ExcludeTags:         optionalBoolPtr(state.ExcludeTags),
		SkipPermissionCheck: optionalBoolPtr(state.SkipPermissionCheck),
		IsArchived:          optionalBoolPtr(state.IsArchived),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma reports", err.Error())
		return
	}
	state.ID = types.StringValue("reports")
	state.Reports = make([]reportDocModel, 0, len(values))
	for i := range values {
		state.Reports = append(state.Reports, reportDoc(&values[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
