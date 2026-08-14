package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type apiCredentialResource struct{ configuredResource }

var (
	_ resource.Resource                = (*apiCredentialResource)(nil)
	_ resource.ResourceWithConfigure   = (*apiCredentialResource)(nil)
	_ resource.ResourceWithImportState = (*apiCredentialResource)(nil)
	_ resource.ResourceWithModifyPlan  = (*apiCredentialResource)(nil)
)

type apiCredentialResourceModel struct {
	ID                  types.String         `tfsdk:"id"`
	Name                types.String         `tfsdk:"name"`
	Description         types.String         `tfsdk:"description"`
	AuthMethod          types.String         `tfsdk:"auth_method"`
	Allowlist           types.Set            `tfsdk:"allowlist"`
	CredentialJSON      jsontypes.Normalized `tfsdk:"credential_json"`
	CredentialWO        types.String         `tfsdk:"credential_wo"`
	CredentialWOVersion types.Int64          `tfsdk:"credential_wo_version"`
}

func NewAPICredentialResource() resource.Resource { return &apiCredentialResource{} }

func (r *apiCredentialResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_credential"
}

func (r *apiCredentialResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}

func (r *apiCredentialResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Manages credentials used to call third-party APIs from Sigma. This resource does not manage Sigma organization API client keys. `credential_wo` and `credential_wo_version` must be supplied together. After credentials were managed, rotating secrets requires a strictly greater version and a known `credential_wo`. Updates without a version bump omit `credential` so Sigma retains existing secrets.", Attributes: map[string]schema.Attribute{
		"id":                    schema.StringAttribute{Computed: true, MarkdownDescription: "API credential ID."},
		"name":                  schema.StringAttribute{Required: true, MarkdownDescription: "Display name."},
		"description":           schema.StringAttribute{Optional: true, MarkdownDescription: "Description."},
		"auth_method":           schema.StringAttribute{Computed: true, MarkdownDescription: "Authentication method."},
		"allowlist":             schema.SetAttribute{Required: true, ElementType: types.StringType, MarkdownDescription: "Non-empty hostname glob allowlist."},
		"credential_json":       schema.StringAttribute{Computed: true, CustomType: jsontypes.NormalizedType{}, MarkdownDescription: "Nonsensitive credential projection returned by Sigma."},
		"credential_wo":         schema.StringAttribute{Optional: true, WriteOnly: true, Sensitive: true, MarkdownDescription: "Write-only credential JSON. Supported methods are `basic`, `bearer`, `apiKey`, `oAuthClientCredentials`, and `awsSigV4`. Required together with `credential_wo_version` on create and rotation."},
		"credential_wo_version": schema.Int64Attribute{Optional: true, MarkdownDescription: "Must be supplied with `credential_wo`. Increment to rotate secrets. Updates without a version bump omit the credential body."},
	}}
}

func (r *apiCredentialResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	var plan apiCredentialResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var config apiCredentialResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(writeOnlyVersionPair(config.CredentialWO, plan.CredentialWOVersion, "credential_wo", "credential_wo_version")...)
	if req.State.Raw.IsNull() {
		return
	}
	var state apiCredentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.CredentialWOVersion.IsNull() || state.CredentialWOVersion.IsUnknown() {
		return
	}
	if plan.CredentialWOVersion.IsNull() {
		resp.Diagnostics.AddError("Cannot remove credential_wo_version", "After credential_wo has been managed, credential_wo_version cannot be removed.")
		return
	}
	if !plan.CredentialWOVersion.Equal(state.CredentialWOVersion) {
		resp.Diagnostics.Append(managedWriteOnlyUpdate(state.CredentialWOVersion, plan.CredentialWOVersion, config.CredentialWO, "credential_wo", "credential_wo_version")...)
	}
}

func apiCredentialInput(ctx context.Context, plan *apiCredentialResourceModel, credential types.String, requireCredential bool) (sigma.APICredentialInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	allowlist, allowDiags := knownStringSet(ctx, plan.Allowlist, "allowlist")
	diags.Append(allowDiags...)
	if diags.HasError() {
		return sigma.APICredentialInput{}, diags
	}
	if len(allowlist) == 0 {
		diags.AddError("Invalid allowlist", "allowlist must not be empty")
		return sigma.APICredentialInput{}, diags
	}
	var body []byte
	if requireCredential {
		if credential.IsNull() || credential.IsUnknown() {
			diags.AddError("Invalid credential_wo", "credential_wo is required")
			return sigma.APICredentialInput{}, diags
		}
		canonical, err := canonicalJSON([]byte(credential.ValueString()))
		if err != nil {
			diags.AddError("Invalid credential_wo", err.Error())
			return sigma.APICredentialInput{}, diags
		}
		body = []byte(canonical)
	}
	return sigma.APICredentialInput{Name: plan.Name.ValueString(), Description: plan.Description.ValueString(), Allowlist: allowlist, Credential: body}, diags
}

func setAPICredential(ctx context.Context, state *apiCredentialResourceModel, value *sigma.APICredential) diag.Diagnostics {
	var diags diag.Diagnostics
	priorVersion := state.CredentialWOVersion
	state.ID = types.StringValue(value.APICredentialID)
	state.Name = types.StringValue(value.Name)
	state.Description = types.StringValue(value.Description)
	state.AuthMethod = types.StringValue(value.AuthMethod)
	state.CredentialJSON = normalizedFromRaw(value.Credential)
	state.CredentialWOVersion = priorVersion
	allowlist, allowDiags := stringSetValue(ctx, value.Allowlist)
	diags.Append(allowDiags...)
	state.Allowlist = allowlist
	return diags
}

func (r *apiCredentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apiCredentialResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var config apiCredentialResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(writeOnlyVersionPair(config.CredentialWO, plan.CredentialWOVersion, "credential_wo", "credential_wo_version")...)
	if resp.Diagnostics.HasError() {
		return
	}
	input, diags := apiCredentialInput(ctx, &plan, config.CredentialWO, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.CreateAPICredential(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma API credential", err.Error())
		return
	}
	resp.Diagnostics.Append(setAPICredential(ctx, &plan, value)...)
	if resp.Diagnostics.HasError() {
		return
	}
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
	resp.Diagnostics.Append(setAPICredential(ctx, &state, value)...)
	if resp.Diagnostics.HasError() {
		return
	}
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
	if !state.CredentialWOVersion.IsNull() && plan.CredentialWOVersion.IsNull() {
		resp.Diagnostics.AddError("Cannot remove credential_wo_version", "After credential_wo has been managed, credential_wo_version cannot be removed.")
		return
	}
	if resendingCredential {
		resp.Diagnostics.Append(managedWriteOnlyUpdate(state.CredentialWOVersion, plan.CredentialWOVersion, config.CredentialWO, "credential_wo", "credential_wo_version")...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	input, diags := apiCredentialInput(ctx, &plan, config.CredentialWO, resendingCredential)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.UpdateAPICredential(ctx, state.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Sigma API credential", err.Error())
		return
	}
	resp.Diagnostics.Append(setAPICredential(ctx, &plan, value)...)
	if resp.Diagnostics.HasError() {
		return
	}
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
