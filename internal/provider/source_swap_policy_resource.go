package provider

import (
	"context"
	"fmt"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type sourceSwapPolicyResource struct{ configuredResource }

var (
	_ resource.Resource                = (*sourceSwapPolicyResource)(nil)
	_ resource.ResourceWithConfigure   = (*sourceSwapPolicyResource)(nil)
	_ resource.ResourceWithImportState = (*sourceSwapPolicyResource)(nil)
)

type sourceSwapPolicyModel struct {
	ID               types.String `tfsdk:"id"`
	Type             types.String `tfsdk:"type"`
	Name             types.String `tfsdk:"name"`
	FromConnectionID types.String `tfsdk:"from_connection_id"`
	SwapsJSON        types.String `tfsdk:"swaps_json"`
}

func NewSourceSwapPolicyResource() resource.Resource { return &sourceSwapPolicyResource{} }

func (r *sourceSwapPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_source_swap_policy"
}
func (r *sourceSwapPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *sourceSwapPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Sigma source swap policy. `swaps_json` is polymorphic by `type` (`connection` or `deployment`). " + betaAPINotice,
		Attributes: map[string]schema.Attribute{
			"id":                 schema.StringAttribute{Computed: true, MarkdownDescription: "Source swap policy ID."},
			"type":               schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Policy type: `connection` or `deployment`."},
			"name":               schema.StringAttribute{Required: true, MarkdownDescription: "Source swap policy name."},
			"from_connection_id": schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Connection ID to swap from."},
			"swaps_json":         schema.StringAttribute{Required: true, MarkdownDescription: "JSON swaps object accepted by the Sigma source swap policy API."},
		},
	}
}
func sourceSwapInput(plan *sourceSwapPolicyModel) (sigma.SourceSwapPolicyInput, error) {
	swaps, err := rawJSON(plan.SwapsJSON)
	if err != nil {
		return sigma.SourceSwapPolicyInput{}, fmt.Errorf("decode swaps_json: %w", err)
	}
	if len(swaps) == 0 {
		return sigma.SourceSwapPolicyInput{}, fmt.Errorf("swaps_json is required")
	}
	return sigma.SourceSwapPolicyInput{
		Type:             plan.Type.ValueString(),
		Name:             plan.Name.ValueString(),
		FromConnectionID: plan.FromConnectionID.ValueString(),
		Swaps:            swaps,
	}, nil
}
func setSourceSwapPolicy(state *sourceSwapPolicyModel, value *sigma.SourceSwapPolicy) {
	state.ID = types.StringValue(value.PolicyID)
	state.Type = types.StringValue(value.Type)
	state.Name = types.StringValue(value.Name)
	state.FromConnectionID = types.StringValue(value.FromConnectionID)
	if len(value.Swaps) > 0 && string(value.Swaps) != "null" {
		state.SwapsJSON = jsonString(value.Swaps)
	}
}
func (r *sourceSwapPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sourceSwapPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in, err := sourceSwapInput(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Sigma source swap policy configuration", err.Error())
		return
	}
	value, err := r.client.CreateSourceSwapPolicy(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma source swap policy", err.Error())
		return
	}
	setSourceSwapPolicy(&plan, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *sourceSwapPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sourceSwapPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.GetSourceSwapPolicy(ctx, state.ID.ValueString())
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma source swap policy", err.Error())
		return
	}
	setSourceSwapPolicy(&state, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *sourceSwapPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan sourceSwapPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in, err := sourceSwapInput(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Sigma source swap policy configuration", err.Error())
		return
	}
	value, err := r.client.UpdateSourceSwapPolicy(ctx, plan.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Sigma source swap policy", err.Error())
		return
	}
	setSourceSwapPolicy(&plan, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *sourceSwapPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sourceSwapPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteSourceSwapPolicy(ctx, state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma source swap policy", err.Error())
	}
}
func (r *sourceSwapPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importPassthrough(ctx, req, resp)
}
