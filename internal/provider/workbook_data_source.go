package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*workbookDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*workbookDataSource)(nil)
)

type workbookDataSource struct{ configuredDataSource }

type workbookDocModel struct {
	ID            types.String  `tfsdk:"id"`
	URLID         types.String  `tfsdk:"url_id"`
	Name          types.String  `tfsdk:"name"`
	URL           types.String  `tfsdk:"url"`
	Path          types.String  `tfsdk:"path"`
	LatestVersion types.Float64 `tfsdk:"latest_version"`
	OwnerID       types.String  `tfsdk:"owner_id"`
	CreatedBy     types.String  `tfsdk:"created_by"`
	UpdatedBy     types.String  `tfsdk:"updated_by"`
	CreatedAt     types.String  `tfsdk:"created_at"`
	UpdatedAt     types.String  `tfsdk:"updated_at"`
	IsArchived    types.Bool    `tfsdk:"is_archived"`
}

func NewWorkbookDataSource() datasource.DataSource { return &workbookDataSource{} }

func (d *workbookDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workbook"
}
func (d *workbookDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}
func (d *workbookDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Retrieves a Sigma workbook by ID.", Attributes: workbookDataAttributes(true)}
}
func (d *workbookDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state workbookDocModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := d.client.GetWorkbook(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma workbook", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, workbookDoc(value))...)
}

func workbookDataAttributes(requireID bool) map[string]schema.Attribute {
	id := schema.StringAttribute{Computed: true, MarkdownDescription: "Workbook ID."}
	if requireID {
		id = schema.StringAttribute{Required: true, MarkdownDescription: "Workbook ID."}
	}
	return map[string]schema.Attribute{
		"id":             id,
		"url_id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Workbook URL ID."},
		"name":           schema.StringAttribute{Computed: true, MarkdownDescription: "Workbook name."},
		"url":            schema.StringAttribute{Computed: true, MarkdownDescription: "Workbook URL."},
		"path":           schema.StringAttribute{Computed: true, MarkdownDescription: "Workbook path."},
		"latest_version": schema.Float64Attribute{Computed: true, MarkdownDescription: "Latest workbook version."},
		"owner_id":       schema.StringAttribute{Computed: true, MarkdownDescription: "Owner member ID."},
		"created_by":     schema.StringAttribute{Computed: true, MarkdownDescription: "Creator member ID."},
		"updated_by":     schema.StringAttribute{Computed: true, MarkdownDescription: "Last updater member ID."},
		"created_at":     schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp."},
		"updated_at":     schema.StringAttribute{Computed: true, MarkdownDescription: "Update timestamp."},
		"is_archived":    schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the workbook is archived."},
	}
}

func workbookDoc(value *sigma.Workbook) workbookDocModel {
	return workbookDocModel{
		ID: types.StringValue(value.WorkbookID), URLID: types.StringValue(value.WorkbookURLID), Name: types.StringValue(value.Name),
		URL: types.StringValue(value.URL), Path: types.StringValue(value.Path), LatestVersion: types.Float64Value(value.LatestVersion),
		OwnerID: types.StringValue(value.OwnerID), CreatedBy: types.StringValue(value.CreatedBy), UpdatedBy: types.StringValue(value.UpdatedBy),
		CreatedAt: types.StringValue(value.CreatedAt), UpdatedAt: types.StringValue(value.UpdatedAt), IsArchived: types.BoolValue(value.IsArchived),
	}
}
