package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type teamMembersResource struct{ configuredResource }

var (
	_ resource.Resource                = (*teamMembersResource)(nil)
	_ resource.ResourceWithConfigure   = (*teamMembersResource)(nil)
	_ resource.ResourceWithImportState = (*teamMembersResource)(nil)
)

type teamMembersModel struct {
	ID        types.String `tfsdk:"id"`
	TeamID    types.String `tfsdk:"team_id"`
	MemberIDs types.Set    `tfsdk:"member_ids"`
}

func NewTeamMembersResource() resource.Resource { return &teamMembersResource{} }
func (r *teamMembersResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_members"
}
func (r *teamMembersResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *teamMembersResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Authoritatively manages all members of a Sigma team. On apply, members not listed in `member_ids` are removed; on destroy, every member currently in state is removed. Do not use with `sigma_team_member` for the same team.",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Team ID."},
			"team_id":    schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "Team ID."},
			"member_ids": schema.SetAttribute{Required: true, ElementType: types.StringType, MarkdownDescription: "Authoritative set of member IDs. Members absent from this set are removed from the team."},
		},
	}
}
func (r *teamMembersResource) sync(ctx context.Context, plan *teamMembersModel, diagnostics interface{ AddError(string, string) }) bool {
	var desired []string
	plan.MemberIDs.ElementsAs(ctx, &desired, false)
	current, err := r.client.ListTeamMembers(ctx, plan.TeamID.ValueString())
	if err != nil {
		diagnostics.AddError("Unable to read Sigma team members", err.Error())
		return false
	}
	have, want := map[string]bool{}, map[string]bool{}
	for _, member := range current {
		have[member.UserID] = true
	}
	for _, id := range desired {
		want[id] = true
	}
	var add, remove []string
	for id := range want {
		if !have[id] {
			add = append(add, id)
		}
	}
	for id := range have {
		if !want[id] {
			remove = append(remove, id)
		}
	}
	if err := r.client.UpdateTeamMembers(ctx, plan.TeamID.ValueString(), add, remove); err != nil {
		diagnostics.AddError("Unable to update Sigma team members", err.Error())
		return false
	}
	plan.ID = plan.TeamID
	return true
}
func (r *teamMembersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teamMembersModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if !resp.Diagnostics.HasError() && r.sync(ctx, &plan, &resp.Diagnostics) {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}
func (r *teamMembersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state teamMembersModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	members, err := r.client.ListTeamMembers(ctx, state.TeamID.ValueString())
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma team members", err.Error())
		return
	}
	ids := make([]string, 0, len(members))
	for _, member := range members {
		ids = append(ids, member.UserID)
	}
	state.MemberIDs, _ = types.SetValueFrom(ctx, types.StringType, ids)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *teamMembersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan teamMembersModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if !resp.Diagnostics.HasError() && r.sync(ctx, &plan, &resp.Diagnostics) {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}
func (r *teamMembersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state teamMembersModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var ids []string
	state.MemberIDs.ElementsAs(ctx, &ids, false)
	if err := r.client.UpdateTeamMembers(ctx, state.TeamID.ValueString(), nil, ids); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to remove Sigma team members", err.Error())
	}
}
func (r *teamMembersResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("team_id"), req.ID)...)
}
