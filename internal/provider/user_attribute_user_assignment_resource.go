package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*userAttributeUserAssignmentResource)(nil)
	_ resource.ResourceWithConfigure   = (*userAttributeUserAssignmentResource)(nil)
	_ resource.ResourceWithImportState = (*userAttributeUserAssignmentResource)(nil)
)

type userAttributeUserAssignmentResource struct{ configuredResource }

type assignmentModel struct {
	UserAttributeID types.String
	TargetID        types.String
	Value           types.String
}

func NewUserAttributeUserAssignmentResource() resource.Resource {
	return &userAttributeUserAssignmentResource{}
}

func (r *userAttributeUserAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_attribute_user_assignment"
}

func (r *userAttributeUserAssignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}

func (r *userAttributeUserAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = userAttributeAssignmentSchema("user", "User ID.", "Manages a Sigma user attribute assignment for one user.")
}

func userAttributeAssignmentSchema(target, targetDescription, description string) schema.Schema {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	return schema.Schema{
		MarkdownDescription: description,
		Attributes: map[string]schema.Attribute{
			"id":                schema.StringAttribute{Computed: true, MarkdownDescription: "Composite assignment ID."},
			"user_attribute_id": schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "User attribute ID."},
			target + "_id":      schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: targetDescription},
			"value":             schema.StringAttribute{Required: true, MarkdownDescription: "Assigned string value."},
		},
	}
}

func (r *userAttributeUserAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := assignmentFromPlan(ctx, req.Plan.GetAttribute, &resp.Diagnostics, "user_id")
	if !resp.Diagnostics.HasError() && setUserAttributeUser(ctx, r.client, plan, &resp.Diagnostics) {
		setAssignmentState(ctx, plan, resp.State.SetAttribute, &resp.Diagnostics, "user_id")
	}
}

func (r *userAttributeUserAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	readUserAttributeAssignment(ctx, req, resp, "user_id", func(attributeID string) ([]sigma.AttributeAssignment, error) {
		return r.client.ListUserAttributeUsers(ctx, attributeID)
	}, func(assignment sigma.AttributeAssignment) string {
		return assignment.UserID
	})
}

func (r *userAttributeUserAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := assignmentFromPlan(ctx, req.Plan.GetAttribute, &resp.Diagnostics, "user_id")
	if !resp.Diagnostics.HasError() && setUserAttributeUser(ctx, r.client, plan, &resp.Diagnostics) {
		setAssignmentState(ctx, plan, resp.State.SetAttribute, &resp.Diagnostics, "user_id")
	}
}

func (r *userAttributeUserAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var attributeID, targetID types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("user_attribute_id"), &attributeID)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("user_id"), &targetID)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteUserAttributeUser(ctx, attributeID.ValueString(), targetID.ValueString()); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma user attribute assignment", err.Error())
	}
}

func (r *userAttributeUserAssignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importUserAttributeAssignment(ctx, req, resp, "user_id", "userId")
}

func setUserAttributeUser(ctx context.Context, client *sigma.Client, plan *assignmentModel, diagnostics interface{ AddError(string, string) }) bool {
	if err := client.SetUserAttributeUser(ctx, plan.UserAttributeID.ValueString(), plan.TargetID.ValueString(), plan.Value.ValueString()); err != nil {
		diagnostics.AddError("Unable to set Sigma user attribute assignment", err.Error())
		return false
	}
	return true
}

func assignmentFromPlan(
	ctx context.Context,
	get func(context.Context, path.Path, any) diag.Diagnostics,
	diagnostics *diag.Diagnostics,
	targetAttr string,
) *assignmentModel {
	plan := &assignmentModel{}
	diagnostics.Append(get(ctx, path.Root("user_attribute_id"), &plan.UserAttributeID)...)
	diagnostics.Append(get(ctx, path.Root(targetAttr), &plan.TargetID)...)
	diagnostics.Append(get(ctx, path.Root("value"), &plan.Value)...)
	return plan
}

func setAssignmentState(
	ctx context.Context,
	plan *assignmentModel,
	set func(context.Context, path.Path, any) diag.Diagnostics,
	diagnostics *diag.Diagnostics,
	targetAttr string,
) {
	diagnostics.Append(set(ctx, path.Root("id"), plan.UserAttributeID.ValueString()+"/"+plan.TargetID.ValueString())...)
	diagnostics.Append(set(ctx, path.Root("user_attribute_id"), plan.UserAttributeID)...)
	diagnostics.Append(set(ctx, path.Root(targetAttr), plan.TargetID)...)
	diagnostics.Append(set(ctx, path.Root("value"), plan.Value)...)
}

func readUserAttributeAssignment(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
	targetAttr string,
	list func(attributeID string) ([]sigma.AttributeAssignment, error),
	matchID func(sigma.AttributeAssignment) string,
) {
	var attributeID, targetID types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("user_attribute_id"), &attributeID)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root(targetAttr), &targetID)...)
	if resp.Diagnostics.HasError() {
		return
	}
	assignments, err := list(attributeID.ValueString())
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma user attribute assignments", err.Error())
		return
	}
	for _, assignment := range assignments {
		if matchID(assignment) == targetID.ValueString() {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("value"), assignment.Value.Val)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}

func importUserAttributeAssignment(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse, targetAttr, targetToken string) {
	attributeID, targetID, ok := splitCompositeImportID(req.ID)
	if !ok {
		resp.Diagnostics.AddError("Invalid import ID", "Use `userAttributeId/"+targetToken+"` with non-empty segments.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_attribute_id"), attributeID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(targetAttr), targetID)...)
}
