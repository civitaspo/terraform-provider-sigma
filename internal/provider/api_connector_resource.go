package provider

import (
	"context"
	"fmt"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type apiConnectorResource struct{ configuredResource }

var (
	_ resource.Resource                = (*apiConnectorResource)(nil)
	_ resource.ResourceWithConfigure   = (*apiConnectorResource)(nil)
	_ resource.ResourceWithImportState = (*apiConnectorResource)(nil)
)

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
