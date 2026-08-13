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

type templateDocModel struct {
	ID            types.String  `tfsdk:"id"`
	URLID         types.String  `tfsdk:"url_id"`
	Name          types.String  `tfsdk:"name"`
	URL           types.String  `tfsdk:"url"`
	Path          types.String  `tfsdk:"path"`
	LatestVersion types.Float64 `tfsdk:"latest_version"`
	CreatedBy     types.String  `tfsdk:"created_by"`
	UpdatedBy     types.String  `tfsdk:"updated_by"`
	CreatedAt     types.String  `tfsdk:"created_at"`
	UpdatedAt     types.String  `tfsdk:"updated_at"`
	IsArchived    types.Bool    `tfsdk:"is_archived"`
}

type templatesDocModel struct {
	ID        types.String       `tfsdk:"id"`
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
	resp.Schema = schema.Schema{MarkdownDescription: "Lists Sigma templates.", Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
		"templates": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Templates.", NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"id":             schema.StringAttribute{Computed: true, MarkdownDescription: "Template ID."},
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
		}}},
	}}
}
func (d *templatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	values, err := d.client.ListTemplates(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma templates", err.Error())
		return
	}
	state := templatesDocModel{ID: types.StringValue("templates")}
	for i := range values {
		state.Templates = append(state.Templates, templateDoc(&values[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func templateDoc(value *sigma.Template) templateDocModel {
	return templateDocModel{
		ID: types.StringValue(value.TemplateID), URLID: types.StringValue(value.TemplateURLID), Name: types.StringValue(value.Name),
		URL: types.StringValue(value.URL), Path: types.StringValue(value.Path), LatestVersion: types.Float64Value(value.LatestVersion),
		CreatedBy: types.StringValue(value.CreatedBy), UpdatedBy: types.StringValue(value.UpdatedBy),
		CreatedAt: types.StringValue(value.CreatedAt), UpdatedAt: types.StringValue(value.UpdatedAt), IsArchived: types.BoolValue(value.IsArchived),
	}
}
