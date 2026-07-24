package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type identityDataSource struct {
	client *sigma.Client
	kind   string
}

func newIdentityDataSource(kind string) datasource.DataSource { return &identityDataSource{kind: kind} }
func NewMemberDataSource() datasource.DataSource              { return newIdentityDataSource("member") }
func NewMembersDataSource() datasource.DataSource             { return newIdentityDataSource("members") }
func NewTeamDataSource() datasource.DataSource                { return newIdentityDataSource("team") }
func NewTeamsDataSource() datasource.DataSource               { return newIdentityDataSource("teams") }
func NewAccountTypesDataSource() datasource.DataSource        { return newIdentityDataSource("account_types") }
func NewUserAttributesDataSource() datasource.DataSource {
	return newIdentityDataSource("user_attributes")
}

func (d *identityDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.kind
}
func (d *identityDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sigma.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected data source configuration type", "The Sigma provider returned an unexpected client type.")
		return
	}
	d.client = client
}
func (d *identityDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	switch d.kind {
	case "member":
		resp.Schema = schema.Schema{MarkdownDescription: "Retrieves a Sigma member by ID.", Attributes: memberDataAttributes(true)}
	case "team":
		resp.Schema = schema.Schema{MarkdownDescription: "Retrieves a Sigma team by ID.", Attributes: teamDataAttributes(true)}
	case "members":
		resp.Schema = schema.Schema{MarkdownDescription: "Lists all Sigma members.", Attributes: map[string]schema.Attribute{
			"id":      schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
			"members": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Members.", NestedObject: schema.NestedAttributeObject{Attributes: memberDataAttributes(false)}},
		}}
	case "teams":
		resp.Schema = schema.Schema{MarkdownDescription: "Lists all Sigma teams using the paginated v2.1 endpoint.", Attributes: map[string]schema.Attribute{
			"id":    schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
			"teams": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Teams.", NestedObject: schema.NestedAttributeObject{Attributes: teamDataAttributes(false)}},
		}}
	case "account_types":
		resp.Schema = schema.Schema{MarkdownDescription: "Lists Sigma account types.", Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
			"account_types": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Account types.", NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{Computed: true, MarkdownDescription: "Account type ID."}, "name": schema.StringAttribute{Computed: true, MarkdownDescription: "Account type name."},
				"description": schema.StringAttribute{Computed: true, MarkdownDescription: "Description."}, "is_custom": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the type is custom."},
			}}},
		}}
	case "user_attributes":
		resp.Schema = schema.Schema{MarkdownDescription: "Lists Sigma user attributes.", Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
			"user_attributes": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "User attributes.", NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{Computed: true, MarkdownDescription: "User attribute ID."}, "name": schema.StringAttribute{Computed: true, MarkdownDescription: "Name."},
				"description": schema.StringAttribute{Computed: true, MarkdownDescription: "Description."}, "default_value": schema.StringAttribute{Computed: true, MarkdownDescription: "Default string value."},
			}}},
		}}
	}
}

func memberDataAttributes(requireID bool) map[string]schema.Attribute {
	id := schema.StringAttribute{Computed: true, MarkdownDescription: "Member ID."}
	if requireID {
		id = schema.StringAttribute{Required: true, MarkdownDescription: "Member ID."}
	}
	return map[string]schema.Attribute{
		"id": id, "email": schema.StringAttribute{Computed: true, MarkdownDescription: "Email address."},
		"first_name": schema.StringAttribute{Computed: true, MarkdownDescription: "First name."}, "last_name": schema.StringAttribute{Computed: true, MarkdownDescription: "Last name."},
		"member_type": schema.StringAttribute{Computed: true, MarkdownDescription: "Account type."}, "user_kind": schema.StringAttribute{Computed: true, MarkdownDescription: "User kind."},
		"organization_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Organization ID."}, "is_archived": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether deactivated."},
		"is_inactive": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether inactive through SCIM."},
	}
}
func teamDataAttributes(requireID bool) map[string]schema.Attribute {
	id := schema.StringAttribute{Computed: true, MarkdownDescription: "Team ID."}
	if requireID {
		id = schema.StringAttribute{Required: true, MarkdownDescription: "Team ID."}
	}
	return map[string]schema.Attribute{
		"id": id, "name": schema.StringAttribute{Computed: true, MarkdownDescription: "Team name."}, "description": schema.StringAttribute{Computed: true, MarkdownDescription: "Description."},
		"visibility": schema.StringAttribute{Computed: true, MarkdownDescription: "Visibility."}, "is_archived": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether archived."},
	}
}

