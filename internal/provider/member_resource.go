package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type memberResource struct{ configuredResource }

var (
	_ resource.Resource                = (*memberResource)(nil)
	_ resource.ResourceWithConfigure   = (*memberResource)(nil)
	_ resource.ResourceWithImportState = (*memberResource)(nil)
	_ resource.ResourceWithModifyPlan  = (*memberResource)(nil)
)

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
	preserveString := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	preserveBool := []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Sigma member. Destroy deactivates the member; Sigma does not permanently delete users. Recreating a member with the same email does not reactivate an archived account; import the archived member and set `is_archived = false` instead. `send_invite` is create-only and cannot be changed after create. Destroy refuses members marked inactive through SCIM (`is_inactive`); deactivate those users in your identity provider, or remove the resource from state with `terraform state rm`. The API does not expose a SCIM-provisioned flag for still-active members. Team membership is managed with `sigma_team_member`.",
		Attributes: map[string]schema.Attribute{
			"id":    schema.StringAttribute{Computed: true, MarkdownDescription: "Member ID."},
			"email": schema.StringAttribute{Required: true, MarkdownDescription: "Member email address."},
			"first_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Member first name.",
			},
			"last_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Member last name.",
			},
			"member_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       preserveString,
				MarkdownDescription: "Account type name.",
			},
			"user_kind": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       preserveString,
				MarkdownDescription: "Member kind: `internal`, `guest`, or `embed`.",
			},
			"organization_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Organization ID."},
			"home_folder_id": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       preserveString,
				MarkdownDescription: "ID of the member's My Documents folder (`homeFolderId`).",
			},
			"is_archived": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       preserveBool,
				MarkdownDescription: "Whether the member is deactivated. Create supports omitting this attribute or setting `false`. An imported archived member can be reactivated by setting `is_archived = false`.",
			},
			"is_inactive": schema.BoolAttribute{
				Computed:            true,
				PlanModifiers:       preserveBool,
				MarkdownDescription: "Whether the member is archived by SCIM. Destroy refuses when this is true.",
			},
			"send_invite": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "When set, passed as the `sendInvite` query parameter on `POST /v2/members`. Create-only; changing this value after create is an error and does not replace the member.",
			},
			"new_owner_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "On destroy, PATCH `newOwnerId` with `isArchived=true` before DELETE so documents transfer to this member instead of the API credential owner. Do not set together with `archive_documents = true`.",
			},
			"archive_documents": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "On destroy, PATCH `archiveDocuments` with `isArchived=true` before DELETE. Archives the member's documents instead of transferring them, and also archives scheduled exports. Cannot be true when `new_owner_id` is set.",
			},
			"archive_scheduled_exports": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "On destroy, PATCH `archiveScheduledExports` with `isArchived=true` before DELETE. Archives scheduled exports instead of transferring them. Can be combined with `new_owner_id`.",
			},
		},
	}
}

func (r *memberResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	var plan memberModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if req.State.Raw.IsNull() {
		if knownTrue(plan.IsArchived) {
			resp.Diagnostics.AddAttributeError(
				path.Root("is_archived"),
				"Cannot create an archived Sigma member",
				"sigma_member create supports omitting `is_archived` or setting it to false. Import an archived member and set `is_archived = false` to reactivate it.",
			)
		}
	} else {
		var state memberModel
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if !plan.SendInvite.Equal(state.SendInvite) && !plan.SendInvite.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("send_invite"),
				"Cannot change send_invite",
				"send_invite is create-only and is sent as the sendInvite query parameter on POST /v2/members. Changing it after create is not supported and does not replace the member.",
			)
		}
	}
	validateMemberDestroyPolicy(plan, &resp.Diagnostics)
}

func validateMemberDestroyPolicy(plan memberModel, diags *diag.Diagnostics) {
	hasOwner := !plan.NewOwnerID.IsNull() && !plan.NewOwnerID.IsUnknown() && plan.NewOwnerID.ValueString() != ""
	if hasOwner && knownTrue(plan.ArchiveDocuments) {
		diags.AddAttributeError(
			path.Root("archive_documents"),
			"Invalid member destroy policy",
			"archive_documents cannot be true when new_owner_id is set. Archive documents instead of transferring them, or transfer them to new_owner_id.",
		)
	}
}

