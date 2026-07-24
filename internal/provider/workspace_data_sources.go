package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type workspaceDataSource struct {
	client *sigma.Client
	kind   string
}

func NewWorkspaceDataSource() datasource.DataSource  { return &workspaceDataSource{kind: "workspace"} }
func NewWorkspacesDataSource() datasource.DataSource { return &workspaceDataSource{kind: "workspaces"} }
func NewFilesDataSource() datasource.DataSource      { return &workspaceDataSource{kind: "files"} }

func (d *workspaceDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_" + d.kind
}

func (d *workspaceDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}
	client, ok := request.ProviderData.(*sigma.Client)
	if !ok {
		response.Diagnostics.AddError("Unexpected data source configuration type", "The Sigma provider returned an unexpected client type.")
		return
	}
	d.client = client
}

func (d *workspaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	switch d.kind {
	case "workspace":
		response.Schema = schema.Schema{
			MarkdownDescription: "Retrieves a Sigma workspace by ID.",
			Attributes:          workspaceDataAttributes(true),
		}
	case "workspaces":
		response.Schema = schema.Schema{
			MarkdownDescription: "Lists all Sigma workspaces using the paginated v2.1 endpoint.",
			Attributes: map[string]schema.Attribute{
				"id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
				"workspaces": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Workspaces.", NestedObject: schema.NestedAttributeObject{Attributes: workspaceDataAttributes(false)}},
			},
		}
	case "files":
		response.Schema = schema.Schema{
			MarkdownDescription: "Lists Sigma files with optional API filters.",
			Attributes: map[string]schema.Attribute{
				"id":                   schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
				"name":                 schema.StringAttribute{Optional: true, MarkdownDescription: "File name filter."},
				"permission":           schema.StringAttribute{Optional: true, MarkdownDescription: "Permission filter: `view`, `explore`, `organize`, or `edit`."},
				"type_filters":         schema.SetAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "File type filters supported by the Sigma files API."},
				"parent_id":            schema.StringAttribute{Optional: true, MarkdownDescription: "Parent folder ID."},
				"direct_children_only": schema.BoolAttribute{Optional: true, MarkdownDescription: "Whether to return only direct children of the parent."},
				"files":                schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Files.", NestedObject: schema.NestedAttributeObject{Attributes: fileDataAttributes()}},
			},
		}
	}
}

func workspaceDataAttributes(requireID bool) map[string]schema.Attribute {
	id := schema.StringAttribute{Computed: true, MarkdownDescription: "Workspace ID."}
	if requireID {
		id = schema.StringAttribute{Required: true, MarkdownDescription: "Workspace ID."}
	}
	return map[string]schema.Attribute{
		"id":         id,
		"url_id":     schema.StringAttribute{Computed: true, MarkdownDescription: "Base62 identifier used in Sigma workspace URLs."},
		"name":       schema.StringAttribute{Computed: true, MarkdownDescription: "Workspace name."},
		"created_by": schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the member who created the workspace."},
		"updated_by": schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the member who last updated the workspace."},
		"created_at": schema.StringAttribute{Computed: true, MarkdownDescription: "Workspace creation timestamp."},
		"updated_at": schema.StringAttribute{Computed: true, MarkdownDescription: "Workspace update timestamp."},
	}
}

func fileDataAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id":                   schema.StringAttribute{Computed: true, MarkdownDescription: "Inode ID."},
		"url_id":               schema.StringAttribute{Computed: true, MarkdownDescription: "Identifier used in Sigma URLs."},
		"name":                 schema.StringAttribute{Computed: true, MarkdownDescription: "File name."},
		"type":                 schema.StringAttribute{Computed: true, MarkdownDescription: "File type."},
		"parent_id":            schema.StringAttribute{Computed: true, MarkdownDescription: "Parent folder ID."},
		"parent_url_id":        schema.StringAttribute{Computed: true, MarkdownDescription: "Parent identifier used in Sigma URLs."},
		"permission":           schema.StringAttribute{Computed: true, MarkdownDescription: "Effective permission."},
		"path":                 schema.StringAttribute{Computed: true, MarkdownDescription: "File path."},
		"badge":                schema.StringAttribute{Computed: true, MarkdownDescription: "Badge applied to the file."},
		"is_archived":          schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the file is archived."},
		"description":          schema.StringAttribute{Computed: true, MarkdownDescription: "File description."},
		"owner_id":             schema.StringAttribute{Computed: true, MarkdownDescription: "Member ID of the owner."},
		"parent_source_url_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Source document URL ID for a tenant deployment."},
		"created_by":           schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the member who created the file."},
		"updated_by":           schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the member who last updated the file."},
		"created_at":           schema.StringAttribute{Computed: true, MarkdownDescription: "File creation timestamp."},
		"updated_at":           schema.StringAttribute{Computed: true, MarkdownDescription: "File update timestamp."},
	}
}

