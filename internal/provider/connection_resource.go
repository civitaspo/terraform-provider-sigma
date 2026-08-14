package provider

import (
	"context"
	"fmt"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type connectionResource struct{ configuredResource }

var (
	_ resource.Resource                = (*connectionResource)(nil)
	_ resource.ResourceWithConfigure   = (*connectionResource)(nil)
	_ resource.ResourceWithImportState = (*connectionResource)(nil)
	_ resource.ResourceWithModifyPlan  = (*connectionResource)(nil)
)

var connectionTimeoutAttrTypes = map[string]attr.Type{
	"default":   types.Float64Type,
	"dashboard": types.Float64Type,
	"download":  types.Float64Type,
	"worksheet": types.Float64Type,
}

type connectionResourceModel struct {
	ID                       types.String         `tfsdk:"id"`
	Name                     types.String         `tfsdk:"name"`
	Type                     types.String         `tfsdk:"type"`
	DetailsJSON              jsontypes.Normalized `tfsdk:"details_json"`
	DescriptionJSON          jsontypes.Normalized `tfsdk:"description_json"`
	PoolSizesJSON            jsontypes.Normalized `tfsdk:"pool_sizes_json"`
	TimeoutSecs              types.Float64        `tfsdk:"timeout_secs"`
	UseFriendlyNames         types.Bool           `tfsdk:"use_friendly_names"`
	UseOauth                 types.Bool           `tfsdk:"use_oauth"`
	CredentialsWO            types.String         `tfsdk:"credentials_wo"`
	CredentialsWOVersion     types.Int64          `tfsdk:"credentials_wo_version"`
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

func NewConnectionResource() resource.Resource { return &connectionResource{} }

func (r *connectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection"
}

func (r *connectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}

func (r *connectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	preserveString := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	preserveBool := []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}
	preserveFloat := []planmodifier.Float64{float64planmodifier.UseStateForUnknown()}
	preserveList := []planmodifier.List{listplanmodifier.UseStateForUnknown()}
	preserveObject := []planmodifier.Object{objectplanmodifier.UseStateForUnknown()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Sigma warehouse connection. `details_json` is required and must contain the complete non-secret warehouse configuration for every create and PUT. Put write-only fields such as `password`, `serviceAccount`, and `clientSecret` in `credentials_wo`. Sigma's update connection API replaces warehouse details entirely, so any update after credentials were managed requires a strictly greater `credentials_wo_version` and a resupplied `credentials_wo`. Restore is not a Terraform attribute. GET does not return warehouse details, so import is ID-only: configuration must include complete `details_json` before Terraform can update.",
		Attributes: map[string]schema.Attribute{
			"id":                     schema.StringAttribute{Computed: true, PlanModifiers: preserveString, MarkdownDescription: "Connection ID."},
			"name":                   schema.StringAttribute{Required: true, MarkdownDescription: "Connection name."},
			"type":                   schema.StringAttribute{Computed: true, PlanModifiers: preserveString, MarkdownDescription: "Warehouse type returned by Sigma."},
			"details_json":           schema.StringAttribute{Required: true, CustomType: jsontypes.NormalizedType{}, MarkdownDescription: "Required JSON object containing non-secret warehouse-specific connection details. Every PUT sends this object plus any write-only credentials. GET does not return warehouse details; do not infer them from flattened GET metadata."},
			"description_json":       schema.StringAttribute{Optional: true, Computed: true, CustomType: jsontypes.NormalizedType{}, PlanModifiers: preserveString, MarkdownDescription: "JSON object accepted by Sigma as the connection description."},
			"pool_sizes_json":        schema.StringAttribute{Optional: true, Computed: true, CustomType: jsontypes.NormalizedType{}, PlanModifiers: preserveString, MarkdownDescription: "JSON object configuring connection pool sizes."},
			"timeout_secs":           schema.Float64Attribute{Optional: true, Computed: true, PlanModifiers: preserveFloat, MarkdownDescription: "Request `timeoutSecs`. When GET returns `timeout.default`, Terraform maps that value here because the documented semantics match."},
			"use_friendly_names":     schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: preserveBool, MarkdownDescription: "Request `useFriendlyNames`. GET reports this as `friendlyName`."},
			"use_oauth":              schema.BoolAttribute{Computed: true, PlanModifiers: preserveBool, MarkdownDescription: "Whether the connection uses OAuth, as returned by Sigma GET `useOauth`. Not settable; warehouse OAuth configuration remains in `details_json`."},
			"credentials_wo":         schema.StringAttribute{Optional: true, WriteOnly: true, Sensitive: true, MarkdownDescription: "Write-only JSON object merged into a copy of `details_json` for the outbound request only. Required together with `credentials_wo_version` on create/rotation. Never stored in state."},
			"credentials_wo_version": schema.Int64Attribute{Optional: true, MarkdownDescription: "Must be supplied with `credentials_wo`. After credentials have been managed, every PUT requires a known strictly greater version and a known write-only payload from configuration."},
			"organization_id":        schema.StringAttribute{Computed: true, PlanModifiers: preserveString, MarkdownDescription: "Organization ID."},
			"is_sample":              schema.BoolAttribute{Computed: true, PlanModifiers: preserveBool, MarkdownDescription: "Whether this is the Sigma sample data connection."},
			"is_audit_log":           schema.BoolAttribute{Computed: true, PlanModifiers: preserveBool, MarkdownDescription: "Whether audit logging is enabled."},
			"last_active_at":         schema.StringAttribute{Computed: true, PlanModifiers: preserveString, MarkdownDescription: "Last activity timestamp."},
			"created_by":             schema.StringAttribute{Computed: true, PlanModifiers: preserveString, MarkdownDescription: "Member ID that created the connection."},
			"updated_by":             schema.StringAttribute{Computed: true, PlanModifiers: preserveString, MarkdownDescription: "Member or process that last updated the connection."},
			"created_at":             schema.StringAttribute{Computed: true, PlanModifiers: preserveString, MarkdownDescription: "Creation timestamp."},
			"updated_at":             schema.StringAttribute{Computed: true, PlanModifiers: preserveString, MarkdownDescription: "Update timestamp."},
			"is_archived":            schema.BoolAttribute{Computed: true, PlanModifiers: preserveBool, MarkdownDescription: "Whether the connection is archived."},
			"account":                schema.StringAttribute{Computed: true, PlanModifiers: preserveString, MarkdownDescription: "Account associated with the connection. Response-only; not copied into PUT details."},
			"warehouse":              schema.StringAttribute{Computed: true, PlanModifiers: preserveString, MarkdownDescription: "Warehouse associated with the connection. Response-only; not copied into PUT details."},
			"user":                   schema.StringAttribute{Computed: true, PlanModifiers: preserveString, MarkdownDescription: "User associated with the connection. Response-only; not copied into PUT details."},
			"role":                   schema.StringAttribute{Computed: true, PlanModifiers: preserveString, MarkdownDescription: "Role used by the connection user. Response-only; not copied into PUT details."},
			"timeout": schema.SingleNestedAttribute{
				Computed:            true,
				PlanModifiers:       preserveObject,
				MarkdownDescription: "Complete GET `timeout` object (`default`, `dashboard`, `download`, `worksheet`).",
				Attributes: map[string]schema.Attribute{
					"default":   schema.Float64Attribute{Computed: true, MarkdownDescription: "Default timeout in seconds."},
					"dashboard": schema.Float64Attribute{Computed: true, MarkdownDescription: "Dashboard timeout in seconds."},
					"download":  schema.Float64Attribute{Computed: true, MarkdownDescription: "Download timeout in seconds."},
					"worksheet": schema.Float64Attribute{Computed: true, MarkdownDescription: "Worksheet timeout in seconds."},
				},
			},
			"write_access":                      schema.BoolAttribute{Computed: true, PlanModifiers: preserveBool, MarkdownDescription: "Whether write access is enabled."},
			"friendly_name":                     schema.BoolAttribute{Computed: true, PlanModifiers: preserveBool, MarkdownDescription: "GET `friendlyName`."},
			"writebacks_json":                   schema.StringAttribute{Computed: true, CustomType: jsontypes.NormalizedType{}, PlanModifiers: preserveString, MarkdownDescription: "Non-OAuth write-back configuration JSON."},
			"writeback_schemas_json":            schema.StringAttribute{Computed: true, CustomType: jsontypes.NormalizedType{}, PlanModifiers: preserveString, MarkdownDescription: "OAuth write-back schema configuration JSON."},
			"input_table_audit_log_schema_json": schema.StringAttribute{Computed: true, CustomType: jsontypes.NormalizedType{}, PlanModifiers: preserveString, MarkdownDescription: "Input table write-ahead log schema JSON."},
			"materialization_warehouse":         schema.StringAttribute{Computed: true, PlanModifiers: preserveString, MarkdownDescription: "Warehouse used for materialization jobs."},
			"exports_warehouse":                 schema.StringAttribute{Computed: true, PlanModifiers: preserveString, MarkdownDescription: "Warehouse used for export jobs."},
			"oauth_metadata_url":                schema.StringAttribute{Computed: true, PlanModifiers: preserveString, MarkdownDescription: "Connection-level OAuth metadata URL."},
			"oauth_client_id":                   schema.StringAttribute{Computed: true, PlanModifiers: preserveString, MarkdownDescription: "Connection-level OAuth client ID."},
			"oauth_scopes":                      schema.ListAttribute{Computed: true, ElementType: types.StringType, PlanModifiers: preserveList, MarkdownDescription: "Connection-level OAuth scopes."},
			"oauth_idp_type":                    schema.StringAttribute{Computed: true, PlanModifiers: preserveString, MarkdownDescription: "OAuth provider type."},
			"oauth_use_pkce":                    schema.BoolAttribute{Computed: true, PlanModifiers: preserveBool, MarkdownDescription: "Whether connection-level OAuth uses PKCE."},
			"oauth_use_jwt":                     schema.BoolAttribute{Computed: true, PlanModifiers: preserveBool, MarkdownDescription: "Whether connection-level OAuth uses JWT bearer tokens."},
			"oauth_audience":                    schema.StringAttribute{Computed: true, PlanModifiers: preserveString, MarkdownDescription: "OAuth federation audience."},
			"is_independent_oauth":              schema.BoolAttribute{Computed: true, PlanModifiers: preserveBool, MarkdownDescription: "Whether the connection uses connection-level OAuth."},
			"user_attributes_json":              schema.StringAttribute{Computed: true, CustomType: jsontypes.NormalizedType{}, PlanModifiers: preserveString, MarkdownDescription: "User attributes JSON associated with the connection."},
			"role_switching":                    schema.StringAttribute{Computed: true, PlanModifiers: preserveString, MarkdownDescription: "Snowflake OAuth role-switching setting."},
		},
	}
}

