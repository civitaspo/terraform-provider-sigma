package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type configuredResource struct{ client *sigma.Client }

func (r *configuredResource) configure(request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}
	client, ok := request.ProviderData.(*sigma.Client)
	if !ok {
		response.Diagnostics.AddError("Unexpected resource configuration type", "The Sigma provider returned an unexpected client type.")
		return
	}
	r.client = client
}

func importPassthrough(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

type memberResource struct{ configuredResource }
type memberModel struct {
	ID             types.String `tfsdk:"id"`
	Email          types.String `tfsdk:"email"`
	FirstName      types.String `tfsdk:"first_name"`
	LastName       types.String `tfsdk:"last_name"`
	MemberType     types.String `tfsdk:"member_type"`
	UserKind       types.String `tfsdk:"user_kind"`
	OrganizationID types.String `tfsdk:"organization_id"`
	IsArchived     types.Bool   `tfsdk:"is_archived"`
	IsInactive     types.Bool   `tfsdk:"is_inactive"`
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
		MarkdownDescription: "Manages a Sigma member. Destroy deactivates the member; Sigma does not permanently delete users. Recreating a deactivated member reactivates the archived account with the same email when possible.",
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Computed: true, MarkdownDescription: "Member ID."},
			"email":           schema.StringAttribute{Required: true, MarkdownDescription: "Member email address."},
			"first_name":      schema.StringAttribute{Required: true, MarkdownDescription: "Member first name."},
			"last_name":       schema.StringAttribute{Required: true, MarkdownDescription: "Member last name."},
			"member_type":     schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Account type name."},
			"user_kind":       schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Member kind: `internal`, `guest`, or `embed`."},
			"organization_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Organization ID."},
			"is_archived":     schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the member is deactivated."},
			"is_inactive":     schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the member is inactive through SCIM."},
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
func (r *memberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan memberModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	member, err := r.client.CreateMember(ctx, sigma.CreateMemberInput{
		Email: plan.Email.ValueString(), FirstName: plan.FirstName.ValueString(), LastName: plan.LastName.ValueString(),
		MemberType: plan.MemberType.ValueString(), UserKind: plan.UserKind.ValueString(),
	})
	if err != nil {
		existing, findErr := r.client.FindMemberByEmail(ctx, plan.Email.ValueString(), true)
		if findErr != nil || existing == nil || !existing.IsArchived {
			resp.Diagnostics.AddError("Unable to create Sigma member", err.Error())
			return
		}
		input := memberUpdateInput(&plan)
		archived := false
		input.IsArchived = &archived
		reactivated, reactivateErr := r.client.UpdateMember(ctx, existing.MemberID, input)
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
}

type teamResource struct{ configuredResource }
type teamModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Visibility  types.String `tfsdk:"visibility"`
	IsArchived  types.Bool   `tfsdk:"is_archived"`
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
		MarkdownDescription: "Manages a Sigma team.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Team ID."},
			"name":        schema.StringAttribute{Required: true, MarkdownDescription: "Team name."},
			"description": schema.StringAttribute{Optional: true, MarkdownDescription: "Team description."},
			"visibility":  schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Team visibility: `public` or `private`."},
			"is_archived": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the team is archived."},
		},
	}
}
func (r *teamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teamModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	team, err := r.client.CreateTeam(ctx, sigma.CreateTeamInput{Name: plan.Name.ValueString(), Description: plan.Description.ValueString(), Visibility: plan.Visibility.ValueString()})
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
	team, err := r.client.GetTeam(ctx, state.ID.ValueString())
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
	if resp.Diagnostics.HasError() {
		return
	}
	name, description, visibility := plan.Name.ValueString(), plan.Description.ValueString(), plan.Visibility.ValueString()
	team, err := r.client.UpdateTeam(ctx, plan.ID.ValueString(), sigma.UpdateTeamInput{Name: &name, Description: &description, Visibility: &visibility})
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
	if err := r.client.DeleteTeam(ctx, state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
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
}

type teamMemberResource struct{ configuredResource }
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
		MarkdownDescription: "Manages one member of a Sigma team. Do not use with `sigma_team_members` for the same team.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true, MarkdownDescription: "Composite ID in `teamId/memberId` form."},
			"team_id":       schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Team ID."},
			"member_id":     schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Member ID."},
			"is_team_admin": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the member is a team administrator. The update-members API cannot modify this value."},
		},
	}
}
func (r *teamMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teamMemberModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.UpdateTeamMembers(ctx, plan.TeamID.ValueString(), []string{plan.MemberID.ValueString()}, nil); err != nil {
		resp.Diagnostics.AddError("Unable to add Sigma team member", err.Error())
		return
	}
	plan.ID = types.StringValue(plan.TeamID.ValueString() + "/" + plan.MemberID.ValueString())
	plan.IsTeamAdmin = types.BoolValue(false)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *teamMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state teamMemberModel
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
	for _, member := range members {
		if member.UserID == state.MemberID.ValueString() {
			state.IsTeamAdmin = types.BoolValue(member.IsTeamAdmin)
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}
func (r *teamMemberResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
}
func (r *teamMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state teamMemberModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.UpdateTeamMembers(ctx, state.TeamID.ValueString(), nil, []string{state.MemberID.ValueString()}); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to remove Sigma team member", err.Error())
	}
}
func (r *teamMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Use `teamId/memberId`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("team_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("member_id"), parts[1])...)
}

type teamMembersResource struct{ configuredResource }
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
		MarkdownDescription: "Authoritatively manages all members of a Sigma team. Do not use with `sigma_team_member` for the same team.",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Team ID."},
			"team_id":    schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "Team ID."},
			"member_ids": schema.SetAttribute{Required: true, ElementType: types.StringType, MarkdownDescription: "Authoritative set of member IDs."},
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

type accountTypeResource struct{ configuredResource }
type accountTypeModel struct {
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	Description             types.String `tfsdk:"description"`
	Permissions             types.Set    `tfsdk:"permissions"`
	IsCustom                types.Bool   `tfsdk:"is_custom"`
	ReassignToAccountTypeID types.String `tfsdk:"reassign_to_account_type_id"`
}

func NewAccountTypeResource() resource.Resource { return &accountTypeResource{} }
func (r *accountTypeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account_type"
}
func (r *accountTypeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *accountTypeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a custom Sigma account type. The API has no update or get-by-ID endpoint, so configuration changes replace it. Import by account type name.",
		Attributes: map[string]schema.Attribute{
			"id":                          schema.StringAttribute{Computed: true, MarkdownDescription: "Account type ID."},
			"name":                        schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Account type name."},
			"description":                 schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Account type description."},
			"permissions":                 schema.SetAttribute{Required: true, ElementType: types.StringType, PlanModifiers: []planmodifier.Set{setplanmodifier.RequiresReplace()}, MarkdownDescription: "Enabled permission names."},
			"is_custom":                   schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this is a custom account type."},
			"reassign_to_account_type_id": schema.StringAttribute{Optional: true, PlanModifiers: replace, MarkdownDescription: "Account type ID to receive users when this type is deleted."},
		},
	}
}
func (r *accountTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan accountTypeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var permissions []string
	plan.Permissions.ElementsAs(ctx, &permissions, false)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.CreateAccountType(ctx, plan.Name.ValueString(), plan.Description.ValueString(), permissions)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma account type", err.Error())
		return
	}
	setAccountType(&plan, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *accountTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state accountTypeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	lookup := state.ID.ValueString()
	if !state.Name.IsNull() && !state.Name.IsUnknown() && state.Name.ValueString() != "" {
		lookup = state.Name.ValueString()
	}
	value, err := r.client.FindAccountType(ctx, lookup)
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma account type", err.Error())
		return
	}
	setAccountType(&state, value)
	permissions, err := r.client.ListAccountTypePermissions(ctx, value.AccountTypeID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma account type permissions", err.Error())
		return
	}
	names := make([]string, len(permissions))
	for i := range permissions {
		names[i] = permissions[i].Permission
	}
	state.Permissions, _ = types.SetValueFrom(ctx, types.StringType, names)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *accountTypeResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
}
func (r *accountTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state accountTypeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if !resp.Diagnostics.HasError() {
		if err := r.client.DeleteAccountType(ctx, state.ID.ValueString(), state.ReassignToAccountTypeID.ValueString()); err != nil && !sigma.IsNotFound(err) {
			resp.Diagnostics.AddError("Unable to delete Sigma account type", err.Error())
		}
	}
}
func (r *accountTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}
func setAccountType(state *accountTypeModel, value *sigma.AccountType) {
	state.ID = types.StringValue(value.AccountTypeID)
	state.Name = types.StringValue(value.AccountTypeName)
	state.Description = types.StringValue(value.Description)
	state.IsCustom = types.BoolValue(value.IsCustom)
}

