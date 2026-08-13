package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type fileResource struct{ configuredResource }

var (
	_ resource.Resource                   = (*fileResource)(nil)
	_ resource.ResourceWithConfigure      = (*fileResource)(nil)
	_ resource.ResourceWithImportState    = (*fileResource)(nil)
	_ resource.ResourceWithValidateConfig = (*fileResource)(nil)
)

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
