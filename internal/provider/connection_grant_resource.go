package provider

import (
	"context"
	"fmt"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*connectionGrantResource)(nil)
	_ resource.ResourceWithConfigure   = (*connectionGrantResource)(nil)
	_ resource.ResourceWithImportState = (*connectionGrantResource)(nil)
)

type connectionGrantResource struct{ configuredResource }

type connectionGrantModel struct {
	ID               types.String `tfsdk:"id"`
	ConnectionID     types.String `tfsdk:"connection_id"`
	ConnectionPathID types.String `tfsdk:"connection_path_id"`
	MemberID         types.String `tfsdk:"member_id"`
	TeamID           types.String `tfsdk:"team_id"`
	Permission       types.String `tfsdk:"permission"`
}

func NewConnectionGrantResource() resource.Resource { return &connectionGrantResource{} }

func (r *connectionGrantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection_grant"
}

func (r *connectionGrantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}

func (r *connectionGrantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema.MarkdownDescription = "Manages a permission grant on a Sigma connection."
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id":                 schema.StringAttribute{Computed: true, MarkdownDescription: "Grant ID."},
		"member_id":          schema.StringAttribute{Optional: true, PlanModifiers: replace, MarkdownDescription: "Member ID. Exactly one of `member_id` or `team_id` is required."},
		"team_id":            schema.StringAttribute{Optional: true, PlanModifiers: replace, MarkdownDescription: "Team ID. Exactly one of `member_id` or `team_id` is required."},
		"permission":         schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Permission: `usage` or `annotate`."},
		"connection_id":      schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Connection ID."},
		"connection_path_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Connection path ID; only used by `sigma_connection_path_grant`."},
	}
}

func validateConnectionGrant(plan *connectionGrantModel) error {
	member, team := plan.MemberID.ValueString(), plan.TeamID.ValueString()
	if (member == "") == (team == "") {
		return fmt.Errorf("exactly one of member_id or team_id must be configured")
	}
	if plan.Permission.ValueString() != "usage" && plan.Permission.ValueString() != "annotate" {
		return fmt.Errorf("permission must be usage or annotate")
	}
	return nil
}

func lookupConnectionGrant(values []sigma.Grant, memberID, teamID, permission, grantID string) (*sigma.Grant, error) {
	if grantID != "" {
		for i := range values {
			if values[i].GrantID == grantID {
				return &values[i], nil
			}
		}
		return nil, &sigma.APIError{StatusCode: 404, Message: "grant not found"}
	}
	var matches []sigma.Grant
	for i := range values {
		value := values[i]
		memberMatches := memberID != "" && value.MemberID != nil && *value.MemberID == memberID
		teamMatches := teamID != "" && value.TeamID != nil && *value.TeamID == teamID
		if (memberMatches || teamMatches) && value.Permission == permission {
			matches = append(matches, value)
		}
	}
	switch len(matches) {
	case 1:
		return &matches[0], nil
	case 0:
		return nil, &sigma.APIError{StatusCode: 404, Message: "grant not found"}
	default:
		return nil, fmt.Errorf("multiple grants matched grantee and permission; refusing to select the first match")
	}
}

func (r *connectionGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan connectionGrantModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := validateConnectionGrant(&plan); err != nil {
		resp.Diagnostics.AddError("Invalid Sigma grant configuration", err.Error())
		return
	}
	if err := r.client.CreateConnectionGrant(ctx, plan.ConnectionID.ValueString(), plan.MemberID.ValueString(), plan.TeamID.ValueString(), plan.Permission.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma grant", err.Error())
		return
	}
	values, err := r.client.ListConnectionGrants(ctx, plan.ConnectionID.ValueString())
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
	plan.ConnectionPathID = types.StringNull()
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *connectionGrantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state connectionGrantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	values, err := r.client.ListConnectionGrants(ctx, state.ConnectionID.ValueString())
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

func (r *connectionGrantResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
}

func (r *connectionGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state connectionGrantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteConnectionGrant(ctx, state.ConnectionID.ValueString(), state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma grant", err.Error())
	}
}

func (r *connectionGrantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parentID, grantID, ok := splitCompositeImportID(req.ID)
	if !ok {
		resp.Diagnostics.AddError("Invalid import ID", "Use `connectionId/grantId` or `connectionPathId/grantId` with non-empty segments.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), grantID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("connection_id"), parentID)...)
}
