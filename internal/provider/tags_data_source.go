package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*tagsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*tagsDataSource)(nil)
)

type tagsDataSource struct{ configuredDataSource }

type tagDocModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Color       types.String `tfsdk:"color"`
	Description types.String `tfsdk:"description"`
	OwnerID     types.String `tfsdk:"owner_id"`
	IsArchived  types.Bool   `tfsdk:"is_archived"`
	CreatedBy   types.String `tfsdk:"created_by"`
	UpdatedBy   types.String `tfsdk:"updated_by"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

type tagsDocModel struct {
	ID   types.String  `tfsdk:"id"`
	Tags []tagDocModel `tfsdk:"tags"`
}

func NewTagsDataSource() datasource.DataSource { return &tagsDataSource{} }

func (d *tagsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tags"
}
func (d *tagsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}
func (d *tagsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Lists Sigma version tags.", Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
		"tags": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Version tags.", NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Version tag ID."},
			"name":        schema.StringAttribute{Computed: true, MarkdownDescription: "Tag name."},
			"color":       schema.StringAttribute{Computed: true, MarkdownDescription: "Tag color."},
			"description": schema.StringAttribute{Computed: true, MarkdownDescription: "Tag description."},
			"owner_id":    schema.StringAttribute{Computed: true, MarkdownDescription: "Owner member ID."},
			"is_archived": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the tag is archived."},
			"created_by":  schema.StringAttribute{Computed: true, MarkdownDescription: "Creator member ID."},
			"updated_by":  schema.StringAttribute{Computed: true, MarkdownDescription: "Last updater member ID."},
			"created_at":  schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp."},
			"updated_at":  schema.StringAttribute{Computed: true, MarkdownDescription: "Update timestamp."},
		}}},
	}}
}
func (d *tagsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	values, err := d.client.ListTags(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma tags", err.Error())
		return
	}
	state := tagsDocModel{ID: types.StringValue("tags")}
	for i := range values {
		state.Tags = append(state.Tags, tagDoc(&values[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func tagDoc(value *sigma.Tag) tagDocModel {
	state := tagDocModel{
		ID: types.StringValue(value.VersionTagID), Name: types.StringValue(value.Name), Color: types.StringValue(value.Color),
		OwnerID: types.StringValue(value.OwnerID), IsArchived: types.BoolValue(value.IsArchived), CreatedBy: types.StringValue(value.CreatedBy),
		UpdatedBy: types.StringValue(value.UpdatedBy), CreatedAt: types.StringValue(value.CreatedAt), UpdatedAt: types.StringValue(value.UpdatedAt),
		Description: types.StringNull(),
	}
	if value.Description != nil {
		state.Description = types.StringValue(*value.Description)
	}
	return state
}
