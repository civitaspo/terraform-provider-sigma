package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*connectionPathGrantResource)(nil)
	_ resource.ResourceWithConfigure   = (*connectionPathGrantResource)(nil)
	_ resource.ResourceWithImportState = (*connectionPathGrantResource)(nil)
)

type connectionPathGrantResource struct{ configuredResource }

func NewConnectionPathGrantResource() resource.Resource {
	return &connectionPathGrantResource{}
}

func (r *connectionPathGrantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection_path_grant"
}

func (r *connectionPathGrantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}

func (r *connectionPathGrantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema.MarkdownDescription = "Manages a permission grant on a Sigma connection path."
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id":                 schema.StringAttribute{Computed: true, MarkdownDescription: "Grant ID."},
		"member_id":          schema.StringAttribute{Optional: true, PlanModifiers: replace, MarkdownDescription: "Member ID. Exactly one of `member_id` or `team_id` is required."},
		"team_id":            schema.StringAttribute{Optional: true, PlanModifiers: replace, MarkdownDescription: "Team ID. Exactly one of `member_id` or `team_id` is required."},
		"permission":         schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Permission: `usage` or `annotate`."},
		"connection_id":      schema.StringAttribute{Computed: true, MarkdownDescription: "Connection ID; only used by `sigma_connection_grant`."},
		"connection_path_id": schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Connection path ID (`inodeId` or `urlId`)."},
	}
}

func (r *connectionPathGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan connectionGrantModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := validateConnectionGrant(&plan); err != nil {
		resp.Diagnostics.AddError("Invalid Sigma grant configuration", err.Error())
		return
	}
	if err := r.client.CreateConnectionPathGrant(ctx, plan.ConnectionPathID.ValueString(), plan.MemberID.ValueString(), plan.TeamID.ValueString(), plan.Permission.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma grant", err.Error())
		return
	}
	values, err := r.client.ListConnectionPathGrants(ctx, plan.ConnectionPathID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to locate created Sigma grant", err.Error())
		return
	}
	value, err := lookupConnectionGrant(values, plan.MemberID.ValueString(), plan.TeamID.ValueString(), plan.Permission.ValueString(), "")
	if err != nil {
		resp.Diagnostics.AddError("Unable to locate created Sigma grant", err.Error())
		return
	}
	plan.ID = types.StringValue(value.GrantID)
	plan.ConnectionID = types.StringNull()
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *connectionPathGrantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state connectionGrantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	values, err := r.client.ListConnectionPathGrants(ctx, state.ConnectionPathID.ValueString())
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma grant", err.Error())
		return
	}
	value, err := lookupConnectionGrant(values, "", "", "", state.ID.ValueString())
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma grant", err.Error())
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

func (r *connectionPathGrantResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
}

func (r *connectionPathGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state connectionGrantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteConnectionPathGrant(ctx, state.ConnectionPathID.ValueString(), state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma grant", err.Error())
	}
}

func (r *connectionPathGrantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parentID, grantID, ok := splitCompositeImportID(req.ID)
	if !ok {
		resp.Diagnostics.AddError("Invalid import ID", "Use `connectionId/grantId` or `connectionPathId/grantId` with non-empty segments.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), grantID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("connection_path_id"), parentID)...)
}
