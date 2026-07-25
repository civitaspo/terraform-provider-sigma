package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func canonicalJSON(raw []byte) (string, error) {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", err
	}
	encoded, err := json.Marshal(value)
	return string(encoded), err
}

func rawJSON(value types.String) (json.RawMessage, error) {
	if value.IsNull() || value.IsUnknown() || strings.TrimSpace(value.ValueString()) == "" {
		return nil, nil
	}
	canonical, err := canonicalJSON([]byte(value.ValueString()))
	return json.RawMessage(canonical), err
}

func mergeJSON(base types.String, overlay types.String) (json.RawMessage, error) {
	var left map[string]any
	if base.IsNull() || base.IsUnknown() {
		left = map[string]any{}
	} else if err := json.Unmarshal([]byte(base.ValueString()), &left); err != nil {
		return nil, err
	}
	if !overlay.IsNull() && !overlay.IsUnknown() {
		var right map[string]any
		if err := json.Unmarshal([]byte(overlay.ValueString()), &right); err != nil {
			return nil, err
		}
		for key, value := range right {
			left[key] = value
		}
	}
	return json.Marshal(left)
}

func jsonString(value json.RawMessage) types.String {
	if len(value) == 0 || string(value) == "null" {
		return types.StringNull()
	}
	canonical, err := canonicalJSON(value)
	if err != nil {
		return types.StringNull()
	}
	return types.StringValue(canonical)
}

func jsonValuesEqual(left, right types.String) bool {
	if left.Equal(right) {
		return true
	}
	if left.IsNull() || right.IsNull() || left.IsUnknown() || right.IsUnknown() {
		return false
	}
	leftCanonical, leftErr := canonicalJSON([]byte(left.ValueString()))
	rightCanonical, rightErr := canonicalJSON([]byte(right.ValueString()))
	if leftErr != nil || rightErr != nil {
		return false
	}
	return leftCanonical == rightCanonical
}

type connectionResource struct{ configuredResource }
type connectionResourceModel struct {
	ID                   types.String  `tfsdk:"id"`
	Name                 types.String  `tfsdk:"name"`
	Type                 types.String  `tfsdk:"type"`
	DetailsJSON          types.String  `tfsdk:"details_json"`
	DescriptionJSON      types.String  `tfsdk:"description_json"`
	PoolSizesJSON        types.String  `tfsdk:"pool_sizes_json"`
	TimeoutSecs          types.Float64 `tfsdk:"timeout_secs"`
	UseFriendlyNames     types.Bool    `tfsdk:"use_friendly_names"`
	CredentialsWO        types.String  `tfsdk:"credentials_wo"`
	CredentialsWOVersion types.Int64   `tfsdk:"credentials_wo_version"`
}

