package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type accountTypeResource struct{ configuredResource }

var (
	_ resource.Resource                = (*accountTypeResource)(nil)
	_ resource.ResourceWithConfigure   = (*accountTypeResource)(nil)
	_ resource.ResourceWithImportState = (*accountTypeResource)(nil)
)

type accountTypeModel struct {
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	Description             types.String `tfsdk:"description"`
	Permissions             types.Set    `tfsdk:"permissions"`
	IsCustom                types.Bool   `tfsdk:"is_custom"`
	ReassignToAccountTypeID types.String `tfsdk:"reassign_to_account_type_id"`
}

func NewAccountTypeResource() resource.Resource { return &accountTypeResource{} }
func (r *accountTypeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account_type"
}
func (r *accountTypeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *accountTypeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a custom Sigma account type. The API has no update or get-by-ID endpoint, so name, description, and permissions changes replace it. Import by account type name. `reassign_to_account_type_id` is destroy-time only and can change without recreation.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Account type ID."},
			"name":        schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Account type name."},
			"description": schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Account type description."},
			"permissions": schema.SetAttribute{Required: true, ElementType: types.StringType, PlanModifiers: []planmodifier.Set{setplanmodifier.RequiresReplace()}, MarkdownDescription: "Enabled permission names."},
			"is_custom":   schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this is a custom account type."},
			"reassign_to_account_type_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Account type ID that receives users when this type is destroyed (`reassignToAccountTypeId` query parameter). Used only on delete; changing it updates state without recreating the account type. Live Sigma requires a UUID even when no users are assigned.",
			},
		},
	}
}
func (r *accountTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan accountTypeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	name, nameDiags := knownString(plan.Name, "name")
	resp.Diagnostics.Append(nameDiags...)
	description, descriptionDiags := knownString(plan.Description, "description")
	resp.Diagnostics.Append(descriptionDiags...)
	permissions, permissionDiags := knownStringSet(ctx, plan.Permissions, "permissions")
	resp.Diagnostics.Append(permissionDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.CreateAccountType(ctx, name, description, permissions)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma account type", err.Error())
		return
	}
	setAccountType(&plan, value)
	converted, convertDiags := stringSetValue(ctx, permissions)
	resp.Diagnostics.Append(convertDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Permissions = converted
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *accountTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state accountTypeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	lookup := state.ID.ValueString()
	if name := optionalStringPtr(state.Name); name != nil && *name != "" {
		lookup = *name
	}
	if lookup == "" {
		resp.Diagnostics.AddError("Unable to read Sigma account type", "The account type ID or name is unknown.")
		return
	}
	value, err := r.client.FindAccountType(ctx, lookup)
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma account type", err.Error())
		return
	}
	setAccountType(&state, value)
	permissions, err := r.client.ListAccountTypePermissions(ctx, value.AccountTypeID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma account type permissions", err.Error())
		return
	}
	names := make([]string, 0, len(permissions))
	for i := range permissions {
		names = append(names, permissions[i].Permission)
	}
	converted, convertDiags := stringSetValue(ctx, names)
	resp.Diagnostics.Append(convertDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Permissions = converted
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *accountTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state accountTypeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Only destroy-time reassign_to_account_type_id is mutable in place; other fields ForceNew.
	state.ReassignToAccountTypeID = plan.ReassignToAccountTypeID
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *accountTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state accountTypeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, diags := knownString(state.ID, "id")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	reassign := ""
	if value := optionalStringPtr(state.ReassignToAccountTypeID); value != nil {
		reassign = *value
	}
	if err := r.client.DeleteAccountType(ctx, id, reassign); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma account type", err.Error())
	}
}
func (r *accountTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}
func setAccountType(state *accountTypeModel, value *sigma.AccountType) {
	state.ID = types.StringValue(value.AccountTypeID)
	state.Name = types.StringValue(value.AccountTypeName)
	state.Description = types.StringValue(value.Description)
	state.IsCustom = types.BoolValue(value.IsCustom)
}
