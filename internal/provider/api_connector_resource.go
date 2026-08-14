package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type apiConnectorResource struct{ configuredResource }

var (
	_ resource.Resource                = (*apiConnectorResource)(nil)
	_ resource.ResourceWithConfigure   = (*apiConnectorResource)(nil)
	_ resource.ResourceWithImportState = (*apiConnectorResource)(nil)
	_ resource.ResourceWithModifyPlan  = (*apiConnectorResource)(nil)
)

type apiConnectorResourceModel struct {
	ID               types.String         `tfsdk:"id"`
	Name             types.String         `tfsdk:"name"`
	Description      types.String         `tfsdk:"description"`
	ParamsJSON       jsontypes.Normalized `tfsdk:"params_json"`
	ConfigJSON       jsontypes.Normalized `tfsdk:"config_json"`
	AuthID           types.String         `tfsdk:"auth_id"`
	SecretsWO        types.String         `tfsdk:"secrets_wo"`
	SecretsWOVersion types.Int64          `tfsdk:"secrets_wo_version"`
}

func NewAPIConnectorResource() resource.Resource { return &apiConnectorResource{} }

func (r *apiConnectorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_connector"
}

func (r *apiConnectorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}

func (r *apiConnectorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Manages a third-party API connector in Sigma. `secrets_wo` and `secrets_wo_version` must be supplied together. After secrets were managed, sending `params` requires a strictly greater version and a known `secrets_wo`. Metadata-only updates omit `params` so existing secrets are retained. Read does not replace configured params with a redacted GET body.", Attributes: map[string]schema.Attribute{
		"id":                 schema.StringAttribute{Computed: true, MarkdownDescription: "API connector ID."},
		"name":               schema.StringAttribute{Required: true, MarkdownDescription: "Display name."},
		"description":        schema.StringAttribute{Optional: true, MarkdownDescription: "Description."},
		"params_json":        schema.StringAttribute{Required: true, CustomType: jsontypes.NormalizedType{}, MarkdownDescription: "JSON request parameters (`method`, `url`, headers, path/query parameters, and body)."},
		"config_json":        schema.StringAttribute{Optional: true, Computed: true, CustomType: jsontypes.NormalizedType{}, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "JSON timeout, retry, redirect, and rate limit configuration."},
		"auth_id":            schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Associated `sigma_api_credential` ID."},
		"secrets_wo":         schema.StringAttribute{Optional: true, WriteOnly: true, Sensitive: true, MarkdownDescription: "Write-only JSON object merged into `params_json` for static secret parameters. Required together with `secrets_wo_version`."},
		"secrets_wo_version": schema.Int64Attribute{Optional: true, MarkdownDescription: "Must be supplied with `secrets_wo`. Increment when rotating secrets or changing `params_json` after secrets were managed."},
	}}
}

func (r *apiConnectorResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	var plan apiConnectorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var config apiConnectorResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(writeOnlyVersionPair(config.SecretsWO, plan.SecretsWOVersion, "secrets_wo", "secrets_wo_version")...)
	if req.State.Raw.IsNull() {
		return
	}
	var state apiConnectorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.SecretsWOVersion.IsNull() || state.SecretsWOVersion.IsUnknown() {
		return
	}
	if plan.SecretsWOVersion.IsNull() {
		resp.Diagnostics.AddError("Cannot remove secrets_wo_version", "After secrets_wo has been managed, secrets_wo_version cannot be removed.")
		return
	}
	paramsChanged := !plan.ParamsJSON.Equal(state.ParamsJSON)
	resending := !plan.SecretsWOVersion.Equal(state.SecretsWOVersion)
	if paramsChanged && !resending {
		resp.Diagnostics.AddError(
			"Cannot update Sigma API connector params without resending secrets",
			"Increment `secrets_wo_version` and supply `secrets_wo` whenever `params_json` changes so secret parameters are not cleared.",
		)
		return
	}
	if resending {
		resp.Diagnostics.Append(managedWriteOnlyUpdate(state.SecretsWOVersion, plan.SecretsWOVersion, config.SecretsWO, "secrets_wo", "secrets_wo_version")...)
	}
}

func apiConnectorInput(plan *apiConnectorResourceModel, secrets types.String) (sigma.APIConnectorInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	paramsObject, _, paramsDiags := knownNormalizedObject(plan.ParamsJSON, "params_json")
	diags.Append(paramsDiags...)
	if diags.HasError() {
		return sigma.APIConnectorInput{}, diags
	}
	params, mergeDiags := mergeObjectWithWriteOnly(paramsObject, secrets, "secrets_wo")
	diags.Append(mergeDiags...)
	config, configDiags := optionalNormalizedJSON(plan.ConfigJSON, "config_json")
	diags.Append(configDiags...)
	if diags.HasError() {
		return sigma.APIConnectorInput{}, diags
	}
	input := sigma.APIConnectorInput{Name: plan.Name.ValueString(), Description: plan.Description.ValueString(), Params: params, Config: config}
	if !plan.AuthID.IsNull() && !plan.AuthID.IsUnknown() {
		value := plan.AuthID.ValueString()
		input.AuthID = &value
	}
	return input, diags
}

func setAPIConnector(state *apiConnectorResourceModel, value *sigma.APIConnector, preserveParams bool) {
	priorParams := state.ParamsJSON
	priorVersion := state.SecretsWOVersion
	state.ID = types.StringValue(value.APIConnectorID)
	state.Name = types.StringValue(value.Name)
	state.Description = types.StringValue(value.Description)
	if preserveParams {
		state.ParamsJSON = priorParams
	} else {
		state.ParamsJSON = normalizedFromRaw(value.Params)
	}
	state.ConfigJSON = normalizedFromRaw(value.Config)
	state.SecretsWOVersion = priorVersion
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
	resp.Diagnostics.Append(writeOnlyVersionPair(config.SecretsWO, plan.SecretsWOVersion, "secrets_wo", "secrets_wo_version")...)
	if resp.Diagnostics.HasError() {
		return
	}
	input, diags := apiConnectorInput(&plan, config.SecretsWO)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.CreateAPIConnector(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma API connector", err.Error())
		return
	}
	setAPIConnector(&plan, value, true)
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
	preserveParams := !state.SecretsWOVersion.IsNull() && !state.SecretsWOVersion.IsUnknown()
	setAPIConnector(&state, value, preserveParams)
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
		if !plan.ParamsJSON.Equal(state.ParamsJSON) {
			resp.Diagnostics.AddError(
				"Cannot update Sigma API connector params without resending secrets",
				"This connector manages static secrets via `secrets_wo`. Increment `secrets_wo_version` and supply `secrets_wo` whenever `params_json` changes so secret parameters are not cleared. Metadata-only updates may omit a version bump.",
			)
			return
		}
		omitParams = true
	}
	if resendingSecrets {
		resp.Diagnostics.Append(managedWriteOnlyUpdate(state.SecretsWOVersion, plan.SecretsWOVersion, config.SecretsWO, "secrets_wo", "secrets_wo_version")...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	input, diags := apiConnectorInput(&plan, config.SecretsWO)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if omitParams {
		input.Params = nil
	}
	value, err := r.client.UpdateAPIConnector(ctx, state.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Sigma API connector", err.Error())
		return
	}
	setAPIConnector(&plan, value, true)
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
