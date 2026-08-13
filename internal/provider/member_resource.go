package provider

import (
	"context"
	"fmt"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type memberResource struct{ configuredResource }

var (
	_ resource.Resource                = (*memberResource)(nil)
	_ resource.ResourceWithConfigure   = (*memberResource)(nil)
	_ resource.ResourceWithImportState = (*memberResource)(nil)
)

type memberAddToTeamModel struct {
	TeamID      types.String `tfsdk:"team_id"`
	IsTeamAdmin types.Bool   `tfsdk:"is_team_admin"`
}
type memberModel struct {
	ID                      types.String `tfsdk:"id"`
	Email                   types.String `tfsdk:"email"`
	FirstName               types.String `tfsdk:"first_name"`
	LastName                types.String `tfsdk:"last_name"`
	MemberType              types.String `tfsdk:"member_type"`
	UserKind                types.String `tfsdk:"user_kind"`
	OrganizationID          types.String `tfsdk:"organization_id"`
	HomeFolderID            types.String `tfsdk:"home_folder_id"`
	IsArchived              types.Bool   `tfsdk:"is_archived"`
	IsInactive              types.Bool   `tfsdk:"is_inactive"`
	SendInvite              types.Bool   `tfsdk:"send_invite"`
	AddToTeams              types.Set    `tfsdk:"add_to_teams"`
	NewOwnerID              types.String `tfsdk:"new_owner_id"`
	ArchiveDocuments        types.Bool   `tfsdk:"archive_documents"`
	ArchiveScheduledExports types.Bool   `tfsdk:"archive_scheduled_exports"`
}

func NewMemberResource() resource.Resource { return &memberResource{} }
func (r *memberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_member"
}
func (r *memberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *memberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Sigma member. Destroy deactivates the member; Sigma does not permanently delete users. Recreating a deactivated member reactivates the archived account with the same email when possible. `add_to_teams` and `send_invite` apply only to `POST /v2/members` (not reactivation). Destroy refuses members marked inactive through SCIM (`is_inactive`); deactivate those users in your identity provider, or remove the resource from state with `terraform state rm`. The API does not expose a SCIM-provisioned flag for still-active members. Do not also manage the same team membership with `sigma_team_member` or `sigma_team_members`.",
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Computed: true, MarkdownDescription: "Member ID."},
			"email":           schema.StringAttribute{Required: true, MarkdownDescription: "Member email address."},
			"first_name":      schema.StringAttribute{Required: true, MarkdownDescription: "Member first name."},
			"last_name":       schema.StringAttribute{Required: true, MarkdownDescription: "Member last name."},
			"member_type":     schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Account type name."},
			"user_kind":       schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Member kind: `internal`, `guest`, or `embed`."},
			"organization_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Organization ID."},
			"home_folder_id": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "ID of the member's My Documents folder (`homeFolderId`).",
			},
			"is_archived": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the member is deactivated."},
			"is_inactive": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the member is archived by SCIM. Destroy refuses when this is true."},
			"send_invite": schema.BoolAttribute{
				Optional:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
				MarkdownDescription: "When set, passed as the `sendInvite` query parameter on `POST /v2/members`. Changing this value forces replacement. Recreating an archived member does not resend this parameter.",
			},
			"add_to_teams": schema.SetNestedAttribute{
				Optional:            true,
				PlanModifiers:       []planmodifier.Set{setplanmodifier.RequiresReplace()},
				MarkdownDescription: "Teams to add the member to at create time (`addToTeams`). Changing this value forces replacement. Recreating an archived member does not re-apply this list. Do not also manage the same membership with `sigma_team_member` or `sigma_team_members`.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"team_id":       schema.StringAttribute{Required: true, MarkdownDescription: "Team ID."},
						"is_team_admin": schema.BoolAttribute{Optional: true, MarkdownDescription: "Whether the member is a team administrator. Only settable on member create."},
					},
				},
			},
			"new_owner_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "On destroy, PATCH `newOwnerId` with `isArchived=true` before DELETE so documents transfer to this member instead of the API credential owner.",
			},
			"archive_documents": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "On destroy, PATCH `archiveDocuments` with `isArchived=true` before DELETE. Archives the member's documents instead of transferring them, and also archives scheduled exports. Do not set together with `new_owner_id`.",
			},
			"archive_scheduled_exports": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "On destroy, PATCH `archiveScheduledExports` with `isArchived=true` before DELETE. Archives scheduled exports instead of transferring them.",
			},
		},
	}
}
func memberUpdateInput(plan *memberModel) sigma.UpdateMemberInput {
	first, last, email := plan.FirstName.ValueString(), plan.LastName.ValueString(), plan.Email.ValueString()
	input := sigma.UpdateMemberInput{FirstName: &first, LastName: &last, Email: &email}
	if !plan.MemberType.IsNull() && !plan.MemberType.IsUnknown() {
		memberType := plan.MemberType.ValueString()
		input.MemberType = &memberType
	}
	if !plan.UserKind.IsNull() && !plan.UserKind.IsUnknown() {
		userKind := plan.UserKind.ValueString()
		input.UserKind = &userKind
	}
	return input
}
func memberCreateInput(ctx context.Context, plan *memberModel) (sigma.CreateMemberInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	input := sigma.CreateMemberInput{
		Email: plan.Email.ValueString(), FirstName: plan.FirstName.ValueString(), LastName: plan.LastName.ValueString(),
		MemberType: plan.MemberType.ValueString(), UserKind: plan.UserKind.ValueString(),
	}
	if !plan.SendInvite.IsNull() && !plan.SendInvite.IsUnknown() {
		sendInvite := plan.SendInvite.ValueBool()
		input.SendInvite = &sendInvite
	}
	if !plan.AddToTeams.IsNull() && !plan.AddToTeams.IsUnknown() {
		var teams []memberAddToTeamModel
		diags.Append(plan.AddToTeams.ElementsAs(ctx, &teams, false)...)
		for _, team := range teams {
			item := sigma.AddToTeamInput{TeamID: team.TeamID.ValueString()}
			if !team.IsTeamAdmin.IsNull() && !team.IsTeamAdmin.IsUnknown() {
				admin := team.IsTeamAdmin.ValueBool()
				item.IsTeamAdmin = &admin
			}
			input.AddToTeams = append(input.AddToTeams, item)
		}
	}
	return input, diags
}
func memberHasDestroyOptions(state *memberModel) bool {
	return !state.NewOwnerID.IsNull() || !state.ArchiveDocuments.IsNull() || !state.ArchiveScheduledExports.IsNull()
}
func memberDeactivateInput(state *memberModel) sigma.UpdateMemberInput {
	archived := true
	input := sigma.UpdateMemberInput{IsArchived: &archived}
	if !state.NewOwnerID.IsNull() && !state.NewOwnerID.IsUnknown() {
		id := state.NewOwnerID.ValueString()
		input.NewOwnerID = &id
	}
	if !state.ArchiveDocuments.IsNull() && !state.ArchiveDocuments.IsUnknown() {
		archiveDocuments := state.ArchiveDocuments.ValueBool()
		input.ArchiveDocuments = &archiveDocuments
	}
	if !state.ArchiveScheduledExports.IsNull() && !state.ArchiveScheduledExports.IsUnknown() {
		archiveScheduledExports := state.ArchiveScheduledExports.ValueBool()
		input.ArchiveScheduledExports = &archiveScheduledExports
	}
	return input
}
func (r *memberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan memberModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input, diags := memberCreateInput(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	member, err := r.client.CreateMember(ctx, input)
	if err != nil {
		existing, findErr := r.client.FindMemberByEmail(ctx, plan.Email.ValueString(), true)
		if findErr != nil || existing == nil || !existing.IsArchived {
			resp.Diagnostics.AddError("Unable to create Sigma member", err.Error())
			return
		}
		update := memberUpdateInput(&plan)
		archived := false
		update.IsArchived = &archived
		reactivated, reactivateErr := r.client.UpdateMember(ctx, existing.MemberID, update)
		if reactivateErr != nil {
			resp.Diagnostics.AddError(
				"Unable to reactivate archived Sigma member",
				fmt.Sprintf("Create failed (%s); reactivating archived member %s also failed: %s", err.Error(), existing.MemberID, reactivateErr.Error()),
			)
			return
		}
		member = reactivated
	}
	setMember(&plan, member)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *memberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state memberModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	member, err := r.client.GetMember(ctx, state.ID.ValueString())
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma member", err.Error())
		return
	}
	setMember(&state, member)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *memberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan memberModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state memberModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	member, err := r.client.UpdateMember(ctx, state.ID.ValueString(), memberUpdateInput(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Sigma member", err.Error())
		return
	}
	setMember(&plan, member)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *memberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state memberModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	member, err := r.client.GetMember(ctx, state.ID.ValueString())
	if sigma.IsNotFound(err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma member before deactivate", err.Error())
		return
	}
	if member.IsInactive || state.IsInactive.ValueBool() {
		resp.Diagnostics.AddError(
			"Cannot deactivate SCIM-managed Sigma member",
			"Sigma reports this member as inactive through SCIM (isInactive). Do not call DELETE /v2/members for SCIM-managed users; deactivate them in your identity provider instead. To drop Terraform management without calling the API, use `terraform state rm`.",
		)
		return
	}
	if memberHasDestroyOptions(&state) {
		if _, err := r.client.UpdateMember(ctx, state.ID.ValueString(), memberDeactivateInput(&state)); err != nil && !sigma.IsNotFound(err) {
			resp.Diagnostics.AddError("Unable to deactivate Sigma member with destroy options", err.Error())
			return
		}
	}
	if err := r.client.DeleteMember(ctx, state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to deactivate Sigma member", err.Error())
	}
}
func (r *memberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importPassthrough(ctx, req, resp)
}
func setMember(state *memberModel, value *sigma.Member) {
	state.ID = types.StringValue(value.MemberID)
	state.Email = types.StringValue(value.Email)
	state.FirstName = types.StringValue(value.FirstName)
	state.LastName = types.StringValue(value.LastName)
	state.MemberType = types.StringValue(value.MemberType)
	state.UserKind = types.StringValue(value.UserKind)
	state.OrganizationID = types.StringValue(value.OrganizationID)
	state.IsArchived = types.BoolValue(value.IsArchived)
	state.IsInactive = types.BoolValue(value.IsInactive)
	if value.HomeFolderID != "" {
		state.HomeFolderID = types.StringValue(value.HomeFolderID)
	} else if state.HomeFolderID.IsUnknown() {
		state.HomeFolderID = types.StringNull()
	}
}
