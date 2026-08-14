package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*connectionDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*connectionDataSource)(nil)
)

type connectionDataSource struct{ configuredDataSource }

type connectionDataModel struct {
	ID                       types.String         `tfsdk:"id"`
	Name                     types.String         `tfsdk:"name"`
	Type                     types.String         `tfsdk:"type"`
	DescriptionJSON          jsontypes.Normalized `tfsdk:"description_json"`
	PoolSizesJSON            jsontypes.Normalized `tfsdk:"pool_sizes_json"`
	TimeoutSecs              types.Float64        `tfsdk:"timeout_secs"`
	UseFriendlyNames         types.Bool           `tfsdk:"use_friendly_names"`
	UseOauth                 types.Bool           `tfsdk:"use_oauth"`
	OrganizationID           types.String         `tfsdk:"organization_id"`
	IsSample                 types.Bool           `tfsdk:"is_sample"`
	IsAuditLog               types.Bool           `tfsdk:"is_audit_log"`
	LastActiveAt             types.String         `tfsdk:"last_active_at"`
	CreatedBy                types.String         `tfsdk:"created_by"`
	UpdatedBy                types.String         `tfsdk:"updated_by"`
	CreatedAt                types.String         `tfsdk:"created_at"`
	UpdatedAt                types.String         `tfsdk:"updated_at"`
	IsArchived               types.Bool           `tfsdk:"is_archived"`
	Account                  types.String         `tfsdk:"account"`
	Warehouse                types.String         `tfsdk:"warehouse"`
	User                     types.String         `tfsdk:"user"`
	Role                     types.String         `tfsdk:"role"`
	Timeout                  types.Object         `tfsdk:"timeout"`
	WriteAccess              types.Bool           `tfsdk:"write_access"`
	FriendlyName             types.Bool           `tfsdk:"friendly_name"`
	WritebacksJSON           jsontypes.Normalized `tfsdk:"writebacks_json"`
	WritebackSchemasJSON     jsontypes.Normalized `tfsdk:"writeback_schemas_json"`
	InputTableAuditLogSchema jsontypes.Normalized `tfsdk:"input_table_audit_log_schema_json"`
	MaterializationWarehouse types.String         `tfsdk:"materialization_warehouse"`
	ExportsWarehouse         types.String         `tfsdk:"exports_warehouse"`
	OauthMetadataURL         types.String         `tfsdk:"oauth_metadata_url"`
	OauthClientID            types.String         `tfsdk:"oauth_client_id"`
	OauthScopes              types.List           `tfsdk:"oauth_scopes"`
	OauthIdpType             types.String         `tfsdk:"oauth_idp_type"`
	OauthUsePkce             types.Bool           `tfsdk:"oauth_use_pkce"`
	OauthUseJwt              types.Bool           `tfsdk:"oauth_use_jwt"`
	OauthAudience            types.String         `tfsdk:"oauth_audience"`
	IsIndependentOAuth       types.Bool           `tfsdk:"is_independent_oauth"`
	UserAttributesJSON       jsontypes.Normalized `tfsdk:"user_attributes_json"`
	RoleSwitching            types.String         `tfsdk:"role_switching"`
}

func NewConnectionDataSource() datasource.DataSource { return &connectionDataSource{} }

func (d *connectionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection"
}

func (d *connectionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}

func (d *connectionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Retrieves a Sigma connection by ID. Credentials are never exposed.", Attributes: connectionDataAttributes(true)}
}

func (d *connectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state connectionDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, idDiags := knownString(state.ID, "id")
	resp.Diagnostics.Append(idDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := d.client.GetConnection(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma connection", err.Error())
		return
	}
	mapped, diags := connectionData(ctx, value)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, mapped)...)
}

