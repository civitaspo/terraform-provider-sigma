package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = (*grantResource)(nil)
	_ resource.ResourceWithConfigure      = (*grantResource)(nil)
	_ resource.ResourceWithImportState    = (*grantResource)(nil)
	_ resource.ResourceWithValidateConfig = (*grantResource)(nil)
)

type grantResource struct{ configuredResource }

type grantModel struct {
	ID             types.String `tfsdk:"id"`
	InodeID        types.String `tfsdk:"inode_id"`
	MemberID       types.String `tfsdk:"member_id"`
	TeamID         types.String `tfsdk:"team_id"`
	Permission     types.String `tfsdk:"permission"`
	TagID          types.String `tfsdk:"tag_id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	InodeType      types.String `tfsdk:"inode_type"`
	CreatedBy      types.String `tfsdk:"created_by"`
	UpdatedBy      types.String `tfsdk:"updated_by"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

func NewGrantResource() resource.Resource { return &grantResource{} }

func (r *grantResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_grant"
}

func (r *grantResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.configure(request, response)
}

func (r *grantResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = grantSchema(
		"Manages a generic Sigma inode grant.",
		"Inode ID.",
		"Permission allowed by the inode type: `admin`, `annotate`, `update`, `usage`, `writeback`, `create`, `organize`, `explore`, `view`, `edit`, or `apply`.",
	)
}

func grantSchema(description, inodeDescription, permissionDescription string) schema.Schema {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	return schema.Schema{
		MarkdownDescription: description,
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Computed: true, MarkdownDescription: "Grant ID."},
			"inode_id":        schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: inodeDescription},
			"member_id":       schema.StringAttribute{Optional: true, PlanModifiers: replace, MarkdownDescription: "Member ID receiving the grant. Exactly one of `member_id` or `team_id` must be set."},
			"team_id":         schema.StringAttribute{Optional: true, PlanModifiers: replace, MarkdownDescription: "Team ID receiving the grant. Exactly one of `member_id` or `team_id` must be set."},
			"permission":      schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: permissionDescription},
			"tag_id":          schema.StringAttribute{Optional: true, PlanModifiers: replace, MarkdownDescription: "Optional version tag ID. Supported by generic, workbook, and report grants. Changing this forces a new resource. Tagged workbook/report grants are created through the generic grants API so Terraform receives a stable grant ID."},
			"organization_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Organization ID."},
			"inode_type":      schema.StringAttribute{Computed: true, MarkdownDescription: "Inode type."},
			"created_by":      schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the member who created the grant."},
			"updated_by":      schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the member who last updated the grant."},
			"created_at":      schema.StringAttribute{Computed: true, MarkdownDescription: "Grant creation timestamp."},
			"updated_at":      schema.StringAttribute{Computed: true, MarkdownDescription: "Grant update timestamp."},
		},
	}
}

func (r *grantResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var config grantModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}
	if config.MemberID.IsUnknown() || config.TeamID.IsUnknown() || config.Permission.IsUnknown() || config.TagID.IsUnknown() {
		return
	}
	validateGrant(&config, genericGrantPermissions, true, &response.Diagnostics)
}

func (r *grantResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan grantModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() || !validateGrant(&plan, genericGrantPermissions, true, &response.Diagnostics) {
		return
	}
	value, err := r.client.CreateGrant(ctx, sigma.CreateGrantInput{
		Grantee:    sigma.Grantee{MemberID: plan.MemberID.ValueString(), TeamID: plan.TeamID.ValueString()},
		Permission: plan.Permission.ValueString(),
		InodeID:    plan.InodeID.ValueString(),
		TagID:      plan.TagID.ValueString(),
	})
	if err != nil {
		response.Diagnostics.AddError("Unable to create Sigma grant", err.Error())
		return
	}
	if value == nil {
		value, err = lookupListedGrant(func() ([]sigma.Grant, error) {
			return r.client.ListGrants(ctx, plan.InodeID.ValueString())
		}, &plan, "")
		if err != nil {
			response.Diagnostics.AddError("Unable to locate created Sigma grant", err.Error())
			return
		}
	}
	setGrant(&plan, value)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *grantResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state grantModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	value, err := r.client.GetGrant(ctx, state.ID.ValueString())
	if sigma.IsNotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError("Unable to read Sigma grant", err.Error())
		return
	}
	setGrant(&state, value)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *grantResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {}

func (r *grantResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state grantModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteGrant(ctx, state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
		response.Diagnostics.AddError("Unable to delete Sigma grant", err.Error())
	}
}

