package provider

import (
	"context"
	"fmt"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type apiCredentialResource struct{ configuredResource }

var (
	_ resource.Resource                = (*apiCredentialResource)(nil)
	_ resource.ResourceWithConfigure   = (*apiCredentialResource)(nil)
	_ resource.ResourceWithImportState = (*apiCredentialResource)(nil)
)

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
