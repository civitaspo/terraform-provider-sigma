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

type workbookTagModel struct {
	VersionTagID          types.String  `tfsdk:"version_tag_id"`
	Name                  types.String  `tfsdk:"name"`
	SourceWorkbookVersion types.Float64 `tfsdk:"source_workbook_version"`
	TaggedWorkbookID      types.String  `tfsdk:"tagged_workbook_id"`
	WorkbookTaggedAt      types.String  `tfsdk:"workbook_tagged_at"`
}

type workbookLookupModel struct {
	ID                       types.String       `tfsdk:"id"`
	IncludeTaggedSourceURLID types.Bool         `tfsdk:"include_tagged_source_url_id"`
	URLID                    types.String       `tfsdk:"url_id"`
	Name                     types.String       `tfsdk:"name"`
	URL                      types.String       `tfsdk:"url"`
	Path                     types.String       `tfsdk:"path"`
	LatestVersion            types.Float64      `tfsdk:"latest_version"`
	OwnerID                  types.String       `tfsdk:"owner_id"`
	CreatedBy                types.String       `tfsdk:"created_by"`
	UpdatedBy                types.String       `tfsdk:"updated_by"`
	CreatedAt                types.String       `tfsdk:"created_at"`
	UpdatedAt                types.String       `tfsdk:"updated_at"`
	IsArchived               types.Bool         `tfsdk:"is_archived"`
	Description              types.String       `tfsdk:"description"`
	TaggedSourceURLID        types.String       `tfsdk:"tagged_source_url_id"`
	Tags                     []workbookTagModel `tfsdk:"tags"`
}

type workbookDocModel struct {
	ID                types.String       `tfsdk:"id"`
	URLID             types.String       `tfsdk:"url_id"`
	Name              types.String       `tfsdk:"name"`
	URL               types.String       `tfsdk:"url"`
	Path              types.String       `tfsdk:"path"`
	LatestVersion     types.Float64      `tfsdk:"latest_version"`
	OwnerID           types.String       `tfsdk:"owner_id"`
	CreatedBy         types.String       `tfsdk:"created_by"`
	UpdatedBy         types.String       `tfsdk:"updated_by"`
	CreatedAt         types.String       `tfsdk:"created_at"`
	UpdatedAt         types.String       `tfsdk:"updated_at"`
	IsArchived        types.Bool         `tfsdk:"is_archived"`
	Description       types.String       `tfsdk:"description"`
	TaggedSourceURLID types.String       `tfsdk:"tagged_source_url_id"`
	Tags              []workbookTagModel `tfsdk:"tags"`
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
	var state workbookLookupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, idDiags := knownString(state.ID, "id")
	resp.Diagnostics.Append(idDiags...)
	if abortUnknownInputs(&resp.Diagnostics, state.IncludeTaggedSourceURLID) || resp.Diagnostics.HasError() {
		return
	}
	include := optionalBoolPtr(state.IncludeTaggedSourceURLID)
	value, err := d.client.GetWorkbook(ctx, id, include)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma workbook", err.Error())
		return
	}
	item := workbookDoc(value)
	mapped := workbookLookupModel{
		ID: item.ID, IncludeTaggedSourceURLID: state.IncludeTaggedSourceURLID, URLID: item.URLID, Name: item.Name, URL: item.URL, Path: item.Path,
		LatestVersion: item.LatestVersion, OwnerID: item.OwnerID, CreatedBy: item.CreatedBy, UpdatedBy: item.UpdatedBy,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt, IsArchived: item.IsArchived, Description: item.Description,
		TaggedSourceURLID: item.TaggedSourceURLID, Tags: item.Tags,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, mapped)...)
}

func workbookDataAttributes(requireID bool) map[string]schema.Attribute {
	id := schema.StringAttribute{Computed: true, MarkdownDescription: "Workbook ID."}
	attrs := map[string]schema.Attribute{
		"id":                   id,
		"url_id":               schema.StringAttribute{Computed: true, MarkdownDescription: "Workbook URL ID."},
		"name":                 schema.StringAttribute{Computed: true, MarkdownDescription: "Workbook name."},
		"url":                  schema.StringAttribute{Computed: true, MarkdownDescription: "Workbook URL."},
		"path":                 schema.StringAttribute{Computed: true, MarkdownDescription: "Workbook path."},
		"latest_version":       schema.Float64Attribute{Computed: true, MarkdownDescription: "Latest workbook version."},
		"owner_id":             schema.StringAttribute{Computed: true, MarkdownDescription: "Owner member ID."},
		"created_by":           schema.StringAttribute{Computed: true, MarkdownDescription: "Creator member ID."},
		"updated_by":           schema.StringAttribute{Computed: true, MarkdownDescription: "Last updater member ID."},
		"created_at":           schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp."},
		"updated_at":           schema.StringAttribute{Computed: true, MarkdownDescription: "Update timestamp."},
		"is_archived":          schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the workbook is archived."},
		"description":          schema.StringAttribute{Computed: true, MarkdownDescription: "Workbook description."},
		"tagged_source_url_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Source workbook URL ID for a tenant-deployed workbook. Present when Sigma returns `taggedSourceUrlId`."},
		"tags":                 schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Version tags on the workbook.", NestedObject: schema.NestedAttributeObject{Attributes: workbookTagAttributes()}},
	}
	if requireID {
		attrs["id"] = schema.StringAttribute{Required: true, MarkdownDescription: "Workbook ID."}
		attrs["include_tagged_source_url_id"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "When true, GET includes `includeTaggedSourceUrlId=true` so Sigma may return `taggedSourceUrlId`. List workbooks has no equivalent query parameter."}
	}
	return attrs
}

func workbookTagAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"version_tag_id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Version tag ID."},
		"name":                    schema.StringAttribute{Computed: true, MarkdownDescription: "Tag name."},
		"source_workbook_version": schema.Float64Attribute{Computed: true, MarkdownDescription: "Source workbook version."},
		"tagged_workbook_id":      schema.StringAttribute{Computed: true, MarkdownDescription: "Tagged workbook ID."},
		"workbook_tagged_at":      schema.StringAttribute{Computed: true, MarkdownDescription: "When the workbook was tagged."},
	}
}

func workbookDoc(value *sigma.Workbook) workbookDocModel {
	tags := make([]workbookTagModel, 0, len(value.Tags))
	for _, tag := range value.Tags {
		tags = append(tags, workbookTagModel{
			VersionTagID: types.StringValue(tag.VersionTagID), Name: types.StringValue(tag.Name),
			SourceWorkbookVersion: types.Float64Value(tag.SourceWorkbookVersion), TaggedWorkbookID: types.StringValue(tag.TaggedWorkbookID),
			WorkbookTaggedAt: types.StringValue(tag.WorkbookTaggedAt),
		})
	}
	return workbookDocModel{
		ID: types.StringValue(value.WorkbookID), URLID: types.StringValue(value.WorkbookURLID), Name: types.StringValue(value.Name),
		URL: types.StringValue(value.URL), Path: types.StringValue(value.Path), LatestVersion: types.Float64Value(value.LatestVersion),
		OwnerID: types.StringValue(value.OwnerID), CreatedBy: types.StringValue(value.CreatedBy), UpdatedBy: types.StringValue(value.UpdatedBy),
		CreatedAt: types.StringValue(value.CreatedAt), UpdatedAt: types.StringValue(value.UpdatedAt), IsArchived: types.BoolValue(value.IsArchived),
		Description: nullableString(value.Description), TaggedSourceURLID: nullableString(value.TaggedSourceURLID), Tags: tags,
	}
}