type workspaceDataModel struct {
	ID        types.String `tfsdk:"id"`
	URLID     types.String `tfsdk:"url_id"`
	Name      types.String `tfsdk:"name"`
	CreatedBy types.String `tfsdk:"created_by"`
	UpdatedBy types.String `tfsdk:"updated_by"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

type workspacesDataModel struct {
	ID         types.String         `tfsdk:"id"`
	Workspaces []workspaceDataModel `tfsdk:"workspaces"`
}

type fileDataModel struct {
	ID                types.String `tfsdk:"id"`
	URLID             types.String `tfsdk:"url_id"`
	Name              types.String `tfsdk:"name"`
	Type              types.String `tfsdk:"type"`
	ParentID          types.String `tfsdk:"parent_id"`
	ParentURLID       types.String `tfsdk:"parent_url_id"`
	Permission        types.String `tfsdk:"permission"`
	Path              types.String `tfsdk:"path"`
	Badge             types.String `tfsdk:"badge"`
	IsArchived        types.Bool   `tfsdk:"is_archived"`
	Description       types.String `tfsdk:"description"`
	OwnerID           types.String `tfsdk:"owner_id"`
	ParentSourceURLID types.String `tfsdk:"parent_source_url_id"`
	CreatedBy         types.String `tfsdk:"created_by"`
	UpdatedBy         types.String `tfsdk:"updated_by"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
}

type filesDataModel struct {
	ID                 types.String    `tfsdk:"id"`
	Name               types.String    `tfsdk:"name"`
	Permission         types.String    `tfsdk:"permission"`
	TypeFilters        types.Set       `tfsdk:"type_filters"`
	ParentID           types.String    `tfsdk:"parent_id"`
	DirectChildrenOnly types.Bool      `tfsdk:"direct_children_only"`
	Files              []fileDataModel `tfsdk:"files"`
}

func (d *workspaceDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	switch d.kind {
	case "workspace":
		var state workspaceDataModel
		response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
		if response.Diagnostics.HasError() {
			return
		}
		value, err := d.client.GetWorkspace(ctx, state.ID.ValueString())
		if err != nil {
			response.Diagnostics.AddError("Unable to read Sigma workspace", err.Error())
			return
		}
		state = workspaceData(value)
		response.Diagnostics.Append(response.State.Set(ctx, &state)...)
	case "workspaces":
		values, err := d.client.ListWorkspaces(ctx)
		if err != nil {
			response.Diagnostics.AddError("Unable to list Sigma workspaces", err.Error())
			return
		}
		state := workspacesDataModel{ID: types.StringValue("workspaces")}
		for i := range values {
			state.Workspaces = append(state.Workspaces, workspaceData(&values[i]))
		}
		response.Diagnostics.Append(response.State.Set(ctx, &state)...)
	case "files":
		var state filesDataModel
		response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
		if response.Diagnostics.HasError() {
			return
		}
		var typeFilters []string
		response.Diagnostics.Append(state.TypeFilters.ElementsAs(ctx, &typeFilters, false)...)
		var direct *bool
		if !state.DirectChildrenOnly.IsNull() && !state.DirectChildrenOnly.IsUnknown() {
			value := state.DirectChildrenOnly.ValueBool()
			direct = &value
		}
		values, err := d.client.ListFiles(ctx, sigma.ListFilesOptions{
			Name: state.Name.ValueString(), Permission: state.Permission.ValueString(), TypeFilters: typeFilters,
			ParentID: state.ParentID.ValueString(), DirectChildrenOnly: direct,
		})
		if err != nil {
			response.Diagnostics.AddError("Unable to list Sigma files", err.Error())
			return
		}
		state.ID = types.StringValue("files")
		for i := range values {
			state.Files = append(state.Files, fileData(&values[i]))
		}
		response.Diagnostics.Append(response.State.Set(ctx, &state)...)
	}
}

func workspaceData(value *sigma.Workspace) workspaceDataModel {
	return workspaceDataModel{
		ID: types.StringValue(value.WorkspaceID), URLID: types.StringValue(value.WorkspaceURLID), Name: types.StringValue(value.Name),
		CreatedBy: types.StringValue(value.CreatedBy), UpdatedBy: types.StringValue(value.UpdatedBy),
		CreatedAt: types.StringValue(value.CreatedAt), UpdatedAt: types.StringValue(value.UpdatedAt),
	}
}

func fileData(value *sigma.File) fileDataModel {
	return fileDataModel{
		ID: types.StringValue(value.ID), URLID: types.StringValue(value.URLID), Name: types.StringValue(value.Name), Type: types.StringValue(value.Type),
		ParentID: types.StringValue(value.ParentID), ParentURLID: types.StringValue(value.ParentURLID), Permission: types.StringValue(value.Permission),
		Path: types.StringValue(value.Path), Badge: nullableString(value.Badge), IsArchived: types.BoolValue(value.IsArchived),
		Description: types.StringValue(value.Description), OwnerID: nullableString(value.OwnerID), ParentSourceURLID: types.StringValue(value.ParentSourceURLID),
		CreatedBy: types.StringValue(value.CreatedBy), UpdatedBy: types.StringValue(value.UpdatedBy),
		CreatedAt: types.StringValue(value.CreatedAt), UpdatedAt: types.StringValue(value.UpdatedAt),
	}
}