func connectionDataAttributes(requireID bool) map[string]schema.Attribute {
	id := schema.StringAttribute{Computed: true, MarkdownDescription: "Connection ID."}
	if requireID {
		id = schema.StringAttribute{Required: true, MarkdownDescription: "Connection ID."}
	}
	return map[string]schema.Attribute{
		"id":                 id,
		"name":               schema.StringAttribute{Computed: true, MarkdownDescription: "Connection name."},
		"type":               schema.StringAttribute{Computed: true, MarkdownDescription: "Warehouse type."},
		"description_json":   schema.StringAttribute{Computed: true, CustomType: jsontypes.NormalizedType{}, MarkdownDescription: "Description JSON returned by Sigma."},
		"pool_sizes_json":    schema.StringAttribute{Computed: true, CustomType: jsontypes.NormalizedType{}, MarkdownDescription: "Pool sizes JSON returned by Sigma."},
		"timeout_secs":       schema.Float64Attribute{Computed: true, MarkdownDescription: "Request-facing timeout seconds mapped from `timeout.default` when present."},
		"use_friendly_names": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether friendly names are enabled (`friendlyName`)."},
		"use_oauth":          schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the connection uses OAuth (`useOauth`)."},
		"organization_id":    schema.StringAttribute{Computed: true, MarkdownDescription: "Organization ID."},
		"is_sample":          schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this is the Sigma sample data connection."},
		"is_audit_log":       schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether audit logging is enabled."},
		"last_active_at":     schema.StringAttribute{Computed: true, MarkdownDescription: "Last activity timestamp."},
		"created_by":         schema.StringAttribute{Computed: true, MarkdownDescription: "Member ID that created the connection."},
		"updated_by":         schema.StringAttribute{Computed: true, MarkdownDescription: "Member or process that last updated the connection."},
		"created_at":         schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp."},
		"updated_at":         schema.StringAttribute{Computed: true, MarkdownDescription: "Update timestamp."},
		"is_archived":        schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the connection is archived."},
		"account":            schema.StringAttribute{Computed: true, MarkdownDescription: "Account associated with the connection."},
		"warehouse":          schema.StringAttribute{Computed: true, MarkdownDescription: "Warehouse associated with the connection."},
		"user":               schema.StringAttribute{Computed: true, MarkdownDescription: "User associated with the connection."},
		"role":               schema.StringAttribute{Computed: true, MarkdownDescription: "Role used by the connection user."},
		"timeout": schema.SingleNestedAttribute{Computed: true, MarkdownDescription: "Complete GET `timeout` object.", Attributes: map[string]schema.Attribute{
			"default":   schema.Float64Attribute{Computed: true, MarkdownDescription: "Default timeout in seconds."},
			"dashboard": schema.Float64Attribute{Computed: true, MarkdownDescription: "Dashboard timeout in seconds."},
			"download":  schema.Float64Attribute{Computed: true, MarkdownDescription: "Download timeout in seconds."},
			"worksheet": schema.Float64Attribute{Computed: true, MarkdownDescription: "Worksheet timeout in seconds."},
		}},
		"write_access":                      schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether write access is enabled."},
		"friendly_name":                     schema.BoolAttribute{Computed: true, MarkdownDescription: "GET `friendlyName`."},
		"writebacks_json":                   schema.StringAttribute{Computed: true, CustomType: jsontypes.NormalizedType{}, MarkdownDescription: "Non-OAuth write-back configuration JSON."},
		"writeback_schemas_json":            schema.StringAttribute{Computed: true, CustomType: jsontypes.NormalizedType{}, MarkdownDescription: "OAuth write-back schema configuration JSON."},
		"input_table_audit_log_schema_json": schema.StringAttribute{Computed: true, CustomType: jsontypes.NormalizedType{}, MarkdownDescription: "Input table write-ahead log schema JSON."},
		"materialization_warehouse":         schema.StringAttribute{Computed: true, MarkdownDescription: "Warehouse used for materialization jobs."},
		"exports_warehouse":                 schema.StringAttribute{Computed: true, MarkdownDescription: "Warehouse used for export jobs."},
		"oauth_metadata_url":                schema.StringAttribute{Computed: true, MarkdownDescription: "Connection-level OAuth metadata URL."},
		"oauth_client_id":                   schema.StringAttribute{Computed: true, MarkdownDescription: "Connection-level OAuth client ID. This is not a client secret."},
		"oauth_scopes":                      schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Connection-level OAuth scopes."},
		"oauth_idp_type":                    schema.StringAttribute{Computed: true, MarkdownDescription: "OAuth provider type."},
		"oauth_use_pkce":                    schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether connection-level OAuth uses PKCE."},
		"oauth_use_jwt":                     schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether connection-level OAuth uses JWT bearer tokens."},
		"oauth_audience":                    schema.StringAttribute{Computed: true, MarkdownDescription: "OAuth federation audience."},
		"is_independent_oauth":              schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the connection uses connection-level OAuth."},
		"user_attributes_json":              schema.StringAttribute{Computed: true, CustomType: jsontypes.NormalizedType{}, MarkdownDescription: "User attributes JSON associated with the connection."},
		"role_switching":                    schema.StringAttribute{Computed: true, MarkdownDescription: "Snowflake OAuth role-switching setting."},
	}
}

