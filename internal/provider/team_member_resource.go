package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type teamMemberResource struct{ configuredResource }

var (
	_ resource.Resource                = (*teamMemberResource)(nil)
	_ resource.ResourceWithConfigure   = (*teamMemberResource)(nil)
	_ resource.ResourceWithImportState = (*teamMemberResource)(nil)
)

type teamMemberModel struct {
	ID          types.String `tfsdk:"id"`
	TeamID      types.String `tfsdk:"team_id"`
	MemberID    types.String `tfsdk:"member_id"`
	IsTeamAdmin types.Bool   `tfsdk:"is_team_admin"`
}

func NewTeamMemberResource() resource.Resource { return &teamMemberResource{} }
func (r *teamMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_member"
}
func (r *teamMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *teamMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages one member of a Sigma team. `team_id` and `member_id` force replacement because `PATCH /v2/teams/{teamId}/members` only adds or removes members. `is_team_admin` is read-only; the update-members API cannot change it.",
		Attributes: map[string]schema.Attribute{
			"id":        schema.StringAttribute{Computed: true, MarkdownDescription: "Composite ID in `teamId/memberId` form."},
			"team_id":   schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Team ID."},
			"member_id": schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Member ID."},
			"is_team_admin": schema.BoolAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Whether the member is a team administrator. The update-members API cannot modify this value; Terraform refreshes it from the team members list.",
			},
		},
	}
}
func (r *teamMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teamMemberModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	teamID, teamDiags := knownString(plan.TeamID, "team_id")
	resp.Diagnostics.Append(teamDiags...)
	memberID, memberDiags := knownString(plan.MemberID, "member_id")
	resp.Diagnostics.Append(memberDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.UpdateTeamMembers(ctx, teamID, []string{memberID}, nil); err != nil {
		resp.Diagnostics.AddError("Unable to add Sigma team member", err.Error())
		return
	}
	plan.ID = types.StringValue(teamID + "/" + memberID)
	plan.IsTeamAdmin = types.BoolValue(false)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *teamMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state teamMemberModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	teamID, teamDiags := knownString(state.TeamID, "team_id")
	resp.Diagnostics.Append(teamDiags...)
	memberID, memberDiags := knownString(state.MemberID, "member_id")
	resp.Diagnostics.Append(memberDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	members, err := r.client.ListTeamMembers(ctx, teamID)
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma team members", err.Error())
		return
	}
	for _, member := range members {
		if member.UserID == memberID {
			state.IsTeamAdmin = types.BoolValue(member.IsTeamAdmin)
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}
func (r *teamMemberResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
	// team_id and member_id RequireReplace; is_team_admin is computed and refreshed on Read.
}
func (r *teamMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state teamMemberModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	teamID, teamDiags := knownString(state.TeamID, "team_id")
	resp.Diagnostics.Append(teamDiags...)
	memberID, memberDiags := knownString(state.MemberID, "member_id")
	resp.Diagnostics.Append(memberDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.UpdateTeamMembers(ctx, teamID, nil, []string{memberID}); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to remove Sigma team member", err.Error())
	}
}
func (r *teamMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	teamID, memberID, ok := splitCompositeImportID(req.ID)
	if !ok {
		resp.Diagnostics.AddError("Invalid import ID", "Use `teamId/memberId` with non-empty segments.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("team_id"), teamID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("member_id"), memberID)...)
}
