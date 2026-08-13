package provider

import (
	"context"
	"fmt"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type connectionResource struct{ configuredResource }

var (
	_ resource.Resource                = (*connectionResource)(nil)
	_ resource.ResourceWithConfigure   = (*connectionResource)(nil)
	_ resource.ResourceWithImportState = (*connectionResource)(nil)
)

type connectionResourceModel struct {
	ID                   types.String  `tfsdk:"id"`
	Name                 types.String  `tfsdk:"name"`
	Type                 types.String  `tfsdk:"type"`
	DetailsJSON          types.String  `tfsdk:"details_json"`
	DescriptionJSON      types.String  `tfsdk:"description_json"`
	PoolSizesJSON        types.String  `tfsdk:"pool_sizes_json"`
	TimeoutSecs          types.Float64 `tfsdk:"timeout_secs"`
	UseFriendlyNames     types.Bool    `tfsdk:"use_friendly_names"`
	UseOauth             types.Bool    `tfsdk:"use_oauth"`
	Restore              types.Bool    `tfsdk:"restore"`
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
		MarkdownDescription: "Manages a Sigma warehouse connection. `details_json` is polymorphic by warehouse `type`; put write-only fields such as `password`, `serviceAccount`, and `clientSecret` in `credentials_wo`. Sigma's update connection API replaces warehouse details entirely, so any update that previously sent `credentials_wo` requires incrementing `credentials_wo_version` (and resupplying `credentials_wo`) to avoid clearing authentication. Restore is not a credentials-free path. `use_oauth` is computed from GET `useOauth`; warehouse OAuth settings remain in `details_json`. Sigma's get endpoint does not return warehouse details, so imported resources cannot recover them.",
		Attributes: map[string]schema.Attribute{
			"id":                     schema.StringAttribute{Computed: true, MarkdownDescription: "Connection ID."},
			"name":                   schema.StringAttribute{Required: true, MarkdownDescription: "Connection name."},
			"type":                   schema.StringAttribute{Computed: true, MarkdownDescription: "Warehouse type returned by Sigma."},
			"details_json":           schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "JSON object containing non-secret warehouse-specific connection details."},
			"description_json":       schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "JSON object accepted by Sigma as the connection description."},
			"pool_sizes_json":        schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "JSON object configuring connection pool sizes."},
			"timeout_secs":           schema.Float64Attribute{Optional: true, Computed: true, MarkdownDescription: "Connection timeout in seconds."},
			"use_friendly_names":     schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Whether friendly names are enabled."},
			"use_oauth":              schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the connection uses OAuth, as returned by Sigma GET `useOauth`. Not settable; warehouse OAuth configuration remains in `details_json`."},
			"restore":                schema.BoolAttribute{Optional: true, MarkdownDescription: "When true, PUT `restore=true` to unarchive a deleted connection. Sigma PUT still replaces warehouse details, so restore is not a credentials-free path: increment `credentials_wo_version` and resupply `credentials_wo` whenever credentials were previously managed."},
			"credentials_wo":         schema.StringAttribute{Optional: true, WriteOnly: true, Sensitive: true, MarkdownDescription: "Write-only JSON object merged into `details_json` before create or update. Required whenever `credentials_wo_version` changes."},
			"credentials_wo_version": schema.Int64Attribute{Optional: true, MarkdownDescription: "Set on create when using `credentials_wo`, and increment on every update that should retain or rotate warehouse credentials. Sigma PUT replaces details, so updates without a version bump are rejected when credentials were previously managed. Restore is not a credentials-free path."},
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
	if value.UseOauth != nil {
		state.UseOauth = types.BoolValue(*value.UseOauth)
	} else {
		state.UseOauth = types.BoolNull()
	}
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
	if !plan.Restore.IsNull() && !plan.Restore.IsUnknown() {
		restore := plan.Restore.ValueBool()
		input.Restore = &restore
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
