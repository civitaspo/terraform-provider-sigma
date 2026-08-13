package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type teamResource struct{ configuredResource }

var (
	_ resource.Resource                = (*teamResource)(nil)
	_ resource.ResourceWithConfigure   = (*teamResource)(nil)
	_ resource.ResourceWithImportState = (*teamResource)(nil)
)

type teamModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	Visibility       types.String `tfsdk:"visibility"`
	IsArchived       types.Bool   `tfsdk:"is_archived"`
	CreateTeamFolder types.Bool   `tfsdk:"create_team_folder"`
	WorkspaceID      types.String `tfsdk:"workspace_id"`
}

func NewTeamResource() resource.Resource { return &teamResource{} }
func (r *teamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}
func (r *teamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *teamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Sigma team. `create_team_folder` is accepted only on create (`POST /v2/teams`); changing it forces replacement.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Team ID."},
			"name":        schema.StringAttribute{Required: true, MarkdownDescription: "Team name."},
			"description": schema.StringAttribute{Optional: true, MarkdownDescription: "Team description."},
			"visibility": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Team visibility: `public` or `private`.",
			},
			"is_archived": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the team is archived."},
			"create_team_folder": schema.BoolAttribute{
				Optional:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
				MarkdownDescription: "When true, Sigma creates a workspace associated with the team at create time. The API only accepts `createTeamFolder` on `POST /v2/teams`; changing this value forces replacement.",
			},
			"workspace_id": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "ID of the team workspace when one exists. Present on create responses; later GET and list responses may omit it, in which case Terraform keeps the last known value.",
			},
		},
	}
}
func (r *teamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teamModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	name, diags := knownString(plan.Name, "name")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := sigma.CreateTeamInput{Name: name}
	if description := optionalStringPtr(plan.Description); description != nil {
		input.Description = *description
	}
	if visibility := optionalStringPtr(plan.Visibility); visibility != nil {
		input.Visibility = *visibility
	}
	if createTeamFolder := optionalBoolPtr(plan.CreateTeamFolder); createTeamFolder != nil {
		input.CreateTeamFolder = createTeamFolder
	}
	team, err := r.client.CreateTeam(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma team", err.Error())
		return
	}
	setTeam(&plan, team)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *teamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state teamModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, diags := knownString(state.ID, "id")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	team, err := r.client.GetTeam(ctx, id)
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma team", err.Error())
		return
	}
	setTeam(&state, team)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *teamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan teamModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state teamModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, diags := knownString(state.ID, "id")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	team, err := r.client.UpdateTeam(ctx, id, sigma.UpdateTeamInput{
		Name:        changedStringPtr(plan.Name, state.Name),
		Description: changedStringPtr(plan.Description, state.Description),
		Visibility:  changedStringPtr(plan.Visibility, state.Visibility),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Sigma team", err.Error())
		return
	}
	setTeam(&plan, team)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *teamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state teamModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, diags := knownString(state.ID, "id")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteTeam(ctx, id); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma team", err.Error())
	}
}
func (r *teamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importPassthrough(ctx, req, resp)
}
func setTeam(state *teamModel, value *sigma.Team) {
	state.ID = types.StringValue(value.TeamID)
	state.Name = types.StringValue(value.Name)
	if value.Description == nil {
		state.Description = types.StringNull()
	} else {
		state.Description = types.StringValue(*value.Description)
	}
	state.Visibility = types.StringValue(value.Visibility)
	state.IsArchived = types.BoolValue(value.IsArchived)
	if value.WorkspaceID != nil && *value.WorkspaceID != "" {
		state.WorkspaceID = types.StringValue(*value.WorkspaceID)
	} else if state.WorkspaceID.IsUnknown() {
		state.WorkspaceID = types.StringNull()
	}
}