func NewConnectionResource() resource.Resource { return &connectionResource{} }
func (r *connectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection"
}
func (r *connectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *connectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Sigma warehouse connection. `details_json` is polymorphic by warehouse `type`; put write-only fields such as `password`, `serviceAccount`, and `clientSecret` in `credentials_wo`. Sigma's update connection API replaces warehouse details entirely, so any update that previously sent `credentials_wo` requires incrementing `credentials_wo_version` (and resupplying `credentials_wo`) to avoid clearing authentication. Sigma's get endpoint does not return warehouse details, so imported resources cannot recover them.",
		Attributes: map[string]schema.Attribute{
			"id":                     schema.StringAttribute{Computed: true, MarkdownDescription: "Connection ID."},
			"name":                   schema.StringAttribute{Required: true, MarkdownDescription: "Connection name."},
			"type":                   schema.StringAttribute{Computed: true, MarkdownDescription: "Warehouse type returned by Sigma."},
			"details_json":           schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "JSON object containing non-secret warehouse-specific connection details."},
			"description_json":       schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "JSON object accepted by Sigma as the connection description."},
			"pool_sizes_json":        schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "JSON object configuring connection pool sizes."},
			"timeout_secs":           schema.Float64Attribute{Optional: true, Computed: true, MarkdownDescription: "Connection timeout in seconds."},
			"use_friendly_names":     schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Whether friendly names are enabled."},
			"credentials_wo":         schema.StringAttribute{Optional: true, WriteOnly: true, Sensitive: true, MarkdownDescription: "Write-only JSON object merged into `details_json` before create or update. Required whenever `credentials_wo_version` changes."},
			"credentials_wo_version": schema.Int64Attribute{Optional: true, MarkdownDescription: "Set on create when using `credentials_wo`, and increment on every update that should retain or rotate warehouse credentials. Sigma PUT replaces details, so updates without a version bump are rejected when credentials were previously managed."},
		},
	}
}
func connectionInput(plan *connectionResourceModel) (sigma.ConnectionInput, error) {
	details, err := mergeJSON(plan.DetailsJSON, plan.CredentialsWO)
	if err != nil {
		return sigma.ConnectionInput{}, fmt.Errorf("decode connection details JSON: %w", err)
	}
	if len(details) == 0 || string(details) == "{}" {
		return sigma.ConnectionInput{}, fmt.Errorf("details_json or credentials_wo must define connection details")
	}
	description, err := rawJSON(plan.DescriptionJSON)
	if err != nil {
		return sigma.ConnectionInput{}, fmt.Errorf("decode description_json: %w", err)
	}
	poolSizes, err := rawJSON(plan.PoolSizesJSON)
	if err != nil {
		return sigma.ConnectionInput{}, fmt.Errorf("decode pool_sizes_json: %w", err)
	}
	input := sigma.ConnectionInput{Details: details, Name: plan.Name.ValueString(), Description: description, PoolSizes: poolSizes}
	if !plan.TimeoutSecs.IsNull() && !plan.TimeoutSecs.IsUnknown() {
		value := plan.TimeoutSecs.ValueFloat64()
		input.TimeoutSecs = &value
	}
	if !plan.UseFriendlyNames.IsNull() && !plan.UseFriendlyNames.IsUnknown() {
		value := plan.UseFriendlyNames.ValueBool()
		input.UseFriendlyNames = &value
	}
	return input, nil
}
func setConnection(state *connectionResourceModel, value *sigma.Connection) {
	state.ID = types.StringValue(value.ConnectionID)
	state.Name = types.StringValue(value.Name)
	state.Type = types.StringValue(value.Type)
	if len(value.Description) > 0 && string(value.Description) != "null" {
		state.DescriptionJSON = jsonString(value.Description)
	}
	if len(value.PoolSizes) > 0 && string(value.PoolSizes) != "null" {
		state.PoolSizesJSON = jsonString(value.PoolSizes)
	}
	if value.TimeoutSecs != nil {
		state.TimeoutSecs = types.Float64Value(*value.TimeoutSecs)
	}
	state.UseFriendlyNames = types.BoolValue(value.FriendlyName)
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
	plan.CredentialsWO = config.CredentialsWO
	input, err := connectionInput(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Sigma connection configuration", err.Error())
		return
	}
	value, err := r.client.CreateConnection(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma connection", err.Error())
		return
	}
	setConnection(&plan, value)
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
	setConnection(&state, value)
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
	managedCredentials := !state.CredentialsWOVersion.IsNull() && !state.CredentialsWOVersion.IsUnknown()
	resendingCredentials := !plan.CredentialsWOVersion.Equal(state.CredentialsWOVersion)
	// Sigma PUT replaces warehouse details entirely. Never send details without
	// write-only credentials once they have been managed via credentials_wo_version.
	if managedCredentials && !resendingCredentials {
		resp.Diagnostics.AddError(
			"Cannot update Sigma connection without resending credentials",
			"Sigma's update connection API replaces warehouse details entirely. Increment `credentials_wo_version` and supply `credentials_wo` on every update so warehouse authentication is not cleared.",
		)
		return
	}
	if resendingCredentials {
		plan.CredentialsWO = config.CredentialsWO
	}
	input, err := connectionInput(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Sigma connection configuration", err.Error())
		return
	}
	value, err := r.client.UpdateConnection(ctx, state.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Sigma connection", err.Error())
		return
	}
	setConnection(&plan, value)
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

type connectionGrantResource struct {
	configuredResource
	path bool
}
type connectionGrantModel struct {
	ID               types.String `tfsdk:"id"`
	ConnectionID     types.String `tfsdk:"connection_id"`
	ConnectionPathID types.String `tfsdk:"connection_path_id"`
	MemberID         types.String `tfsdk:"member_id"`
	TeamID           types.String `tfsdk:"team_id"`
	Permission       types.String `tfsdk:"permission"`
}

func NewConnectionGrantResource() resource.Resource { return &connectionGrantResource{} }
func NewConnectionPathGrantResource() resource.Resource {
	return &connectionGrantResource{path: true}
}
func (r *connectionGrantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	if r.path {
		resp.TypeName = req.ProviderTypeName + "_connection_path_grant"
	} else {
		resp.TypeName = req.ProviderTypeName + "_connection_grant"
	}
}
func (r *connectionGrantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *connectionGrantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	attrs := map[string]schema.Attribute{
		"id":                 schema.StringAttribute{Computed: true, MarkdownDescription: "Grant ID."},
		"member_id":          schema.StringAttribute{Optional: true, PlanModifiers: replace, MarkdownDescription: "Member ID. Exactly one of `member_id` or `team_id` is required."},
		"team_id":            schema.StringAttribute{Optional: true, PlanModifiers: replace, MarkdownDescription: "Team ID. Exactly one of `member_id` or `team_id` is required."},
		"permission":         schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Permission: `usage` or `annotate`."},
		"connection_id":      schema.StringAttribute{Computed: true, MarkdownDescription: "Connection ID; only used by `sigma_connection_grant`."},
		"connection_path_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Connection path ID; only used by `sigma_connection_path_grant`."},
	}
	if r.path {
		resp.Schema.MarkdownDescription = "Manages a permission grant on a Sigma connection path."
		attrs["connection_path_id"] = schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Connection path ID (`inodeId` or `urlId`)."}
	} else {
		resp.Schema.MarkdownDescription = "Manages a permission grant on a Sigma connection."
		attrs["connection_id"] = schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Connection ID."}
	}
	resp.Schema.Attributes = attrs
}
func validateConnectionGrant(plan *connectionGrantModel) error {
	member, team := plan.MemberID.ValueString(), plan.TeamID.ValueString()
	if (member == "") == (team == "") {
		return fmt.Errorf("exactly one of member_id or team_id must be configured")
	}
	if plan.Permission.ValueString() != "usage" && plan.Permission.ValueString() != "annotate" {
		return fmt.Errorf("permission must be usage or annotate")
	}
	return nil
}
func findConnectionGrant(values []sigma.Grant, memberID, teamID, permission, grantID string) *sigma.Grant {
	for i := range values {
		value := &values[i]
		if grantID != "" && value.GrantID == grantID {
			return value
		}
		memberMatches := memberID != "" && value.MemberID != nil && *value.MemberID == memberID
		teamMatches := teamID != "" && value.TeamID != nil && *value.TeamID == teamID
		if (memberMatches || teamMatches) && value.Permission == permission {
			return value
		}
	}
	return nil
}
func (r *connectionGrantResource) list(ctx context.Context, state *connectionGrantModel) ([]sigma.Grant, error) {
	if r.path {
		return r.client.ListConnectionPathGrants(ctx, state.ConnectionPathID.ValueString())
	}
	return r.client.ListConnectionGrants(ctx, state.ConnectionID.ValueString())
}
func (r *connectionGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan connectionGrantModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := validateConnectionGrant(&plan); err != nil {
		resp.Diagnostics.AddError("Invalid Sigma grant configuration", err.Error())
		return
	}
	var err error
	if r.path {
		err = r.client.CreateConnectionPathGrant(ctx, plan.ConnectionPathID.ValueString(), plan.MemberID.ValueString(), plan.TeamID.ValueString(), plan.Permission.ValueString())
	} else {
		err = r.client.CreateConnectionGrant(ctx, plan.ConnectionID.ValueString(), plan.MemberID.ValueString(), plan.TeamID.ValueString(), plan.Permission.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma grant", err.Error())
		return
	}
	values, err := r.list(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Unable to locate created Sigma grant", err.Error())
		return
	}
	value := findConnectionGrant(values, plan.MemberID.ValueString(), plan.TeamID.ValueString(), plan.Permission.ValueString(), "")
	if value == nil {
		resp.Diagnostics.AddError("Unable to locate created Sigma grant", "Sigma accepted the grant but did not return it from the list endpoint.")
		return
	}
	plan.ID = types.StringValue(value.GrantID)
	if r.path {
		plan.ConnectionID = types.StringNull()
	} else {
		plan.ConnectionPathID = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *connectionGrantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state connectionGrantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	values, err := r.list(ctx, &state)
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma grant", err.Error())
		return
	}
	value := findConnectionGrant(values, "", "", "", state.ID.ValueString())
	if value == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	state.Permission = types.StringValue(value.Permission)
	if value.MemberID != nil {
		state.MemberID = types.StringValue(*value.MemberID)
		state.TeamID = types.StringNull()
	} else if value.TeamID != nil {
		state.TeamID = types.StringValue(*value.TeamID)
		state.MemberID = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *connectionGrantResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
}
func (r *connectionGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state connectionGrantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var err error
	if r.path {
		err = r.client.DeleteConnectionPathGrant(ctx, state.ConnectionPathID.ValueString(), state.ID.ValueString())
	} else {
		err = r.client.DeleteConnectionGrant(ctx, state.ConnectionID.ValueString(), state.ID.ValueString())
	}
	if err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma grant", err.Error())
	}
}
func (r *connectionGrantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Use `connectionId/grantId` or `connectionPathId/grantId`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
	if r.path {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("connection_path_id"), parts[0])...)
	} else {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("connection_id"), parts[0])...)
	}
}

