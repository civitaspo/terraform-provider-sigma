package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*templatesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*templatesDataSource)(nil)
)

type templatesDataSource struct{ configuredDataSource }

type templateTagModel struct {
	ID   types.String `tfsdk:"version_tag_id"`
	Name types.String `tfsdk:"name"`
}

type templateDocModel struct {
	ID            types.String       `tfsdk:"id"`
	URLID         types.String       `tfsdk:"url_id"`
	Name          types.String       `tfsdk:"name"`
	URL           types.String       `tfsdk:"url"`
	Path          types.String       `tfsdk:"path"`
	LatestVersion types.Float64      `tfsdk:"latest_version"`
	CreatedBy     types.String       `tfsdk:"created_by"`
	UpdatedBy     types.String       `tfsdk:"updated_by"`
	CreatedAt     types.String       `tfsdk:"created_at"`
	UpdatedAt     types.String       `tfsdk:"updated_at"`
	IsArchived    types.Bool         `tfsdk:"is_archived"`
	Tags          []templateTagModel `tfsdk:"tags"`
}

type templatesDocModel struct {
	ID        types.String       `tfsdk:"id"`
	Source    types.String       `tfsdk:"source"`
	Search    types.String       `tfsdk:"search"`
	Templates []templateDocModel `tfsdk:"templates"`
}

func NewTemplatesDataSource() datasource.DataSource { return &templatesDataSource{} }

func (d *templatesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_templates"
}
func (d *templatesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}
func (d *templatesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Lists Sigma templates." + listCollectionNotice, Attributes: map[string]schema.Attribute{
		"id":        schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
		"source":    schema.StringAttribute{Optional: true, MarkdownDescription: "Template source filter (`source`): `internal` or `external`."},
		"search":    schema.StringAttribute{Optional: true, MarkdownDescription: "Search filter (`search`)."},
		"templates": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Templates in API order.", NestedObject: schema.NestedAttributeObject{Attributes: templateDataAttributes(false)}},
	}}
}
func (d *templatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state templatesDocModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if abortUnknownInputs(&resp.Diagnostics, state.Source, state.Search) {
		return
	}
	values, err := d.client.ListTemplates(ctx, sigma.ListTemplatesOptions{Source: optionalStringPtr(state.Source), Search: optionalStringPtr(state.Search)})
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma templates", err.Error())
		return
	}
	state.ID = types.StringValue("templates")
	state.Templates = make([]templateDocModel, 0, len(values))
	for i := range values {
		state.Templates = append(state.Templates, templateDoc(&values[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func templateDataAttributes(requireID bool) map[string]schema.Attribute {
	id := schema.StringAttribute{Computed: true, MarkdownDescription: "Template ID."}
	if requireID {
		id = schema.StringAttribute{Required: true, MarkdownDescription: "Template ID."}
	}
	return map[string]schema.Attribute{
		"id":             id,
		"url_id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Template URL ID."},
		"name":           schema.StringAttribute{Computed: true, MarkdownDescription: "Template name."},
		"url":            schema.StringAttribute{Computed: true, MarkdownDescription: "Template URL."},
		"path":           schema.StringAttribute{Computed: true, MarkdownDescription: "Template path."},
		"latest_version": schema.Float64Attribute{Computed: true, MarkdownDescription: "Latest template version."},
		"created_by":     schema.StringAttribute{Computed: true, MarkdownDescription: "Creator member ID."},
		"updated_by":     schema.StringAttribute{Computed: true, MarkdownDescription: "Last updater member ID."},
		"created_at":     schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp."},
		"updated_at":     schema.StringAttribute{Computed: true, MarkdownDescription: "Update timestamp."},
		"is_archived":    schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the template is archived."},
		"tags": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Version tags on the template.", NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"version_tag_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Version tag ID."},
			"name":           schema.StringAttribute{Computed: true, MarkdownDescription: "Tag name."},
		}}},
	}
}

func templateDoc(value *sigma.Template) templateDocModel {
	tags := make([]templateTagModel, 0, len(value.Tags))
	for _, tag := range value.Tags {
		tags = append(tags, templateTagModel{ID: types.StringValue(tag.VersionTagID), Name: types.StringValue(tag.Name)})
	}
	return templateDocModel{
		ID: types.StringValue(value.TemplateID), URLID: types.StringValue(value.TemplateURLID), Name: types.StringValue(value.Name),
		URL: types.StringValue(value.URL), Path: types.StringValue(value.Path), LatestVersion: types.Float64Value(value.LatestVersion),
		CreatedBy: types.StringValue(value.CreatedBy), UpdatedBy: types.StringValue(value.UpdatedBy),
		CreatedAt: types.StringValue(value.CreatedAt), UpdatedAt: types.StringValue(value.UpdatedAt), IsArchived: types.BoolValue(value.IsArchived),
		Tags: tags,
	}
}
