package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type workspaceResource struct{ configuredResource }

type workspaceModel struct {
	ID           types.String `tfsdk:"id"`
	URLID        types.String `tfsdk:"url_id"`
	Name         types.String `tfsdk:"name"`
	NoDuplicates types.Bool   `tfsdk:"no_duplicates"`
	CreatedBy    types.String `tfsdk:"created_by"`
	UpdatedBy    types.String `tfsdk:"updated_by"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

func NewWorkspaceResource() resource.Resource { return &workspaceResource{} }

func (r *workspaceResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_workspace"
}

func (r *workspaceResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.configure(request, response)
}

func (r *workspaceResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Manages a Sigma workspace.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true, MarkdownDescription: "Workspace ID."},
			"url_id":        schema.StringAttribute{Computed: true, MarkdownDescription: "Base62 identifier used in Sigma workspace URLs."},
			"name":          schema.StringAttribute{Required: true, MarkdownDescription: "Workspace name."},
			"no_duplicates": schema.BoolAttribute{Optional: true, MarkdownDescription: "Whether Sigma should reject duplicate workspace names."},
			"created_by":    schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the member who created the workspace."},
			"updated_by":    schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the member who last updated the workspace."},
			"created_at":    schema.StringAttribute{Computed: true, MarkdownDescription: "Workspace creation timestamp."},
			"updated_at":    schema.StringAttribute{Computed: true, MarkdownDescription: "Workspace update timestamp."},
		},
	}
}

func (r *workspaceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan workspaceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	value, err := r.client.CreateWorkspace(ctx, sigma.CreateWorkspaceInput{
		Name: plan.Name.ValueString(), NoDuplicates: plan.NoDuplicates.ValueBool(),
	})
	if err != nil {
		response.Diagnostics.AddError("Unable to create Sigma workspace", err.Error())
		return
	}
	setWorkspace(&plan, value)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *workspaceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state workspaceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	value, err := r.client.GetWorkspace(ctx, state.ID.ValueString())
	if sigma.IsNotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError("Unable to read Sigma workspace", err.Error())
		return
	}
	setWorkspace(&state, value)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *workspaceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan workspaceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	value, err := r.client.UpdateWorkspace(ctx, plan.ID.ValueString(), sigma.UpdateWorkspaceInput{
		Name: plan.Name.ValueString(), NoDuplicates: plan.NoDuplicates.ValueBool(),
	})
	if err != nil {
		response.Diagnostics.AddError("Unable to update Sigma workspace", err.Error())
		return
	}
	setWorkspace(&plan, value)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *workspaceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state workspaceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if !response.Diagnostics.HasError() {
		if err := r.client.DeleteWorkspace(ctx, state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
			response.Diagnostics.AddError("Unable to delete Sigma workspace", err.Error())
		}
	}
}

func (r *workspaceResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	importPassthrough(ctx, request, response)
}

func setWorkspace(state *workspaceModel, value *sigma.Workspace) {
	state.ID = types.StringValue(value.WorkspaceID)
	state.URLID = types.StringValue(value.WorkspaceURLID)
	state.Name = types.StringValue(value.Name)
	state.CreatedBy = types.StringValue(value.CreatedBy)
	state.UpdatedBy = types.StringValue(value.UpdatedBy)
	state.CreatedAt = types.StringValue(value.CreatedAt)
	state.UpdatedAt = types.StringValue(value.UpdatedAt)
}

type fileResource struct{ configuredResource }

type fileModel struct {
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
	SourceInodeID     types.String `tfsdk:"source_inode_id"`
	SourceVersion     types.Int64  `tfsdk:"source_version"`
	Restore           types.Bool   `tfsdk:"restore"`
}

func NewFileResource() resource.Resource { return &fileResource{} }

func (r *fileResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_file"
}

func (r *fileResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.configure(request, response)
}

func (r *fileResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Manages an empty Sigma workspace, folder, workbook, or report through the files API. Workbooks may copy an existing inode version via `source_inode_id` and `source_version`. Set `restore` on update to unarchive a deleted file.",
		Attributes: map[string]schema.Attribute{
			"id":                   schema.StringAttribute{Computed: true, MarkdownDescription: "Inode ID."},
			"url_id":               schema.StringAttribute{Computed: true, MarkdownDescription: "Identifier used in Sigma URLs."},
			"name":                 schema.StringAttribute{Required: true, MarkdownDescription: "File name."},
			"type":                 schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "Creatable file type: `workspace`, `folder`, `workbook`, or `report`."},
			"parent_id":            schema.StringAttribute{Optional: true, MarkdownDescription: "Parent folder ID."},
			"parent_url_id":        schema.StringAttribute{Computed: true, MarkdownDescription: "Parent identifier used in Sigma URLs."},
			"permission":           schema.StringAttribute{Computed: true, MarkdownDescription: "Effective permission."},
			"path":                 schema.StringAttribute{Computed: true, MarkdownDescription: "File path."},
			"badge":                schema.StringAttribute{Computed: true, MarkdownDescription: "Badge applied to the file."},
			"is_archived":          schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the file is archived."},
			"description":          schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "File description."},
			"owner_id":             schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Member ID of the file owner."},
			"parent_source_url_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Source document URL ID for a tenant deployment."},
			"created_by":           schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the member who created the file."},
			"updated_by":           schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the member who last updated the file."},
			"created_at":           schema.StringAttribute{Computed: true, MarkdownDescription: "File creation timestamp."},
			"updated_at":           schema.StringAttribute{Computed: true, MarkdownDescription: "File update timestamp."},
			"source_inode_id": schema.StringAttribute{
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "Workbook source inode ID (`source.inodeId`). Only valid when `type` is `workbook`. Must be set together with `source_version`. Changing this value forces replacement.",
			},
			"source_version": schema.Int64Attribute{
				Optional:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				MarkdownDescription: "Workbook source version (`source.version`). Only valid when `type` is `workbook`. Must be set together with `source_inode_id`. Changing this value forces replacement.",
			},
			"restore": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "When true, PATCH `restore=true` to unarchive a deleted folder or document. The files API only accepts this on update.",
			},
		},
	}
}

func (r *fileResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var config fileModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() || config.Type.IsNull() || config.Type.IsUnknown() {
		return
	}
	if !contains([]string{"workspace", "folder", "workbook", "report"}, config.Type.ValueString()) {
		response.Diagnostics.AddAttributeError(path.Root("type"), "Invalid Sigma file type", "The type must be `workspace`, `folder`, `workbook`, or `report`.")
	}
	sourceSet := !config.SourceInodeID.IsNull() && !config.SourceInodeID.IsUnknown()
	versionSet := !config.SourceVersion.IsNull() && !config.SourceVersion.IsUnknown()
	if sourceSet != versionSet {
		response.Diagnostics.AddError("Invalid Sigma file source", "`source_inode_id` and `source_version` must be set together.")
	}
	if (sourceSet || versionSet) && config.Type.ValueString() != "workbook" {
		response.Diagnostics.AddError("Invalid Sigma file source", "`source_inode_id` and `source_version` are only valid when type is `workbook`.")
	}
}

func (r *fileResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan fileModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	if !contains([]string{"workspace", "folder", "workbook", "report"}, plan.Type.ValueString()) {
		response.Diagnostics.AddError("Invalid Sigma file type", "The type must be `workspace`, `folder`, `workbook`, or `report`.")
		return
	}
	input := sigma.CreateFileInput{
		Type: plan.Type.ValueString(), Name: plan.Name.ValueString(), Description: plan.Description.ValueString(),
		OwnerID: plan.OwnerID.ValueString(), ParentID: plan.ParentID.ValueString(),
	}
	if !plan.SourceInodeID.IsNull() && !plan.SourceInodeID.IsUnknown() && !plan.SourceVersion.IsNull() && !plan.SourceVersion.IsUnknown() {
		input.Source = &sigma.FileSourceInput{InodeID: plan.SourceInodeID.ValueString(), Version: plan.SourceVersion.ValueInt64()}
	}
	value, err := r.client.CreateFile(ctx, input)
	if err != nil {
		response.Diagnostics.AddError("Unable to create Sigma file", err.Error())
		return
	}
	setFile(&plan, value)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *fileResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state fileModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	value, err := r.client.GetFile(ctx, state.ID.ValueString())
	if sigma.IsNotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError("Unable to read Sigma file", err.Error())
		return
	}
	setFile(&state, value)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *fileResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan fileModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	var state fileModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	name, description := plan.Name.ValueString(), plan.Description.ValueString()
	input := sigma.UpdateFileInput{Name: &name, Description: &description}
	if !plan.OwnerID.IsNull() && !plan.OwnerID.IsUnknown() {
		ownerID := plan.OwnerID.ValueString()
		input.OwnerID = &ownerID
	}
	if !plan.ParentID.IsNull() && !plan.ParentID.IsUnknown() {
		parentID := plan.ParentID.ValueString()
		input.ParentID = &parentID
	}
	if !plan.Restore.IsNull() && !plan.Restore.IsUnknown() {
		restore := plan.Restore.ValueBool()
		input.Restore = &restore
	}
	value, err := r.client.UpdateFile(ctx, state.ID.ValueString(), input)
	if err != nil {
		response.Diagnostics.AddError("Unable to update Sigma file", err.Error())
		return
	}
	setFile(&plan, value)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *fileResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state fileModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if !response.Diagnostics.HasError() {
		if err := r.client.DeleteFile(ctx, state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
			response.Diagnostics.AddError("Unable to delete Sigma file", err.Error())
		}
	}
}

func (r *fileResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	importPassthrough(ctx, request, response)
}

func setFile(state *fileModel, value *sigma.File) {
	state.ID = types.StringValue(value.ID)
	state.URLID = types.StringValue(value.URLID)
	state.Name = types.StringValue(value.Name)
	state.Type = types.StringValue(value.Type)
	state.ParentID = types.StringValue(value.ParentID)
	state.ParentURLID = types.StringValue(value.ParentURLID)
	state.Permission = types.StringValue(value.Permission)
	state.Path = types.StringValue(value.Path)
	state.Badge = nullableString(value.Badge)
	state.IsArchived = types.BoolValue(value.IsArchived)
	state.Description = types.StringValue(value.Description)
	state.OwnerID = nullableString(value.OwnerID)
	state.ParentSourceURLID = types.StringValue(value.ParentSourceURLID)
	state.CreatedBy = types.StringValue(value.CreatedBy)
	state.UpdatedBy = types.StringValue(value.UpdatedBy)
	state.CreatedAt = types.StringValue(value.CreatedAt)
	state.UpdatedAt = types.StringValue(value.UpdatedAt)
}

type grantResource struct {
	configuredResource
	kind string
}

type grantModel struct {
	ID             types.String `tfsdk:"id"`
	InodeID        types.String `tfsdk:"inode_id"`
	MemberID       types.String `tfsdk:"member_id"`
	TeamID         types.String `tfsdk:"team_id"`
	Permission     types.String `tfsdk:"permission"`
	TagID          types.String `tfsdk:"tag_id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	InodeType      types.String `tfsdk:"inode_type"`
	CreatedBy      types.String `tfsdk:"created_by"`
	UpdatedBy      types.String `tfsdk:"updated_by"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

func NewGrantResource() resource.Resource          { return &grantResource{kind: "generic"} }
func NewWorkspaceGrantResource() resource.Resource { return &grantResource{kind: "workspace"} }
func NewWorkbookGrantResource() resource.Resource  { return &grantResource{kind: "workbooks"} }
func NewReportGrantResource() resource.Resource    { return &grantResource{kind: "reports"} }

func (r *grantResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	name := r.kind + "_grant"
	switch r.kind {
	case "generic":
		name = "grant"
	case "workbooks":
		name = "workbook_grant"
	case "reports":
		name = "report_grant"
	}
	response.TypeName = request.ProviderTypeName + "_" + name
}

func (r *grantResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.configure(request, response)
}

func (r *grantResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	inodeDescription := "Inode ID."
	switch r.kind {
	case "workspace":
		inodeDescription = "Workspace ID."
	case "workbooks":
		inodeDescription = "Workbook ID."
	case "reports":
		inodeDescription = "Report ID."
	}
	attributes := map[string]schema.Attribute{
		"id":              schema.StringAttribute{Computed: true, MarkdownDescription: "Grant ID."},
		"inode_id":        schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: inodeDescription},
		"member_id":       schema.StringAttribute{Optional: true, PlanModifiers: replace, MarkdownDescription: "Member ID receiving the grant. Exactly one of `member_id` or `team_id` must be set."},
		"team_id":         schema.StringAttribute{Optional: true, PlanModifiers: replace, MarkdownDescription: "Team ID receiving the grant. Exactly one of `member_id` or `team_id` must be set."},
		"permission":      schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: grantPermissionDescription(r.kind)},
		"tag_id":          schema.StringAttribute{Optional: true, PlanModifiers: replace, MarkdownDescription: "Optional version tag ID. Supported by generic, workbook, and report grants. Changing this forces a new resource. Tagged workbook/report grants are created through the generic grants API so Terraform receives a stable grant ID."},
		"organization_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Organization ID."},
		"inode_type":      schema.StringAttribute{Computed: true, MarkdownDescription: "Inode type."},
		"created_by":      schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the member who created the grant."},
		"updated_by":      schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the member who last updated the grant."},
		"created_at":      schema.StringAttribute{Computed: true, MarkdownDescription: "Grant creation timestamp."},
		"updated_at":      schema.StringAttribute{Computed: true, MarkdownDescription: "Grant update timestamp."},
	}
	response.Schema = schema.Schema{
		MarkdownDescription: grantResourceDescription(r.kind),
		Attributes:          attributes,
	}
}

func (r *grantResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var config grantModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}
	if config.MemberID.IsUnknown() || config.TeamID.IsUnknown() || config.Permission.IsUnknown() || config.TagID.IsUnknown() {
		return
	}
	r.validate(&config, &response.Diagnostics)
}

func grantResourceDescription(kind string) string {
	switch kind {
	case "workspace":
		return "Manages a fine-grained Sigma workspace grant."
	case "workbooks":
		return "Manages a Sigma workbook grant."
	case "reports":
		return "Manages a Sigma report grant."
	default:
		return "Manages a generic Sigma inode grant."
	}
}

func grantPermissionDescription(kind string) string {
	switch kind {
	case "workspace":
		return "Workspace permission: `view`, `explore`, `organize`, or `edit`."
	case "workbooks":
		return "Workbook permission: `view`, `explore`, or `edit`."
	case "reports":
		return "Report permission: `view` or `edit`."
	default:
		return "Permission allowed by the inode type: `admin`, `annotate`, `update`, `usage`, `writeback`, `create`, `organize`, `explore`, `view`, `edit`, or `apply`."
	}
}

func (r *grantResource) validate(plan *grantModel, response interface{ AddError(string, string) }) bool {
	memberSet := !plan.MemberID.IsNull() && !plan.MemberID.IsUnknown() && plan.MemberID.ValueString() != ""
	teamSet := !plan.TeamID.IsNull() && !plan.TeamID.IsUnknown() && plan.TeamID.ValueString() != ""
	if memberSet == teamSet {
		response.AddError("Invalid Sigma grant grantee", "Exactly one of `member_id` or `team_id` must be set.")
		return false
	}
	if r.kind == "workspace" && !plan.TagID.IsNull() && !plan.TagID.IsUnknown() && plan.TagID.ValueString() != "" {
		response.AddError("Invalid Sigma workspace grant", "`tag_id` is not supported for workspace grants.")
		return false
	}
	allowed := []string{"admin", "annotate", "update", "usage", "writeback", "create", "organize", "explore", "view", "edit", "apply"}
	switch r.kind {
	case "workspace":
		allowed = []string{"view", "explore", "organize", "edit"}
	case "workbooks":
		allowed = []string{"view", "explore", "edit"}
	case "reports":
		allowed = []string{"view", "edit"}
	}
	if !contains(allowed, plan.Permission.ValueString()) {
		response.AddError("Invalid Sigma grant permission", fmt.Sprintf("Permission %q is not valid for this resource.", plan.Permission.ValueString()))
		return false
	}
	return true
}

func (r *grantResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan grantModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() || !r.validate(&plan, &response.Diagnostics) {
		return
	}
	grantee := sigma.Grantee{MemberID: plan.MemberID.ValueString(), TeamID: plan.TeamID.ValueString()}
	tagID := plan.TagID.ValueString()
	var value *sigma.Grant
	var err error
	switch r.kind {
	case "generic":
		value, err = r.client.CreateGrant(ctx, sigma.CreateGrantInput{
			Grantee: grantee, Permission: plan.Permission.ValueString(), InodeID: plan.InodeID.ValueString(), TagID: tagID,
		})
	case "workspace":
		err = r.client.CreateWorkspaceGrant(ctx, plan.InodeID.ValueString(), grantee, plan.Permission.ValueString())
	case "workbooks", "reports":
		// Prefer the generic grants API when a version tag is set so we receive the
		// created grant ID. List responses do not reliably include tagId, so
		// post-create lookups by grantee+permission alone are ambiguous.
		if tagID != "" {
			value, err = r.client.CreateGrant(ctx, sigma.CreateGrantInput{
				Grantee: grantee, Permission: plan.Permission.ValueString(), InodeID: plan.InodeID.ValueString(), TagID: tagID,
			})
		} else {
			err = r.client.CreateDocumentGrant(ctx, r.kind, plan.InodeID.ValueString(), grantee, plan.Permission.ValueString(), tagID)
		}
	}
	if err != nil {
		response.Diagnostics.AddError("Unable to create Sigma grant", err.Error())
		return
	}
	if value == nil {
		value, err = r.findGrant(ctx, &plan, "")
		if err != nil {
			response.Diagnostics.AddError("Unable to locate created Sigma grant", err.Error())
			return
		}
	}
	setGrant(&plan, value)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *grantResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state grantModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	var value *sigma.Grant
	var err error
	if r.kind == "generic" {
		value, err = r.client.GetGrant(ctx, state.ID.ValueString())
	} else {
		value, err = r.findGrant(ctx, &state, state.ID.ValueString())
	}
	if sigma.IsNotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError("Unable to read Sigma grant", err.Error())
		return
	}
	setGrant(&state, value)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *grantResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {}

func (r *grantResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state grantModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	var err error
	switch r.kind {
	case "generic":
		err = r.client.DeleteGrant(ctx, state.ID.ValueString())
	case "workspace":
		err = r.client.DeleteWorkspaceGrant(ctx, state.InodeID.ValueString(), state.ID.ValueString())
	case "workbooks", "reports":
		err = r.client.DeleteDocumentGrant(ctx, r.kind, state.InodeID.ValueString(), state.ID.ValueString())
	}
	if err != nil && !sigma.IsNotFound(err) {
		response.Diagnostics.AddError("Unable to delete Sigma grant", err.Error())
	}
}

func (r *grantResource) findGrant(ctx context.Context, model *grantModel, grantID string) (*sigma.Grant, error) {
	var values []sigma.Grant
	var err error
	if r.kind == "workspace" {
		values, err = r.client.ListWorkspaceGrants(ctx, model.InodeID.ValueString())
	} else {
		values, err = r.client.ListGrants(ctx, model.InodeID.ValueString())
	}
	if err != nil {
		return nil, err
	}
	if grantID != "" {
		for i := range values {
			if values[i].GrantID == grantID {
				return &values[i], nil
			}
		}
		return nil, &sigma.APIError{StatusCode: 404, Message: "grant not found"}
	}
	var matches []sigma.Grant
	for i := range values {
		if grantMatches(&values[i], model) {
			matches = append(matches, values[i])
		}
	}
	switch len(matches) {
	case 1:
		return &matches[0], nil
	case 0:
		return nil, &sigma.APIError{StatusCode: 404, Message: "grant not found"}
	default:
		return nil, fmt.Errorf("multiple grants matched inode/grantee/permission; set a unique permission or import by grant ID (tag-scoped grants may not be distinguishable in list responses)")
	}
}

func grantMatches(value *sigma.Grant, model *grantModel) bool {
	if value.Permission != model.Permission.ValueString() {
		return false
	}
	desiredTag := ""
	if !model.TagID.IsNull() && !model.TagID.IsUnknown() {
		desiredTag = model.TagID.ValueString()
	}
	actualTag := ""
	if value.TagID != nil {
		actualTag = *value.TagID
	}
	// When the API returns tagId, require an exact match. When it omits tagId
	// (current documented list schema), only accept untagged desired grants so
	// tag-scoped grants are not selected by accident.
	if actualTag != "" || desiredTag != "" {
		if actualTag == "" && desiredTag != "" {
			return false
		}
		if actualTag != desiredTag {
			return false
		}
	}
	if !model.MemberID.IsNull() && model.MemberID.ValueString() != "" {
		return value.MemberID != nil && *value.MemberID == model.MemberID.ValueString()
	}
	return value.TeamID != nil && *value.TeamID == model.TeamID.ValueString()
}

func (r *grantResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	if r.kind == "generic" {
		importPassthrough(ctx, request, response)
		return
	}
	parts := strings.SplitN(request.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		response.Diagnostics.AddError("Invalid import ID", "Use `inodeId/grantId`.")
		return
	}
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("inode_id"), parts[0])...)
}

func setGrant(state *grantModel, value *sigma.Grant) {
	priorTag := state.TagID
	state.ID = types.StringValue(value.GrantID)
	state.InodeID = types.StringValue(value.InodeID)
	state.MemberID = nullableString(value.MemberID)
	state.TeamID = nullableString(value.TeamID)
	state.Permission = types.StringValue(value.Permission)
	if value.TagID != nil {
		state.TagID = types.StringValue(*value.TagID)
	} else if !priorTag.IsNull() && !priorTag.IsUnknown() {
		// List/get schemas may omit tagId; keep the configured tag identity in state.
		state.TagID = priorTag
	} else {
		state.TagID = types.StringNull()
	}
	state.OrganizationID = types.StringValue(value.OrganizationID)
	state.InodeType = types.StringValue(value.InodeType)
	state.CreatedBy = types.StringValue(value.CreatedBy)
	state.UpdatedBy = types.StringValue(value.UpdatedBy)
	state.CreatedAt = types.StringValue(value.CreatedAt)
	state.UpdatedAt = types.StringValue(value.UpdatedAt)
}

func nullableString(value *string) types.String {
	if value == nil {
		return types.StringNull()
	}
	return types.StringValue(*value)
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

var (
	_ resource.ResourceWithImportState    = (*workspaceResource)(nil)
	_ resource.ResourceWithImportState    = (*fileResource)(nil)
	_ resource.ResourceWithImportState    = (*grantResource)(nil)
	_ resource.ResourceWithValidateConfig = (*fileResource)(nil)
	_ resource.ResourceWithValidateConfig = (*grantResource)(nil)
)