type apiConnectorResource struct{ configuredResource }
type apiConnectorResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	ParamsJSON       types.String `tfsdk:"params_json"`
	ConfigJSON       types.String `tfsdk:"config_json"`
	AuthID           types.String `tfsdk:"auth_id"`
	SecretsWO        types.String `tfsdk:"secrets_wo"`
	SecretsWOVersion types.Int64  `tfsdk:"secrets_wo_version"`
}

func NewAPIConnectorResource() resource.Resource { return &apiConnectorResource{} }
func (r *apiConnectorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_connector"
}
func (r *apiConnectorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *apiConnectorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Manages a third-party API connector in Sigma. When `secrets_wo` has been managed via `secrets_wo_version`, changing `params_json` requires incrementing the version and resupplying secrets because PATCH replaces the full `params` object. Metadata-only updates omit `params` so existing secrets are retained.", Attributes: map[string]schema.Attribute{
		"id":                 schema.StringAttribute{Computed: true, MarkdownDescription: "API connector ID."},
		"name":               schema.StringAttribute{Required: true, MarkdownDescription: "Display name."},
		"description":        schema.StringAttribute{Optional: true, MarkdownDescription: "Description."},
		"params_json":        schema.StringAttribute{Required: true, MarkdownDescription: "JSON request parameters (`method`, `url`, headers, path/query parameters, and body)."},
		"config_json":        schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "JSON timeout, retry, redirect, and rate limit configuration."},
		"auth_id":            schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Associated `sigma_api_credential` ID."},
		"secrets_wo":         schema.StringAttribute{Optional: true, WriteOnly: true, Sensitive: true, MarkdownDescription: "Write-only JSON object merged into `params_json` for static secret parameters. Required whenever `secrets_wo_version` changes."},
		"secrets_wo_version": schema.Int64Attribute{Optional: true, MarkdownDescription: "Set on create when using `secrets_wo`, and increment when rotating secrets or changing `params_json` after secrets were managed."},
	}}
}
func apiConnectorInput(plan *apiConnectorResourceModel) (sigma.APIConnectorInput, error) {
	params, err := mergeJSON(plan.ParamsJSON, plan.SecretsWO)
	if err != nil {
		return sigma.APIConnectorInput{}, fmt.Errorf("decode params JSON: %w", err)
	}
	config, err := rawJSON(plan.ConfigJSON)
	if err != nil {
		return sigma.APIConnectorInput{}, fmt.Errorf("decode config JSON: %w", err)
	}
	input := sigma.APIConnectorInput{Name: plan.Name.ValueString(), Description: plan.Description.ValueString(), Params: params, Config: config}
	if !plan.AuthID.IsNull() && !plan.AuthID.IsUnknown() {
		value := plan.AuthID.ValueString()
		input.AuthID = &value
	}
	return input, nil
}

