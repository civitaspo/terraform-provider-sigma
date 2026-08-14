package provider

import (
	"context"
	"fmt"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type folderResource struct{ configuredResource }

var (
	_ resource.Resource                = (*folderResource)(nil)
	_ resource.ResourceWithConfigure   = (*folderResource)(nil)
	_ resource.ResourceWithImportState = (*folderResource)(nil)
)

type folderModel struct {
	ID                types.String `tfsdk:"id"`
	URLID             types.String `tfsdk:"url_id"`
	Name              types.String `tfsdk:"name"`
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

func NewFolderResource() resource.Resource { return &folderResource{} }

func (r *folderResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_folder"
}

func (r *folderResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.configure(request, response)
}

func (r *folderResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	preserve := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	response.Schema = schema.Schema{
		MarkdownDescription: "Manages a Sigma folder through the files API. Create always sends the folder request variant (`type = folder`). Importing or refreshing a non-folder inode is an error. Workbook source-copy and restore are not supported.",
		Attributes: map[string]schema.Attribute{
			"id":   schema.StringAttribute{Computed: true, MarkdownDescription: "Inode ID."},
			"name": schema.StringAttribute{Required: true, MarkdownDescription: "Folder name."},
			"parent_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       preserve,
				MarkdownDescription: "Parent folder or workspace inode ID.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       preserve,
				MarkdownDescription: "Folder description.",
			},
			"owner_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       preserve,
				MarkdownDescription: "Member ID of the folder owner.",
			},
			"url_id":               schema.StringAttribute{Computed: true, MarkdownDescription: "Identifier used in Sigma URLs."},
			"parent_url_id":        schema.StringAttribute{Computed: true, MarkdownDescription: "Parent identifier used in Sigma URLs."},
			"permission":           schema.StringAttribute{Computed: true, MarkdownDescription: "Effective permission."},
			"path":                 schema.StringAttribute{Computed: true, MarkdownDescription: "Folder path."},
			"badge":                schema.StringAttribute{Computed: true, MarkdownDescription: "Badge applied to the folder. Null when Sigma reports no badge."},
			"is_archived":          schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the folder is archived."},
			"parent_source_url_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Source document URL ID for a tenant deployment."},
			"created_by":           schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the member who created the folder."},
			"updated_by":           schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the member who last updated the folder."},
			"created_at":           schema.StringAttribute{Computed: true, MarkdownDescription: "Folder creation timestamp."},
			"updated_at":           schema.StringAttribute{Computed: true, MarkdownDescription: "Folder update timestamp."},
		},
	}
}

func (r *folderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan folderModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	if plan.Name.IsNull() || plan.Name.IsUnknown() {
		response.Diagnostics.AddError("Invalid Sigma folder", "`name` must be known during create.")
		return
	}
	input := sigma.CreateFileInput{Type: "folder", Name: plan.Name.ValueString()}
	if description := optionalStringPtr(plan.Description); description != nil {
		input.Description = *description
	}
	if ownerID := optionalStringPtr(plan.OwnerID); ownerID != nil {
		input.OwnerID = *ownerID
	}
	if parentID := optionalStringPtr(plan.ParentID); parentID != nil {
		input.ParentID = *parentID
	}
	value, err := r.client.CreateFile(ctx, input)
	if err != nil {
		response.Diagnostics.AddError("Unable to create Sigma folder", err.Error())
		return
	}
	if err := requireFolder(value); err != nil {
		response.Diagnostics.AddError("Created Sigma inode is not a folder", err.Error())
		return
	}
	setFolder(&plan, value)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *folderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state folderModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	if state.ID.IsNull() || state.ID.IsUnknown() {
		response.Diagnostics.AddError("Unable to read Sigma folder", "The folder ID is unknown.")
		return
	}
	value, err := r.client.GetFile(ctx, state.ID.ValueString())
	if sigma.IsNotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError("Unable to read Sigma folder", err.Error())
		return
	}
	if err := requireFolder(value); err != nil {
		response.Diagnostics.AddError("Imported Sigma inode is not a folder", err.Error())
		return
	}
	setFolder(&state, value)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *folderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan folderModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	var state folderModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	input := sigma.UpdateFileInput{
		Name:        changedStringPtr(plan.Name, state.Name),
		Description: changedStringPtr(plan.Description, state.Description),
		OwnerID:     changedStringPtr(plan.OwnerID, state.OwnerID),
		ParentID:    changedStringPtr(plan.ParentID, state.ParentID),
	}
	value, err := r.client.UpdateFile(ctx, state.ID.ValueString(), input)
	if err != nil {
		response.Diagnostics.AddError("Unable to update Sigma folder", err.Error())
		return
	}
	if err := requireFolder(value); err != nil {
		response.Diagnostics.AddError("Updated Sigma inode is not a folder", err.Error())
		return
	}
	setFolder(&plan, value)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *folderResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state folderModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteFile(ctx, state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
		response.Diagnostics.AddError("Unable to delete Sigma folder", err.Error())
	}
}

func (r *folderResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	importPassthrough(ctx, request, response)
}

func requireFolder(value *sigma.File) error {
	if value == nil {
		return fmt.Errorf("empty file response")
	}
	if value.Type != "folder" {
		return fmt.Errorf("inode %s has type %q; sigma_folder only manages folders", value.ID, value.Type)
	}
	return nil
}

func setFolder(state *folderModel, value *sigma.File) {
	state.ID = types.StringValue(value.ID)
	state.URLID = types.StringValue(value.URLID)
	state.Name = types.StringValue(value.Name)
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