func (r *grantResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	importPassthrough(ctx, request, response)
}

var genericGrantPermissions = []string{"admin", "annotate", "update", "usage", "writeback", "create", "organize", "explore", "view", "edit", "apply"}
var workspaceGrantPermissions = []string{"view", "explore", "organize", "edit"}
var workbookGrantPermissions = []string{"view", "explore", "edit"}
var reportGrantPermissions = []string{"view", "edit"}

func validateGrant(plan *grantModel, allowed []string, allowTag bool, response interface{ AddError(string, string) }) bool {
	memberSet := !plan.MemberID.IsNull() && !plan.MemberID.IsUnknown() && plan.MemberID.ValueString() != ""
	teamSet := !plan.TeamID.IsNull() && !plan.TeamID.IsUnknown() && plan.TeamID.ValueString() != ""
	if memberSet == teamSet {
		response.AddError("Invalid Sigma grant grantee", "Exactly one of `member_id` or `team_id` must be set.")
		return false
	}
	if !allowTag && !plan.TagID.IsNull() && !plan.TagID.IsUnknown() && plan.TagID.ValueString() != "" {
		response.AddError("Invalid Sigma workspace grant", "`tag_id` is not supported for workspace grants.")
		return false
	}
	if !contains(allowed, plan.Permission.ValueString()) {
		response.AddError("Invalid Sigma grant permission", fmt.Sprintf("Permission %q is not valid for this resource.", plan.Permission.ValueString()))
		return false
	}
	return true
}

func lookupListedGrant(list func() ([]sigma.Grant, error), model *grantModel, grantID string) (*sigma.Grant, error) {
	values, err := list()
	if err != nil {
		return nil, err
	}
	return lookupGrant(values, model, grantID)
}

func lookupGrant(values []sigma.Grant, model *grantModel, grantID string) (*sigma.Grant, error) {
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
		if grantMatches(&values[i], model) {
			matches = append(matches, values[i])
		}
	}
	switch len(matches) {
	case 1:
		return &matches[0], nil
	case 0:
		return nil, &sigma.APIError{StatusCode: 404, Message: "grant not found"}
	default:
		return nil, fmt.Errorf("multiple grants matched inode/grantee/permission; set a unique permission or import by grant ID (tag-scoped grants may not be distinguishable in list responses)")
	}
}

func grantMatches(value *sigma.Grant, model *grantModel) bool {
	if value.Permission != model.Permission.ValueString() {
		return false
	}
	desiredTag := ""
	if !model.TagID.IsNull() && !model.TagID.IsUnknown() {
		desiredTag = model.TagID.ValueString()
	}
	actualTag := ""
	if value.TagID != nil {
		actualTag = *value.TagID
	}
	// When the API returns tagId, require an exact match. When it omits tagId
	// (current documented list schema), only accept untagged desired grants so
	// tag-scoped grants are not selected by accident.
	if actualTag != "" || desiredTag != "" {
		if actualTag == "" && desiredTag != "" {
			return false
		}
		if actualTag != desiredTag {
			return false
		}
	}
	if !model.MemberID.IsNull() && model.MemberID.ValueString() != "" {
		return value.MemberID != nil && *value.MemberID == model.MemberID.ValueString()
	}
	return value.TeamID != nil && *value.TeamID == model.TeamID.ValueString()
}

func importGrantCompositeID(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts := strings.SplitN(request.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		response.Diagnostics.AddError("Invalid import ID", "Use `inodeId/grantId`.")
		return
	}
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("inode_id"), parts[0])...)
}

func setGrant(state *grantModel, value *sigma.Grant) {
	priorTag := state.TagID
	state.ID = types.StringValue(value.GrantID)
	state.InodeID = types.StringValue(value.InodeID)
	state.MemberID = nullableString(value.MemberID)
	state.TeamID = nullableString(value.TeamID)
	state.Permission = types.StringValue(value.Permission)
	if value.TagID != nil {
		state.TagID = types.StringValue(*value.TagID)
	} else if !priorTag.IsNull() && !priorTag.IsUnknown() {
		// List/get schemas may omit tagId; keep the configured tag identity in state.
		state.TagID = priorTag
	} else {
		state.TagID = types.StringNull()
	}
	state.OrganizationID = types.StringValue(value.OrganizationID)
	state.InodeType = types.StringValue(value.InodeType)
	state.CreatedBy = types.StringValue(value.CreatedBy)
	state.UpdatedBy = types.StringValue(value.UpdatedBy)
	state.CreatedAt = types.StringValue(value.CreatedAt)
	state.UpdatedAt = types.StringValue(value.UpdatedAt)
}