func memberUpdateInput(plan, state *memberModel) sigma.UpdateMemberInput {
	return sigma.UpdateMemberInput{
		FirstName:  changedStringPtr(plan.FirstName, state.FirstName),
		LastName:   changedStringPtr(plan.LastName, state.LastName),
		Email:      changedStringPtr(plan.Email, state.Email),
		MemberType: changedStringPtr(plan.MemberType, state.MemberType),
		UserKind:   changedStringPtr(plan.UserKind, state.UserKind),
		IsArchived: changedBoolPtr(plan.IsArchived, state.IsArchived),
	}
}

func memberCreateInput(plan *memberModel) (sigma.CreateMemberInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	email, emailDiags := knownString(plan.Email, "email")
	diags.Append(emailDiags...)
	firstName, firstDiags := knownString(plan.FirstName, "first_name")
	diags.Append(firstDiags...)
	lastName, lastDiags := knownString(plan.LastName, "last_name")
	diags.Append(lastDiags...)
	if diags.HasError() {
		return sigma.CreateMemberInput{}, diags
	}
	input := sigma.CreateMemberInput{Email: email, FirstName: firstName, LastName: lastName}
	if memberType := optionalStringPtr(plan.MemberType); memberType != nil {
		input.MemberType = *memberType
	}
	if userKind := optionalStringPtr(plan.UserKind); userKind != nil {
		input.UserKind = *userKind
	}
	if sendInvite := optionalBoolPtr(plan.SendInvite); sendInvite != nil {
		input.SendInvite = sendInvite
	}
	return input, diags
}

func memberHasDestroyOptions(state *memberModel) bool {
	return !state.NewOwnerID.IsNull() || !state.ArchiveDocuments.IsNull() || !state.ArchiveScheduledExports.IsNull()
}

func memberDeactivateInput(state *memberModel) sigma.UpdateMemberInput {
	archived := true
	input := sigma.UpdateMemberInput{IsArchived: &archived}
	if id := optionalStringPtr(state.NewOwnerID); id != nil {
		input.NewOwnerID = id
	}
	if archiveDocuments := optionalBoolPtr(state.ArchiveDocuments); archiveDocuments != nil {
		input.ArchiveDocuments = archiveDocuments
	}
	if archiveScheduledExports := optionalBoolPtr(state.ArchiveScheduledExports); archiveScheduledExports != nil {
		input.ArchiveScheduledExports = archiveScheduledExports
	}
	return input
}

func (r *memberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan memberModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if knownTrue(plan.IsArchived) {
		resp.Diagnostics.AddAttributeError(
			path.Root("is_archived"),
			"Cannot create an archived Sigma member",
			"sigma_member create supports omitting `is_archived` or setting it to false. Import an archived member and set `is_archived = false` to reactivate it.",
		)
		return
	}
	input, diags := memberCreateInput(&plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	member, err := r.client.CreateMember(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma member", err.Error())
		return
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
	id, diags := knownString(state.ID, "id")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	member, err := r.client.GetMember(ctx, id)
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
	id, diags := knownString(state.ID, "id")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	member, err := r.client.UpdateMember(ctx, id, memberUpdateInput(&plan, &state))
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
	id, diags := knownString(state.ID, "id")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	member, err := r.client.GetMember(ctx, id)
	if sigma.IsNotFound(err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma member before deactivate", err.Error())
		return
	}
	if member.IsInactive || knownTrue(state.IsInactive) {
		resp.Diagnostics.AddError(
			"Cannot deactivate SCIM-managed Sigma member",
			"Sigma reports this member as inactive through SCIM (isInactive). Do not call DELETE /v2/members for SCIM-managed users; deactivate them in your identity provider instead. To drop Terraform management without calling the API, use `terraform state rm`.",
		)
		return
	}
	if memberHasDestroyOptions(&state) {
		if _, err := r.client.UpdateMember(ctx, id, memberDeactivateInput(&state)); err != nil && !sigma.IsNotFound(err) {
			resp.Diagnostics.AddError("Unable to deactivate Sigma member with destroy options", err.Error())
			return
		}
	}
	if err := r.client.DeleteMember(ctx, id); err != nil && !sigma.IsNotFound(err) {
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
