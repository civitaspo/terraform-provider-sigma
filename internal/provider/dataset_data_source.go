package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*datasetDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*datasetDataSource)(nil)
)

const datasetDeprecation = "Sigma datasets are deprecated; prefer data models."

type datasetDataSource struct{ configuredDataSource }

type datasetDocModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	URL         types.String `tfsdk:"url"`
	Path        types.String `tfsdk:"path"`
	Owner       types.String `tfsdk:"owner"`
	CreatedBy   types.String `tfsdk:"created_by"`
	UpdatedBy   types.String `tfsdk:"updated_by"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	IsArchived  types.Bool   `tfsdk:"is_archived"`
}

func NewDatasetDataSource() datasource.DataSource { return &datasetDataSource{} }

func (d *datasetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataset"
}
func (d *datasetDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}
func (d *datasetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a Sigma dataset by ID.",
		DeprecationMessage:  datasetDeprecation,
		Attributes:          datasetDataAttributes(true),
	}
}
func (d *datasetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datasetDocModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := d.client.GetDataset(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma dataset", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, datasetDoc(value))...)
}

func datasetDataAttributes(requireID bool) map[string]schema.Attribute {
	id := schema.StringAttribute{Computed: true, MarkdownDescription: "Dataset ID."}
	if requireID {
		id = schema.StringAttribute{Required: true, MarkdownDescription: "Dataset ID."}
	}
	return map[string]schema.Attribute{
		"id":          id,
		"name":        schema.StringAttribute{Computed: true, MarkdownDescription: "Dataset name."},
		"description": schema.StringAttribute{Computed: true, MarkdownDescription: "Dataset description."},
		"url":         schema.StringAttribute{Computed: true, MarkdownDescription: "Dataset URL."},
		"path":        schema.StringAttribute{Computed: true, MarkdownDescription: "Dataset path."},
		"owner":       schema.StringAttribute{Computed: true, MarkdownDescription: "Owner identifier."},
		"created_by":  schema.StringAttribute{Computed: true, MarkdownDescription: "Creator member ID."},
		"updated_by":  schema.StringAttribute{Computed: true, MarkdownDescription: "Last updater member ID."},
		"created_at":  schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp."},
		"updated_at":  schema.StringAttribute{Computed: true, MarkdownDescription: "Update timestamp."},
		"is_archived": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the dataset is archived."},
	}
}

func datasetDoc(value *sigma.Dataset) datasetDocModel {
	state := datasetDocModel{
		ID: types.StringValue(value.DatasetID), Name: types.StringValue(value.Name), URL: types.StringValue(value.URL),
		Path: types.StringValue(value.Path), Owner: types.StringValue(value.Owner), CreatedBy: types.StringValue(value.CreatedBy),
		UpdatedBy: types.StringValue(value.UpdatedBy), CreatedAt: types.StringValue(value.CreatedAt), UpdatedAt: types.StringValue(value.UpdatedAt),
		IsArchived: types.BoolValue(value.IsArchived), Description: types.StringNull(),
	}
	if value.Description != nil {
		state.Description = types.StringValue(*value.Description)
	}
	return state
}