func connectionData(ctx context.Context, value *sigma.Connection) (connectionDataModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	state := connectionDataModel{
		ID: types.StringValue(value.ConnectionID), Name: types.StringValue(value.Name), Type: types.StringValue(value.Type),
		DescriptionJSON: normalizedFromRaw(value.Description), PoolSizesJSON: normalizedFromRaw(value.PoolSizes),
		UseFriendlyNames: types.BoolValue(value.FriendlyName), FriendlyName: types.BoolValue(value.FriendlyName),
		OrganizationID: types.StringValue(value.OrganizationID), IsSample: types.BoolPointerValue(value.IsSample),
		IsAuditLog: types.BoolPointerValue(value.IsAuditLog), LastActiveAt: types.StringValue(value.LastActiveAt),
		CreatedBy: types.StringValue(value.CreatedBy), UpdatedBy: types.StringValue(value.UpdatedBy),
		CreatedAt: types.StringValue(value.CreatedAt), UpdatedAt: types.StringValue(value.UpdatedAt),
		IsArchived: types.BoolPointerValue(value.IsArchived), Account: types.StringValue(value.Account),
		Warehouse: types.StringValue(value.Warehouse), User: types.StringValue(value.User), Role: types.StringValue(value.Role),
		WriteAccess: types.BoolPointerValue(value.WriteAccess), WritebacksJSON: normalizedFromRaw(value.Writebacks),
		WritebackSchemasJSON: normalizedFromRaw(value.WritebackSchemas), InputTableAuditLogSchema: normalizedFromRaw(value.InputTableAuditLogSchema),
		MaterializationWarehouse: types.StringValue(value.MaterializationWarehouse), ExportsWarehouse: types.StringValue(value.ExportsWarehouse),
		OauthMetadataURL: types.StringValue(value.OauthMetadataURL), OauthClientID: types.StringValue(value.OauthClientID),
		OauthIdpType: types.StringValue(value.OauthIdpType), OauthUsePkce: types.BoolPointerValue(value.OauthUsePkce),
		OauthUseJwt: types.BoolPointerValue(value.OauthUseJwt), OauthAudience: types.StringValue(value.OauthAudience),
		IsIndependentOAuth: types.BoolPointerValue(value.IsIndependentOAuth), UserAttributesJSON: normalizedFromRaw(value.UserAttributes),
		RoleSwitching: types.StringValue(value.RoleSwitching), UseOauth: types.BoolPointerValue(value.UseOauth),
	}
	if value.TimeoutSecs != nil {
		state.TimeoutSecs = types.Float64Value(*value.TimeoutSecs)
	} else {
		state.TimeoutSecs = types.Float64Null()
	}
	if value.Timeout != nil {
		timeout, timeoutDiags := types.ObjectValue(connectionTimeoutAttrTypes, map[string]attr.Value{
			"default":   types.Float64Value(value.Timeout.Default),
			"dashboard": types.Float64PointerValue(value.Timeout.Dashboard),
			"download":  types.Float64PointerValue(value.Timeout.Download),
			"worksheet": types.Float64PointerValue(value.Timeout.Worksheet),
		})
		diags.Append(timeoutDiags...)
		state.Timeout = timeout
	} else {
		state.Timeout = types.ObjectNull(connectionTimeoutAttrTypes)
	}
	scopes := value.OauthScopes
	if scopes == nil {
		scopes = make([]string, 0)
	}
	list, listDiags := types.ListValueFrom(ctx, types.StringType, scopes)
	diags.Append(listDiags...)
	state.OauthScopes = list
	return state, diags
}