type userAttributeResource struct{ configuredResource }
type userAttributeModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	DefaultValue types.String `tfsdk:"default_value"`
}

func NewUserAttributeResource() resource.Resource { return &userAttributeResource{} }
func (r *userAttributeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_attribute"
}
func (r *userAttributeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *userAttributeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Sigma user attribute. The API has no update endpoint, so configuration changes replace it.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true, MarkdownDescription: "User attribute ID."},
			"name":          schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "User attribute name."},
			"description":   schema.StringAttribute{Optional: true, PlanModifiers: replace, MarkdownDescription: "User attribute description."},
			"default_value": schema.StringAttribute{Optional: true, PlanModifiers: replace, MarkdownDescription: "Default string value."},
		},
	}
}
func (r *userAttributeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userAttributeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var defaultValue *sigma.AttributeValue
	if !plan.DefaultValue.IsNull() {
		defaultValue = &sigma.AttributeValue{Val: plan.DefaultValue.ValueString(), Type: "string"}
	}
	value, err := r.client.CreateUserAttribute(ctx, plan.Name.ValueString(), plan.Description.ValueString(), defaultValue)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma user attribute", err.Error())
		return
	}
	setUserAttribute(&plan, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *userAttributeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userAttributeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	value, err := r.client.GetUserAttribute(ctx, state.ID.ValueString())
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma user attribute", err.Error())
		return
	}
	setUserAttribute(&state, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *userAttributeResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
}
func (r *userAttributeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userAttributeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if !resp.Diagnostics.HasError() {
		if err := r.client.DeleteUserAttribute(ctx, state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
			resp.Diagnostics.AddError("Unable to delete Sigma user attribute", err.Error())
		}
	}
}
func (r *userAttributeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importPassthrough(ctx, req, resp)
}
func setUserAttribute(state *userAttributeModel, value *sigma.UserAttribute) {
	state.ID = types.StringValue(value.UserAttributeID)
	state.Name = types.StringValue(value.Name)
	if value.Description == nil {
		state.Description = types.StringNull()
	} else {
		state.Description = types.StringValue(*value.Description)
	}
	if value.DefaultValue == nil {
		state.DefaultValue = types.StringNull()
	} else {
		state.DefaultValue = types.StringValue(value.DefaultValue.Val)
	}
}