// apiConnectorUpdateInput builds a PATCH body. When omitParams is true, params are left
// unchanged on the server so previously stored static secrets are retained.
func apiConnectorUpdateInput(plan *apiConnectorResourceModel, omitParams bool) (sigma.APIConnectorInput, error) {
	input, err := apiConnectorInput(plan)
	if err != nil {
		return sigma.APIConnectorInput{}, err
	}
	if omitParams {
		input.Params = nil
	}
	return input, nil
}
func setAPIConnector(state *apiConnectorResourceModel, value *sigma.APIConnector) {
	state.ID = types.StringValue(value.APIConnectorID)
	state.Name = types.StringValue(value.Name)
	state.Description = types.StringValue(value.Description)
	state.ParamsJSON = jsonString(value.Params)
	state.ConfigJSON = jsonString(value.Config)
	if value.AuthID == nil {
		state.AuthID = types.StringNull()
	} else {
		state.AuthID = types.StringValue(*value.AuthID)
	}
}
func (r *apiConnectorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apiConnectorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var config apiConnectorResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.SecretsWO = config.SecretsWO
	input, err := apiConnectorInput(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Sigma API connector configuration", err.Error())
		return
	}
	value, err := r.client.CreateAPIConnector(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma API connector", err.Error())
		return
	}
	setAPIConnector(&plan, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *apiConnectorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state apiConnectorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.GetAPIConnector(ctx, state.ID.ValueString())
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma API connector", err.Error())
		return
	}
	setAPIConnector(&state, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *apiConnectorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan apiConnectorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state apiConnectorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	var config apiConnectorResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	managedSecrets := !state.SecretsWOVersion.IsNull() && !state.SecretsWOVersion.IsUnknown()
	resendingSecrets := !plan.SecretsWOVersion.Equal(state.SecretsWOVersion)
	omitParams := false
	if managedSecrets && !resendingSecrets {
		if !jsonValuesEqual(plan.ParamsJSON, state.ParamsJSON) {
			resp.Diagnostics.AddError(
				"Cannot update Sigma API connector params without resending secrets",
				"This connector manages static secrets via `secrets_wo`. Increment `secrets_wo_version` and supply `secrets_wo` whenever `params_json` changes so secret parameters are not cleared. Metadata-only updates may omit a version bump.",
			)
			return
		}
		// Leave params unchanged on the server so existing secrets are retained.
		omitParams = true
	}
	if resendingSecrets {
		plan.SecretsWO = config.SecretsWO
	}
	input, err := apiConnectorUpdateInput(&plan, omitParams)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Sigma API connector configuration", err.Error())
		return
	}
	value, err := r.client.UpdateAPIConnector(ctx, state.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Sigma API connector", err.Error())
		return
	}
	setAPIConnector(&plan, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *apiConnectorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apiConnectorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAPIConnector(ctx, state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma API connector", err.Error())
	}
}
func (r *apiConnectorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importPassthrough(ctx, req, resp)
}

type apiCredentialResource struct{ configuredResource }
type apiCredentialResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	AuthMethod          types.String `tfsdk:"auth_method"`
	Allowlist           types.Set    `tfsdk:"allowlist"`
	CredentialJSON      types.String `tfsdk:"credential_json"`
	CredentialWO        types.String `tfsdk:"credential_wo"`
	CredentialWOVersion types.Int64  `tfsdk:"credential_wo_version"`
}

