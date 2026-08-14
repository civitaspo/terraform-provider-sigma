package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*dataModelDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*dataModelDataSource)(nil)
)

type dataModelDataSource struct{ configuredDataSource }

type dataModelTagModel struct {
	VersionTagID  types.String  `tfsdk:"version_tag_id"`
	TagName       types.String  `tfsdk:"tag_name"`
	SourceVersion types.Float64 `tfsdk:"source_version"`
	TaggedAt      types.String  `tfsdk:"tagged_at"`
}

type dataModelLookupModel struct {
	ID            types.String        `tfsdk:"id"`
	ExcludeTags   types.Bool          `tfsdk:"exclude_tags"`
	URLID         types.String        `tfsdk:"url_id"`
	Name          types.String        `tfsdk:"name"`
	URL           types.String        `tfsdk:"url"`
	Path          types.String        `tfsdk:"path"`
	LatestVersion types.Float64       `tfsdk:"latest_version"`
	OwnerID       types.String        `tfsdk:"owner_id"`
	CreatedBy     types.String        `tfsdk:"created_by"`
	UpdatedBy     types.String        `tfsdk:"updated_by"`
	CreatedAt     types.String        `tfsdk:"created_at"`
	UpdatedAt     types.String        `tfsdk:"updated_at"`
	IsArchived    types.Bool          `tfsdk:"is_archived"`
	Tags          []dataModelTagModel `tfsdk:"tags"`
}

type dataModelDocModel struct {
	ID            types.String        `tfsdk:"id"`
	URLID         types.String        `tfsdk:"url_id"`
	Name          types.String        `tfsdk:"name"`
	URL           types.String        `tfsdk:"url"`
	Path          types.String        `tfsdk:"path"`
	LatestVersion types.Float64       `tfsdk:"latest_version"`
	OwnerID       types.String        `tfsdk:"owner_id"`
	CreatedBy     types.String        `tfsdk:"created_by"`
	UpdatedBy     types.String        `tfsdk:"updated_by"`
	CreatedAt     types.String        `tfsdk:"created_at"`
	UpdatedAt     types.String        `tfsdk:"updated_at"`
	IsArchived    types.Bool          `tfsdk:"is_archived"`
	Tags          []dataModelTagModel `tfsdk:"tags"`
}

func NewDataModelDataSource() datasource.DataSource { return &dataModelDataSource{} }

func (d *dataModelDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_model"
}
func (d *dataModelDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}
func (d *dataModelDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Retrieves a Sigma data model by ID.", Attributes: dataModelLookupAttributes()}
}
func (d *dataModelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state dataModelLookupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, idDiags := knownString(state.ID, "id")
	resp.Diagnostics.Append(idDiags...)
	if abortUnknownInputs(&resp.Diagnostics, state.ExcludeTags) || resp.Diagnostics.HasError() {
		return
	}
	value, err := d.client.GetDataModel(ctx, id, optionalBoolPtr(state.ExcludeTags))
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma data model", err.Error())
		return
	}
	item := dataModelDoc(value)
	resp.Diagnostics.Append(resp.State.Set(ctx, dataModelLookupModel{
		ID: item.ID, ExcludeTags: state.ExcludeTags, URLID: item.URLID, Name: item.Name, URL: item.URL, Path: item.Path,
		LatestVersion: item.LatestVersion, OwnerID: item.OwnerID, CreatedBy: item.CreatedBy, UpdatedBy: item.UpdatedBy,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt, IsArchived: item.IsArchived, Tags: item.Tags,
	})...)
}

func dataModelLookupAttributes() map[string]schema.Attribute {
	attrs := dataModelDataAttributes(true)
	attrs["exclude_tags"] = schema.BoolAttribute{Optional: true, MarkdownDescription: "Whether to exclude tags (`excludeTags`). Explicit `false` is sent; null omits the parameter."}
	return attrs
}

func dataModelDataAttributes(requireID bool) map[string]schema.Attribute {
	id := schema.StringAttribute{Computed: true, MarkdownDescription: "Data model ID."}
	if requireID {
		id = schema.StringAttribute{Required: true, MarkdownDescription: "Data model ID."}
	}
	return map[string]schema.Attribute{
		"id":             id,
		"url_id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Data model URL ID."},
		"name":           schema.StringAttribute{Computed: true, MarkdownDescription: "Data model name."},
		"url":            schema.StringAttribute{Computed: true, MarkdownDescription: "Data model URL."},
		"path":           schema.StringAttribute{Computed: true, MarkdownDescription: "Data model path."},
		"latest_version": schema.Float64Attribute{Computed: true, MarkdownDescription: "Latest data model version."},
		"owner_id":       schema.StringAttribute{Computed: true, MarkdownDescription: "Owner member ID."},
		"created_by":     schema.StringAttribute{Computed: true, MarkdownDescription: "Creator member ID."},
		"updated_by":     schema.StringAttribute{Computed: true, MarkdownDescription: "Last updater member ID."},
		"created_at":     schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp."},
		"updated_at":     schema.StringAttribute{Computed: true, MarkdownDescription: "Update timestamp."},
		"is_archived":    schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the data model is archived."},
		"tags": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Version tags on the data model.", NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"version_tag_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Version tag ID."},
			"tag_name":       schema.StringAttribute{Computed: true, MarkdownDescription: "Tag name."},
			"source_version": schema.Float64Attribute{Computed: true, MarkdownDescription: "Source version."},
			"tagged_at":      schema.StringAttribute{Computed: true, MarkdownDescription: "When the data model was tagged."},
		}}},
	}
}

func dataModelDoc(value *sigma.DataModel) dataModelDocModel {
	tags := make([]dataModelTagModel, 0, len(value.Tags))
	for _, tag := range value.Tags {
		tags = append(tags, dataModelTagModel{
			VersionTagID: types.StringValue(tag.VersionTagID), TagName: types.StringValue(tag.TagName),
			SourceVersion: types.Float64Value(tag.SourceVersion), TaggedAt: types.StringValue(tag.TaggedAt),
		})
	}
	return dataModelDocModel{
		ID: types.StringValue(value.DataModelID), URLID: types.StringValue(value.DataModelURLID), Name: types.StringValue(value.Name),
		URL: types.StringValue(value.URL), Path: types.StringValue(value.Path), LatestVersion: types.Float64Value(value.LatestVersion),
		OwnerID: types.StringValue(value.OwnerID), CreatedBy: types.StringValue(value.CreatedBy), UpdatedBy: types.StringValue(value.UpdatedBy),
		CreatedAt: types.StringValue(value.CreatedAt), UpdatedAt: types.StringValue(value.UpdatedAt), IsArchived: types.BoolValue(value.IsArchived),
		Tags: tags,
	}
}