type attributeAssignmentResource struct {
	configuredResource
	target string
}
type assignmentModel struct {
	UserAttributeID types.String
	TargetID        types.String
	Value           types.String
}

func NewUserAttributeTeamAssignmentResource() resource.Resource {
	return &attributeAssignmentResource{target: "team"}
}
func NewUserAttributeUserAssignmentResource() resource.Resource {
	return &attributeAssignmentResource{target: "user"}
}
func (r *attributeAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_attribute_" + r.target + "_assignment"
}
func (r *attributeAssignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *attributeAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	targetLabel := "User"
	if r.target == "team" {
		targetLabel = "Team"
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: fmt.Sprintf("Manages a Sigma user attribute assignment for one %s.", r.target),
		Attributes: map[string]schema.Attribute{
			"id":                schema.StringAttribute{Computed: true, MarkdownDescription: "Composite assignment ID."},
			"user_attribute_id": schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "User attribute ID."},
			r.target + "_id":    schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: targetLabel + " ID."},
			"value":             schema.StringAttribute{Required: true, MarkdownDescription: "Assigned string value."},
		},
	}
}
func (r *attributeAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := r.assignmentFromPlan(ctx, req.Plan.GetAttribute, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() && r.set(ctx, plan, &resp.Diagnostics) {
		r.setAssignmentState(ctx, plan, resp.State.SetAttribute, &resp.Diagnostics)
	}
}
func (r *attributeAssignmentResource) set(ctx context.Context, plan *assignmentModel, diagnostics interface{ AddError(string, string) }) bool {
	var err error
	if r.target == "team" {
		err = r.client.SetUserAttributeTeam(ctx, plan.UserAttributeID.ValueString(), plan.TargetID.ValueString(), plan.Value.ValueString())
	} else {
		err = r.client.SetUserAttributeUser(ctx, plan.UserAttributeID.ValueString(), plan.TargetID.ValueString(), plan.Value.ValueString())
	}
	if err != nil {
		diagnostics.AddError("Unable to set Sigma user attribute assignment", err.Error())
		return false
	}
	return true
}
func (r *attributeAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var attributeID, targetID types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("user_attribute_id"), &attributeID)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root(r.target+"_id"), &targetID)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var assignments []sigma.AttributeAssignment
	var err error
	if r.target == "team" {
		assignments, err = r.client.ListUserAttributeTeams(ctx, attributeID.ValueString())
	} else {
		assignments, err = r.client.ListUserAttributeUsers(ctx, attributeID.ValueString())
	}
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma user attribute assignments", err.Error())
		return
	}
	for _, assignment := range assignments {
		id := assignment.UserID
		if r.target == "team" {
			id = assignment.TeamID
		}
		if id == targetID.ValueString() {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("value"), assignment.Value.Val)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}