func NewAPICredentialResource() resource.Resource { return &apiCredentialResource{} }
func (r *apiCredentialResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_credential"
}
func (r *apiCredentialResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *apiCredentialResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Manages credentials used to call third-party APIs from Sigma. This resource does not manage Sigma organization API client keys. Increment `credential_wo_version` to rotate secrets; updates without a version bump omit `credential` so Sigma retains existing secrets.", Attributes: map[string]schema.Attribute{
		"id":                    schema.StringAttribute{Computed: true, MarkdownDescription: "API credential ID."},
		"name":                  schema.StringAttribute{Required: true, MarkdownDescription: "Display name."},
		"description":           schema.StringAttribute{Optional: true, MarkdownDescription: "Description."},
		"auth_method":           schema.StringAttribute{Computed: true, MarkdownDescription: "Authentication method."},
		"allowlist":             schema.SetAttribute{Required: true, ElementType: types.StringType, MarkdownDescription: "Non-empty hostname glob allowlist."},
		"credential_json":       schema.StringAttribute{Computed: true, MarkdownDescription: "Nonsensitive credential projection returned by Sigma."},
		"credential_wo":         schema.StringAttribute{Optional: true, WriteOnly: true, Sensitive: true, MarkdownDescription: "Write-only credential JSON. Supported methods are `basic`, `bearer`, `apiKey`, `oAuthClientCredentials`, and `awsSigV4`. Required on create and whenever `credential_wo_version` changes."},
		"credential_wo_version": schema.Int64Attribute{Optional: true, MarkdownDescription: "Increment to rotate or resend credential secrets. Updates without a version bump omit the credential body."},
	}}
}
func apiCredentialInput(ctx context.Context, plan *apiCredentialResourceModel, requireCredential bool) (sigma.APICredentialInput, error) {
	var allowlist []string
	diagnostics := plan.Allowlist.ElementsAs(ctx, &allowlist, false)
	if diagnostics.HasError() {
		return sigma.APICredentialInput{}, fmt.Errorf("decode allowlist")
	}
	if len(allowlist) == 0 {
		return sigma.APICredentialInput{}, fmt.Errorf("allowlist must not be empty")
	}
	credential, err := rawJSON(plan.CredentialWO)
	if err != nil {
		return sigma.APICredentialInput{}, fmt.Errorf("decode credential_wo: %w", err)
	}
	if requireCredential && len(credential) == 0 {
		return sigma.APICredentialInput{}, fmt.Errorf("credential_wo is required")
	}
	return sigma.APICredentialInput{Name: plan.Name.ValueString(), Description: plan.Description.ValueString(), Allowlist: allowlist, Credential: credential}, nil
}
func setAPICredential(ctx context.Context, state *apiCredentialResourceModel, value *sigma.APICredential) {
	state.ID = types.StringValue(value.APICredentialID)
	state.Name = types.StringValue(value.Name)
	state.Description = types.StringValue(value.Description)
	state.AuthMethod = types.StringValue(value.AuthMethod)
	state.CredentialJSON = jsonString(value.Credential)
	state.Allowlist, _ = types.SetValueFrom(ctx, types.StringType, value.Allowlist)
}
func (r *apiCredentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apiCredentialResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var config apiCredentialResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.CredentialWO = config.CredentialWO
	input, err := apiCredentialInput(ctx, &plan, true)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Sigma API credential configuration", err.Error())
		return
	}
	value, err := r.client.CreateAPICredential(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma API credential", err.Error())
		return
	}
	setAPICredential(ctx, &plan, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *apiCredentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state apiCredentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.GetAPICredential(ctx, state.ID.ValueString())
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma API credential", err.Error())
		return
	}
	setAPICredential(ctx, &state, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *apiCredentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan apiCredentialResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state apiCredentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	var config apiCredentialResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resendingCredential := !plan.CredentialWOVersion.Equal(state.CredentialWOVersion)
	if resendingCredential {
		plan.CredentialWO = config.CredentialWO
	}
	input, err := apiCredentialInput(ctx, &plan, resendingCredential)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Sigma API credential configuration", err.Error())
		return
	}
	value, err := r.client.UpdateAPICredential(ctx, state.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Sigma API credential", err.Error())
		return
	}
	setAPICredential(ctx, &plan, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *apiCredentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apiCredentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAPICredential(ctx, state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma API credential", err.Error())
	}
}
func (r *apiCredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importPassthrough(ctx, req, resp)
}
