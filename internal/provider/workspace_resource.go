package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type workspaceResource struct{ configuredResource }

var (
	_ resource.Resource                = (*workspaceResource)(nil)
	_ resource.ResourceWithConfigure   = (*workspaceResource)(nil)
	_ resource.ResourceWithImportState = (*workspaceResource)(nil)
)

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
			"id": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Workspace ID.",
			},
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
	id, idDiags := knownString(plan.ID, "id")
	response.Diagnostics.Append(idDiags...)
	if response.Diagnostics.HasError() {
		return
	}
	value, err := r.client.UpdateWorkspace(ctx, id, sigma.UpdateWorkspaceInput{
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