func (r *connectionResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	var plan connectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var config connectionResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(writeOnlyVersionPair(config.CredentialsWO, plan.CredentialsWOVersion, "credentials_wo", "credentials_wo_version")...)
	if req.State.Raw.IsNull() {
		return
	}
	var state connectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !state.CredentialsWOVersion.IsNull() && !state.CredentialsWOVersion.IsUnknown() && (plan.CredentialsWOVersion.IsNull() || plan.CredentialsWOVersion.IsUnknown()) {
		resp.Diagnostics.AddError(
			"Cannot remove credentials_wo_version",
			"After credentials_wo has been managed, credentials_wo_version cannot be removed.",
		)
	}
}

func managedWriteOnlyUpdate(stateVersion, planVersion types.Int64, configPayload types.String, payloadName, versionName string) diag.Diagnostics {
	var diags diag.Diagnostics
	if stateVersion.IsNull() || stateVersion.IsUnknown() {
		return diags
	}
	if planVersion.IsNull() || planVersion.IsUnknown() {
		diags.AddError(
			"Cannot remove "+versionName,
			"After "+payloadName+" has been managed, "+versionName+" cannot be removed.",
		)
		return diags
	}
	if planVersion.Equal(stateVersion) {
		diags.AddError(
			"Cannot update without resending "+payloadName,
			"Sigma replaces secret-bearing fields on update. Increment `"+versionName+"` and supply `"+payloadName+"` on every update after secrets were previously managed.",
		)
		return diags
	}
	if planVersion.ValueInt64() <= stateVersion.ValueInt64() {
		diags.AddError(
			"Invalid "+versionName,
			versionName+" must be a known value strictly greater than the previously applied version.",
		)
		return diags
	}
	if configPayload.IsNull() || configPayload.IsUnknown() {
		diags.AddError(
			"Cannot update without resending "+payloadName,
			"A bumped `"+versionName+"` requires a known `"+payloadName+"` from configuration.",
		)
	}
	return diags
}

