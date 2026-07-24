package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type documentDataSource struct {
	client *sigma.Client
	kind   string
}

func newDocumentDataSource(kind string) datasource.DataSource {
	return &documentDataSource{kind: kind}
}

func NewWorkbookDataSource() datasource.DataSource   { return newDocumentDataSource("workbook") }
func NewWorkbooksDataSource() datasource.DataSource  { return newDocumentDataSource("workbooks") }
func NewReportDataSource() datasource.DataSource     { return newDocumentDataSource("report") }
func NewReportsDataSource() datasource.DataSource    { return newDocumentDataSource("reports") }
func NewDataModelDataSource() datasource.DataSource  { return newDocumentDataSource("data_model") }
func NewDataModelsDataSource() datasource.DataSource { return newDocumentDataSource("data_models") }
func NewDatasetDataSource() datasource.DataSource    { return newDocumentDataSource("dataset") }
func NewDatasetsDataSource() datasource.DataSource   { return newDocumentDataSource("datasets") }
func NewTemplatesDataSource() datasource.DataSource  { return newDocumentDataSource("templates") }
func NewTagsDataSource() datasource.DataSource       { return newDocumentDataSource("tags") }

func (d *documentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.kind
}

func (d *documentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sigma.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected data source configuration type", "The Sigma provider returned an unexpected client type.")
		return
	}
	d.client = client
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

func reportDataAttributes(requireID bool) map[string]schema.Attribute {
	id := schema.StringAttribute{Computed: true, MarkdownDescription: "Report ID."}
	if requireID {
		id = schema.StringAttribute{Required: true, MarkdownDescription: "Report ID."}
	}
	return map[string]schema.Attribute{
		"id":             id,
		"url_id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Report URL ID."},
		"name":           schema.StringAttribute{Computed: true, MarkdownDescription: "Report name."},
		"url":            schema.StringAttribute{Computed: true, MarkdownDescription: "Report URL."},
		"path":           schema.StringAttribute{Computed: true, MarkdownDescription: "Report path."},
		"latest_version": schema.Float64Attribute{Computed: true, MarkdownDescription: "Latest report version."},
		"owner_id":       schema.StringAttribute{Computed: true, MarkdownDescription: "Owner member ID."},
		"created_by":     schema.StringAttribute{Computed: true, MarkdownDescription: "Creator member ID."},
		"updated_by":     schema.StringAttribute{Computed: true, MarkdownDescription: "Last updater member ID."},
		"created_at":     schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp."},
		"updated_at":     schema.StringAttribute{Computed: true, MarkdownDescription: "Update timestamp."},
		"is_archived":    schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the report is archived."},
	}
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
	}
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

func (d *documentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	const datasetDeprecation = "Sigma datasets are deprecated; prefer data models."
	switch d.kind {
	case "workbook":
		resp.Schema = schema.Schema{MarkdownDescription: "Retrieves a Sigma workbook by ID.", Attributes: workbookDataAttributes(true)}
	case "workbooks":
		resp.Schema = schema.Schema{MarkdownDescription: "Lists Sigma workbooks.", Attributes: map[string]schema.Attribute{
			"id":        schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
			"workbooks": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Workbooks.", NestedObject: schema.NestedAttributeObject{Attributes: workbookDataAttributes(false)}},
		}}
	case "report":
		resp.Schema = schema.Schema{MarkdownDescription: "Retrieves a Sigma report by ID.", Attributes: reportDataAttributes(true)}
	case "reports":
		resp.Schema = schema.Schema{MarkdownDescription: "Lists Sigma reports.", Attributes: map[string]schema.Attribute{
			"id":      schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
			"reports": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Reports.", NestedObject: schema.NestedAttributeObject{Attributes: reportDataAttributes(false)}},
		}}
	case "data_model":
		resp.Schema = schema.Schema{MarkdownDescription: "Retrieves a Sigma data model by ID.", Attributes: dataModelDataAttributes(true)}
	case "data_models":
		resp.Schema = schema.Schema{MarkdownDescription: "Lists Sigma data models.", Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
			"data_models": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Data models.", NestedObject: schema.NestedAttributeObject{Attributes: dataModelDataAttributes(false)}},
		}}
	case "dataset":
		resp.Schema = schema.Schema{
			MarkdownDescription: "Retrieves a Sigma dataset by ID.",
			DeprecationMessage:  datasetDeprecation,
			Attributes:          datasetDataAttributes(true),
		}
	case "datasets":
		resp.Schema = schema.Schema{
			MarkdownDescription: "Lists Sigma datasets.",
			DeprecationMessage:  datasetDeprecation,
			Attributes: map[string]schema.Attribute{
				"id":       schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
				"datasets": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Datasets.", NestedObject: schema.NestedAttributeObject{Attributes: datasetDataAttributes(false)}},
			},
		}
	case "templates":
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
	case "tags":
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
}

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

type workbooksDocModel struct {
	ID        types.String       `tfsdk:"id"`
	Workbooks []workbookDocModel `tfsdk:"workbooks"`
}

