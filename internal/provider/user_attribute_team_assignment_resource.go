package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*userAttributeTeamAssignmentResource)(nil)
	_ resource.ResourceWithConfigure   = (*userAttributeTeamAssignmentResource)(nil)
	_ resource.ResourceWithImportState = (*userAttributeTeamAssignmentResource)(nil)
)

type userAttributeTeamAssignmentResource struct{ configuredResource }

func NewUserAttributeTeamAssignmentResource() resource.Resource {
	return &userAttributeTeamAssignmentResource{}
}

func (r *userAttributeTeamAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_attribute_team_assignment"
}

func (r *userAttributeTeamAssignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}

func (r *userAttributeTeamAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = userAttributeAssignmentSchema("team", "Team ID.", "Manages a Sigma user attribute assignment for one team.")
}

func (r *userAttributeTeamAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := assignmentFromPlan(ctx, req.Plan.GetAttribute, &resp.Diagnostics, "team_id")
	if !resp.Diagnostics.HasError() && setUserAttributeTeam(ctx, r.client, plan, &resp.Diagnostics) {
		setAssignmentState(ctx, plan, resp.State.SetAttribute, &resp.Diagnostics, "team_id")
	}
}

func (r *userAttributeTeamAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readUserAttributeAssignment(ctx, req, resp, "team_id", func(attributeID string) ([]sigma.AttributeAssignment, error) {
		return r.client.ListUserAttributeTeams(ctx, attributeID)
	}, func(assignment sigma.AttributeAssignment) string {
		return assignment.TeamID
	})
}

func (r *userAttributeTeamAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := assignmentFromPlan(ctx, req.Plan.GetAttribute, &resp.Diagnostics, "team_id")
	if !resp.Diagnostics.HasError() && setUserAttributeTeam(ctx, r.client, plan, &resp.Diagnostics) {
		setAssignmentState(ctx, plan, resp.State.SetAttribute, &resp.Diagnostics, "team_id")
	}
}

func (r *userAttributeTeamAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var attributeID, targetID types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("user_attribute_id"), &attributeID)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("team_id"), &targetID)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteUserAttributeTeam(ctx, attributeID.ValueString(), targetID.ValueString()); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma user attribute assignment", err.Error())
	}
}

func (r *userAttributeTeamAssignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importUserAttributeAssignment(ctx, req, resp, "team_id", "teamId")
}

func setUserAttributeTeam(ctx context.Context, client *sigma.Client, plan *assignmentModel, diagnostics interface{ AddError(string, string) }) bool {
	if err := client.SetUserAttributeTeam(ctx, plan.UserAttributeID.ValueString(), plan.TargetID.ValueString(), plan.Value.ValueString()); err != nil {
		diagnostics.AddError("Unable to set Sigma user attribute assignment", err.Error())
		return false
	}
	return true
}