func (r *attributeAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := r.assignmentFromPlan(ctx, req.Plan.GetAttribute, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() && r.set(ctx, plan, &resp.Diagnostics) {
		r.setAssignmentState(ctx, plan, resp.State.SetAttribute, &resp.Diagnostics)
	}
}
func (r *attributeAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var attributeID, targetID types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("user_attribute_id"), &attributeID)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root(r.target+"_id"), &targetID)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var err error
	if r.target == "team" {
		err = r.client.DeleteUserAttributeTeam(ctx, attributeID.ValueString(), targetID.ValueString())
	} else {
		err = r.client.DeleteUserAttributeUser(ctx, attributeID.ValueString(), targetID.ValueString())
	}
	if err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma user attribute assignment", err.Error())
	}
}
func (r *attributeAssignmentResource) assignmentFromPlan(
	ctx context.Context,
	get func(context.Context, path.Path, any) diag.Diagnostics,
	diagnostics *diag.Diagnostics,
) *assignmentModel {
	plan := &assignmentModel{}
	diagnostics.Append(get(ctx, path.Root("user_attribute_id"), &plan.UserAttributeID)...)
	diagnostics.Append(get(ctx, path.Root(r.target+"_id"), &plan.TargetID)...)
	diagnostics.Append(get(ctx, path.Root("value"), &plan.Value)...)
	return plan
}
func (r *attributeAssignmentResource) setAssignmentState(
	ctx context.Context,
	plan *assignmentModel,
	set func(context.Context, path.Path, any) diag.Diagnostics,
	diagnostics *diag.Diagnostics,
) {
	diagnostics.Append(set(ctx, path.Root("id"), plan.UserAttributeID.ValueString()+"/"+plan.TargetID.ValueString())...)
	diagnostics.Append(set(ctx, path.Root("user_attribute_id"), plan.UserAttributeID)...)
	diagnostics.Append(set(ctx, path.Root(r.target+"_id"), plan.TargetID)...)
	diagnostics.Append(set(ctx, path.Root("value"), plan.Value)...)
}
func (r *attributeAssignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Use `userAttributeId/"+r.target+"Id`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_attribute_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(r.target+"_id"), parts[1])...)
}

var (
	_ resource.ResourceWithImportState = (*memberResource)(nil)
	_ resource.ResourceWithImportState = (*teamResource)(nil)
	_ resource.ResourceWithImportState = (*teamMemberResource)(nil)
	_ resource.ResourceWithImportState = (*teamMembersResource)(nil)
	_ resource.ResourceWithImportState = (*accountTypeResource)(nil)
	_ resource.ResourceWithImportState = (*userAttributeResource)(nil)
	_ resource.ResourceWithImportState = (*attributeAssignmentResource)(nil)
)
