package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                   = (*workspaceGrantResource)(nil)
	_ resource.ResourceWithConfigure      = (*workspaceGrantResource)(nil)
	_ resource.ResourceWithImportState    = (*workspaceGrantResource)(nil)
	_ resource.ResourceWithValidateConfig = (*workspaceGrantResource)(nil)
)

type workspaceGrantResource struct{ configuredResource }

func NewWorkspaceGrantResource() resource.Resource { return &workspaceGrantResource{} }

func (r *workspaceGrantResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_workspace_grant"
}

func (r *workspaceGrantResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.configure(request, response)
}

func (r *workspaceGrantResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = grantSchema(
		"Manages a fine-grained Sigma workspace grant.",
		"Workspace ID.",
		"Workspace permission: `view`, `explore`, `organize`, or `edit`.",
	)
}

func (r *workspaceGrantResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var config grantModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}
	if config.MemberID.IsUnknown() || config.TeamID.IsUnknown() || config.Permission.IsUnknown() || config.TagID.IsUnknown() {
		return
	}
	validateGrant(&config, workspaceGrantPermissions, false, &response.Diagnostics)
}

func (r *workspaceGrantResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan grantModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() || !validateGrant(&plan, workspaceGrantPermissions, false, &response.Diagnostics) {
		return
	}
	grantee := sigma.Grantee{MemberID: plan.MemberID.ValueString(), TeamID: plan.TeamID.ValueString()}
	err := r.client.CreateWorkspaceGrant(ctx, plan.InodeID.ValueString(), grantee, plan.Permission.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Unable to create Sigma grant", err.Error())
		return
	}
	value, err := lookupListedGrant(func() ([]sigma.Grant, error) {
		return r.client.ListWorkspaceGrants(ctx, plan.InodeID.ValueString())
	}, &plan, "")
	if err != nil {
		response.Diagnostics.AddError("Unable to locate created Sigma grant", err.Error())
		return
	}
	setGrant(&plan, value)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *workspaceGrantResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state grantModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	value, err := lookupListedGrant(func() ([]sigma.Grant, error) {
		return r.client.ListWorkspaceGrants(ctx, state.InodeID.ValueString())
	}, &state, state.ID.ValueString())
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

func (r *workspaceGrantResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
}

func (r *workspaceGrantResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state grantModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteWorkspaceGrant(ctx, state.InodeID.ValueString(), state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
		response.Diagnostics.AddError("Unable to delete Sigma grant", err.Error())
	}
}

func (r *workspaceGrantResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	importGrantCompositeID(ctx, request, response)
}
