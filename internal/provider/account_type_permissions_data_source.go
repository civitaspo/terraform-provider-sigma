package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*accountTypePermissionsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*accountTypePermissionsDataSource)(nil)
)

type accountTypePermissionsDataSource struct{ configuredDataSource }

type accountTypePermissionModel struct {
	Permission  types.String `tfsdk:"permission"`
	Description types.String `tfsdk:"description"`
}

type accountTypePermissionsModel struct {
	ID            types.String                 `tfsdk:"id"`
	AccountTypeID types.String                 `tfsdk:"account_type_id"`
	Permissions   []accountTypePermissionModel `tfsdk:"permissions"`
}

func NewAccountTypePermissionsDataSource() datasource.DataSource {
	return &accountTypePermissionsDataSource{}
}

func (d *accountTypePermissionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account_type_permissions"
}

func (d *accountTypePermissionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}

func (d *accountTypePermissionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists feature permissions for a Sigma account type. The API returns a JSON array rather than a paginated envelope." + listCollectionNotice,
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
			"account_type_id": schema.StringAttribute{Required: true, MarkdownDescription: "Account type ID."},
			"permissions": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Permissions in API order.", NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
				"permission":  schema.StringAttribute{Computed: true, MarkdownDescription: "Permission name (`permission`)."},
				"description": schema.StringAttribute{Computed: true, MarkdownDescription: "Human-readable description (`description`)."},
			}}},
		},
	}
}

func (d *accountTypePermissionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state accountTypePermissionsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, idDiags := knownString(state.AccountTypeID, "account_type_id")
	resp.Diagnostics.Append(idDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	values, err := d.client.ListAccountTypePermissions(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma account type permissions", err.Error())
		return
	}
	state.ID = types.StringValue(id)
	state.Permissions = make([]accountTypePermissionModel, 0, len(values))
	for _, value := range values {
		state.Permissions = append(state.Permissions, accountTypePermissionModel{
			Permission: types.StringValue(value.Permission), Description: types.StringValue(value.Description),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
