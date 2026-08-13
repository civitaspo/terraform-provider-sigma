package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tagResource struct{ configuredResource }

var (
	_ resource.Resource                = (*tagResource)(nil)
	_ resource.ResourceWithConfigure   = (*tagResource)(nil)
	_ resource.ResourceWithImportState = (*tagResource)(nil)
)

type tagModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Color       types.String `tfsdk:"color"`
	Description types.String `tfsdk:"description"`
}

func NewTagResource() resource.Resource { return &tagResource{} }

func (r *tagResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}
func (r *tagResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *tagResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Sigma version tag. The update API only accepts `description`; changing `name` or `color` forces replacement.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Version tag ID (`versionTagId`)."},
			"name":        schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Unique tag name."},
			"color":       schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Tag color: `cyan`, `grass`, `violet`, `plum`, `amber`, or `bronze`."},
			"description": schema.StringAttribute{Optional: true, MarkdownDescription: "Tag description."},
		},
	}
}
func setTag(state *tagModel, value *sigma.Tag) {
	state.ID = types.StringValue(value.VersionTagID)
	state.Name = types.StringValue(value.Name)
	if value.Color != "" {
		state.Color = types.StringValue(value.Color)
	}
	if value.Description != nil {
		state.Description = types.StringValue(*value.Description)
	} else if state.Description.IsUnknown() {
		state.Description = types.StringNull()
	}
}
func (r *tagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tagModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.CreateTag(ctx, sigma.CreateTagInput{
		Name: plan.Name.ValueString(), Color: plan.Color.ValueString(), Description: plan.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma tag", err.Error())
		return
	}
	if refreshed, err := r.client.GetTag(ctx, value.VersionTagID); err == nil {
		value = refreshed
	} else {
		value.Color = plan.Color.ValueString()
		if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
			desc := plan.Description.ValueString()
			value.Description = &desc
		}
	}
	setTag(&plan, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *tagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tagModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.GetTag(ctx, state.ID.ValueString())
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma tag", err.Error())
		return
	}
	setTag(&state, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *tagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tagModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.UpdateTag(ctx, plan.ID.ValueString(), sigma.UpdateTagInput{Description: plan.Description.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Sigma tag", err.Error())
		return
	}
	setTag(&plan, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *tagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tagModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteTag(ctx, state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma tag", err.Error())
	}
}
func (r *tagResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importPassthrough(ctx, req, resp)
}