func connectionInput(plan *connectionResourceModel, credentials types.String) (sigma.ConnectionInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	detailsObject, _, detailsDiags := knownNormalizedObject(plan.DetailsJSON, "details_json")
	diags.Append(detailsDiags...)
	if diags.HasError() {
		return sigma.ConnectionInput{}, diags
	}
	details, mergeDiags := mergeObjectWithWriteOnly(detailsObject, credentials, "credentials_wo")
	diags.Append(mergeDiags...)
	description, descriptionDiags := optionalNormalizedJSON(plan.DescriptionJSON, "description_json")
	diags.Append(descriptionDiags...)
	poolSizes, poolDiags := optionalNormalizedJSON(plan.PoolSizesJSON, "pool_sizes_json")
	diags.Append(poolDiags...)
	if diags.HasError() {
		return sigma.ConnectionInput{}, diags
	}
	input := sigma.ConnectionInput{
		Details:     details,
		Name:        plan.Name.ValueString(),
		Description: description,
		PoolSizes:   poolSizes,
	}
	if !plan.TimeoutSecs.IsNull() && !plan.TimeoutSecs.IsUnknown() {
		value := plan.TimeoutSecs.ValueFloat64()
		input.TimeoutSecs = &value
	}
	if !plan.UseFriendlyNames.IsNull() && !plan.UseFriendlyNames.IsUnknown() {
		value := plan.UseFriendlyNames.ValueBool()
		input.UseFriendlyNames = &value
	}
	return input, diags
}