type reportDocModel struct {
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

type reportsDocModel struct {
	ID      types.String     `tfsdk:"id"`
	Reports []reportDocModel `tfsdk:"reports"`
}

type dataModelDocModel struct {
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

type dataModelsDocModel struct {
	ID         types.String        `tfsdk:"id"`
	DataModels []dataModelDocModel `tfsdk:"data_models"`
}

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

type datasetsDocModel struct {
	ID       types.String      `tfsdk:"id"`
	Datasets []datasetDocModel `tfsdk:"datasets"`
}

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

func workbookDoc(value *sigma.Workbook) workbookDocModel {
	return workbookDocModel{
		ID: types.StringValue(value.WorkbookID), URLID: types.StringValue(value.WorkbookURLID), Name: types.StringValue(value.Name),
		URL: types.StringValue(value.URL), Path: types.StringValue(value.Path), LatestVersion: types.Float64Value(value.LatestVersion),
		OwnerID: types.StringValue(value.OwnerID), CreatedBy: types.StringValue(value.CreatedBy), UpdatedBy: types.StringValue(value.UpdatedBy),
		CreatedAt: types.StringValue(value.CreatedAt), UpdatedAt: types.StringValue(value.UpdatedAt), IsArchived: types.BoolValue(value.IsArchived),
	}
}

func reportDoc(value *sigma.Report) reportDocModel {
	return reportDocModel{
		ID: types.StringValue(value.ReportID), URLID: types.StringValue(value.ReportURLID), Name: types.StringValue(value.Name),
		URL: types.StringValue(value.URL), Path: types.StringValue(value.Path), LatestVersion: types.Float64Value(value.LatestVersion),
		OwnerID: types.StringValue(value.OwnerID), CreatedBy: types.StringValue(value.CreatedBy), UpdatedBy: types.StringValue(value.UpdatedBy),
		CreatedAt: types.StringValue(value.CreatedAt), UpdatedAt: types.StringValue(value.UpdatedAt), IsArchived: types.BoolValue(value.IsArchived),
	}
}

func dataModelDoc(value *sigma.DataModel) dataModelDocModel {
	return dataModelDocModel{
		ID: types.StringValue(value.DataModelID), URLID: types.StringValue(value.DataModelURLID), Name: types.StringValue(value.Name),
		URL: types.StringValue(value.URL), Path: types.StringValue(value.Path), LatestVersion: types.Float64Value(value.LatestVersion),
		OwnerID: types.StringValue(value.OwnerID), CreatedBy: types.StringValue(value.CreatedBy), UpdatedBy: types.StringValue(value.UpdatedBy),
		CreatedAt: types.StringValue(value.CreatedAt), UpdatedAt: types.StringValue(value.UpdatedAt), IsArchived: types.BoolValue(value.IsArchived),
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

func templateDoc(value *sigma.Template) templateDocModel {
	return templateDocModel{
		ID: types.StringValue(value.TemplateID), URLID: types.StringValue(value.TemplateURLID), Name: types.StringValue(value.Name),
		URL: types.StringValue(value.URL), Path: types.StringValue(value.Path), LatestVersion: types.Float64Value(value.LatestVersion),
		CreatedBy: types.StringValue(value.CreatedBy), UpdatedBy: types.StringValue(value.UpdatedBy),
		CreatedAt: types.StringValue(value.CreatedAt), UpdatedAt: types.StringValue(value.UpdatedAt), IsArchived: types.BoolValue(value.IsArchived),
	}
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

func (d *documentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	switch d.kind {
	case "workbook":
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
	case "workbooks":
		values, err := d.client.ListWorkbooks(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Unable to list Sigma workbooks", err.Error())
			return
		}
		state := workbooksDocModel{ID: types.StringValue("workbooks")}
		for i := range values {
			state.Workbooks = append(state.Workbooks, workbookDoc(&values[i]))
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	case "report":
		var state reportDocModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		value, err := d.client.GetReport(ctx, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Unable to read Sigma report", err.Error())
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, reportDoc(value))...)
	case "reports":
		values, err := d.client.ListReports(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Unable to list Sigma reports", err.Error())
			return
		}
		state := reportsDocModel{ID: types.StringValue("reports")}
		for i := range values {
			state.Reports = append(state.Reports, reportDoc(&values[i]))
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	case "data_model":
		var state dataModelDocModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		value, err := d.client.GetDataModel(ctx, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Unable to read Sigma data model", err.Error())
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, dataModelDoc(value))...)
	case "data_models":
		values, err := d.client.ListDataModels(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Unable to list Sigma data models", err.Error())
			return
		}
		state := dataModelsDocModel{ID: types.StringValue("data_models")}
		for i := range values {
			state.DataModels = append(state.DataModels, dataModelDoc(&values[i]))
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	case "dataset":
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
	case "datasets":
		values, err := d.client.ListDatasets(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Unable to list Sigma datasets", err.Error())
			return
		}
		state := datasetsDocModel{ID: types.StringValue("datasets")}
		for i := range values {
			state.Datasets = append(state.Datasets, datasetDoc(&values[i]))
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	case "templates":
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
	case "tags":
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
}
