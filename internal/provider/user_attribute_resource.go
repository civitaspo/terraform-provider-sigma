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

type userAttributeResource struct{ configuredResource }

var (
	_ resource.Resource                = (*userAttributeResource)(nil)
	_ resource.ResourceWithConfigure   = (*userAttributeResource)(nil)
	_ resource.ResourceWithImportState = (*userAttributeResource)(nil)
)

type userAttributeModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	DefaultValue types.String `tfsdk:"default_value"`
}

func NewUserAttributeResource() resource.Resource { return &userAttributeResource{} }
func (r *userAttributeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_attribute"
}
func (r *userAttributeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *userAttributeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Sigma user attribute. The API has no update endpoint, so configuration changes replace it.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true, MarkdownDescription: "User attribute ID."},
			"name":          schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "User attribute name."},
			"description":   schema.StringAttribute{Optional: true, PlanModifiers: replace, MarkdownDescription: "User attribute description."},
			"default_value": schema.StringAttribute{Optional: true, PlanModifiers: replace, MarkdownDescription: "Default string value."},
		},
	}
}
func (r *userAttributeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userAttributeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var defaultValue *sigma.AttributeValue
	if !plan.DefaultValue.IsNull() {
		defaultValue = &sigma.AttributeValue{Val: plan.DefaultValue.ValueString(), Type: "string"}
	}
	value, err := r.client.CreateUserAttribute(ctx, plan.Name.ValueString(), plan.Description.ValueString(), defaultValue)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma user attribute", err.Error())
		return
	}
	setUserAttribute(&plan, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *userAttributeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userAttributeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	value, err := r.client.GetUserAttribute(ctx, state.ID.ValueString())
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma user attribute", err.Error())
		return
	}
	setUserAttribute(&state, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *userAttributeResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
}
func (r *userAttributeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userAttributeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if !resp.Diagnostics.HasError() {
		if err := r.client.DeleteUserAttribute(ctx, state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
			resp.Diagnostics.AddError("Unable to delete Sigma user attribute", err.Error())
		}
	}
}
func (r *userAttributeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importPassthrough(ctx, req, resp)
}
func setUserAttribute(state *userAttributeModel, value *sigma.UserAttribute) {
	state.ID = types.StringValue(value.UserAttributeID)
	state.Name = types.StringValue(value.Name)
	if value.Description == nil {
		state.Description = types.StringNull()
	} else {
		state.Description = types.StringValue(*value.Description)
	}
	if value.DefaultValue == nil {
		state.DefaultValue = types.StringNull()
	} else {
		state.DefaultValue = types.StringValue(value.DefaultValue.Val)
	}
}