type memberDataModel struct {
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
type membersDataModel struct {
	ID      types.String      `tfsdk:"id"`
	Members []memberDataModel `tfsdk:"members"`
}
type teamDataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Visibility  types.String `tfsdk:"visibility"`
	IsArchived  types.Bool   `tfsdk:"is_archived"`
}
type teamsDataModel struct {
	ID    types.String    `tfsdk:"id"`
	Teams []teamDataModel `tfsdk:"teams"`
}
type accountTypeDataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IsCustom    types.Bool   `tfsdk:"is_custom"`
}
type accountTypesDataModel struct {
	ID           types.String           `tfsdk:"id"`
	AccountTypes []accountTypeDataModel `tfsdk:"account_types"`
}
type userAttributeDataModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	DefaultValue types.String `tfsdk:"default_value"`
}
type userAttributesDataModel struct {
	ID             types.String             `tfsdk:"id"`
	UserAttributes []userAttributeDataModel `tfsdk:"user_attributes"`
}

func (d *identityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	switch d.kind {
	case "member":
		var state memberDataModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		value, err := d.client.GetMember(ctx, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Unable to read Sigma member", err.Error())
			return
		}
		state = memberData(value)
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	case "members":
		values, err := d.client.ListMembers(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Unable to list Sigma members", err.Error())
			return
		}
		state := membersDataModel{ID: types.StringValue("members")}
		for i := range values {
			state.Members = append(state.Members, memberData(&values[i]))
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	case "team":
		var state teamDataModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		value, err := d.client.GetTeam(ctx, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Unable to read Sigma team", err.Error())
			return
		}
		state = teamData(value)
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	case "teams":
		values, err := d.client.ListTeams(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Unable to list Sigma teams", err.Error())
			return
		}
		state := teamsDataModel{ID: types.StringValue("teams")}
		for i := range values {
			state.Teams = append(state.Teams, teamData(&values[i]))
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	case "account_types":
		values, err := d.client.ListAccountTypes(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Unable to list Sigma account types", err.Error())
			return
		}
		state := accountTypesDataModel{ID: types.StringValue("account-types")}
		for _, v := range values {
			state.AccountTypes = append(state.AccountTypes, accountTypeDataModel{ID: types.StringValue(v.AccountTypeID), Name: types.StringValue(v.AccountTypeName), Description: types.StringValue(v.Description), IsCustom: types.BoolValue(v.IsCustom)})
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	case "user_attributes":
		values, err := d.client.ListUserAttributes(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Unable to list Sigma user attributes", err.Error())
			return
		}
		state := userAttributesDataModel{ID: types.StringValue("user-attributes")}
		for i := range values {
			state.UserAttributes = append(state.UserAttributes, userAttributeData(&values[i]))
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}
func memberData(v *sigma.Member) memberDataModel {
	return memberDataModel{ID: types.StringValue(v.MemberID), Email: types.StringValue(v.Email), FirstName: types.StringValue(v.FirstName), LastName: types.StringValue(v.LastName), MemberType: types.StringValue(v.MemberType), UserKind: types.StringValue(v.UserKind), OrganizationID: types.StringValue(v.OrganizationID), IsArchived: types.BoolValue(v.IsArchived), IsInactive: types.BoolValue(v.IsInactive)}
}
func teamData(v *sigma.Team) teamDataModel {
	d := types.StringNull()
	if v.Description != nil {
		d = types.StringValue(*v.Description)
	}
	return teamDataModel{ID: types.StringValue(v.TeamID), Name: types.StringValue(v.Name), Description: d, Visibility: types.StringValue(v.Visibility), IsArchived: types.BoolValue(v.IsArchived)}
}
func userAttributeData(v *sigma.UserAttribute) userAttributeDataModel {
	d, x := types.StringNull(), types.StringNull()
	if v.Description != nil {
		d = types.StringValue(*v.Description)
	}
	if v.DefaultValue != nil {
		x = types.StringValue(v.DefaultValue.Val)
	}
	return userAttributeDataModel{ID: types.StringValue(v.UserAttributeID), Name: types.StringValue(v.Name), Description: d, DefaultValue: x}
}