func setConnection(ctx context.Context, state *connectionResourceModel, value *sigma.Connection) diag.Diagnostics {
	var diags diag.Diagnostics
	priorDetails := state.DetailsJSON
	priorVersion := state.CredentialsWOVersion
	state.ID = types.StringValue(value.ConnectionID)
	state.Name = types.StringValue(value.Name)
	state.Type = types.StringValue(value.Type)
	state.DetailsJSON = priorDetails
	state.CredentialsWOVersion = priorVersion
	if len(value.Description) > 0 && string(value.Description) != "null" {
		state.DescriptionJSON = normalizedFromRaw(value.Description)
	}
	if len(value.PoolSizes) > 0 && string(value.PoolSizes) != "null" {
		state.PoolSizesJSON = normalizedFromRaw(value.PoolSizes)
	}
	if value.TimeoutSecs != nil {
		state.TimeoutSecs = types.Float64Value(*value.TimeoutSecs)
	}
	state.UseFriendlyNames = types.BoolValue(value.FriendlyName)
	state.FriendlyName = types.BoolValue(value.FriendlyName)
	if value.UseOauth != nil {
		state.UseOauth = types.BoolValue(*value.UseOauth)
	} else {
		state.UseOauth = types.BoolNull()
	}
	state.OrganizationID = types.StringValue(value.OrganizationID)
	state.IsSample = types.BoolPointerValue(value.IsSample)
	state.IsAuditLog = types.BoolPointerValue(value.IsAuditLog)
	state.LastActiveAt = types.StringValue(value.LastActiveAt)
	state.CreatedBy = types.StringValue(value.CreatedBy)
	state.UpdatedBy = types.StringValue(value.UpdatedBy)
	state.CreatedAt = types.StringValue(value.CreatedAt)
	state.UpdatedAt = types.StringValue(value.UpdatedAt)
	state.IsArchived = types.BoolPointerValue(value.IsArchived)
	state.Account = types.StringValue(value.Account)
	state.Warehouse = types.StringValue(value.Warehouse)
	state.User = types.StringValue(value.User)
	state.Role = types.StringValue(value.Role)
	state.WriteAccess = types.BoolPointerValue(value.WriteAccess)
	state.WritebacksJSON = normalizedFromRaw(value.Writebacks)
	state.WritebackSchemasJSON = normalizedFromRaw(value.WritebackSchemas)
	state.InputTableAuditLogSchema = normalizedFromRaw(value.InputTableAuditLogSchema)
	state.MaterializationWarehouse = types.StringValue(value.MaterializationWarehouse)
	state.ExportsWarehouse = types.StringValue(value.ExportsWarehouse)
	state.OauthMetadataURL = types.StringValue(value.OauthMetadataURL)
	state.OauthClientID = types.StringValue(value.OauthClientID)
	state.OauthIdpType = types.StringValue(value.OauthIdpType)
	state.OauthUsePkce = types.BoolPointerValue(value.OauthUsePkce)
	state.OauthUseJwt = types.BoolPointerValue(value.OauthUseJwt)
	state.OauthAudience = types.StringValue(value.OauthAudience)
	state.IsIndependentOAuth = types.BoolPointerValue(value.IsIndependentOAuth)
	state.UserAttributesJSON = normalizedFromRaw(value.UserAttributes)
	state.RoleSwitching = types.StringValue(value.RoleSwitching)
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
	return diags
}

