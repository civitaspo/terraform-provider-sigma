package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*memberDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*memberDataSource)(nil)
)

type memberDataSource struct{ configuredDataSource }

type memberDataModel struct {
	ID             types.String `tfsdk:"id"`
	Email          types.String `tfsdk:"email"`
	FirstName      types.String `tfsdk:"first_name"`
	LastName       types.String `tfsdk:"last_name"`
	MemberType     types.String `tfsdk:"member_type"`
	UserKind       types.String `tfsdk:"user_kind"`
	OrganizationID types.String `tfsdk:"organization_id"`
	HomeFolderID   types.String `tfsdk:"home_folder_id"`
	IsArchived     types.Bool   `tfsdk:"is_archived"`
	IsInactive     types.Bool   `tfsdk:"is_inactive"`
}

func NewMemberDataSource() datasource.DataSource { return &memberDataSource{} }

func (d *memberDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_member"
}

func (d *memberDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}

func (d *memberDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Retrieves a Sigma member by ID.", Attributes: memberDataAttributes(true)}
}

func (d *memberDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state memberDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, idDiags := knownString(state.ID, "id")
	resp.Diagnostics.Append(idDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := d.client.GetMember(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma member", err.Error())
		return
	}
	state = memberData(value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
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
		"is_inactive":    schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether inactive through SCIM."},
		"home_folder_id": schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the member's My Documents folder."},
	}
}

func memberData(v *sigma.Member) memberDataModel {
	homeFolderID := types.StringNull()
	if v.HomeFolderID != "" {
		homeFolderID = types.StringValue(v.HomeFolderID)
	}
	return memberDataModel{ID: types.StringValue(v.MemberID), Email: types.StringValue(v.Email), FirstName: types.StringValue(v.FirstName), LastName: types.StringValue(v.LastName), MemberType: types.StringValue(v.MemberType), UserKind: types.StringValue(v.UserKind), OrganizationID: types.StringValue(v.OrganizationID), HomeFolderID: homeFolderID, IsArchived: types.BoolValue(v.IsArchived), IsInactive: types.BoolValue(v.IsInactive)}
}