func (r *connectionResource) test(ctx context.Context, id string, resp interface{ AddWarning(string, string) }) {
	result, err := r.client.TestConnection(ctx, id)
	if err != nil {
		resp.AddWarning("Sigma connection test failed", err.Error())
		return
	}
	if result.Read == "FAILED" || result.Write == "FAILED" {
		resp.AddWarning("Sigma connection test failed", fmt.Sprintf("Connection test returned read=%s, write=%s.", result.Read, result.Write))
	}
}

func (r *connectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan connectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var config connectionResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(writeOnlyVersionPair(config.CredentialsWO, plan.CredentialsWOVersion, "credentials_wo", "credentials_wo_version")...)
	if resp.Diagnostics.HasError() {
		return
	}
	input, diags := connectionInput(&plan, config.CredentialsWO)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.CreateConnection(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma connection", err.Error())
		return
	}
	resp.Diagnostics.Append(setConnection(ctx, &plan, value)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.test(ctx, value.ConnectionID, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *connectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state connectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.GetConnection(ctx, state.ID.ValueString())
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma connection", err.Error())
		return
	}
	resp.Diagnostics.Append(setConnection(ctx, &state, value)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *connectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan connectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state connectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	var config connectionResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(managedWriteOnlyUpdate(state.CredentialsWOVersion, plan.CredentialsWOVersion, config.CredentialsWO, "credentials_wo", "credentials_wo_version")...)
	if resp.Diagnostics.HasError() {
		return
	}
	input, diags := connectionInput(&plan, config.CredentialsWO)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.UpdateConnection(ctx, state.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Sigma connection", err.Error())
		return
	}
	resp.Diagnostics.Append(setConnection(ctx, &plan, value)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.test(ctx, value.ConnectionID, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *connectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state connectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteConnection(ctx, state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma connection", err.Error())
	}
}

func (r *connectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importPassthrough(ctx, req, resp)
}
