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

// grantModel, grantSchema, and lookup helpers are shared by specialized grant
// resources. Public sigma_grant was removed in v0.2.
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

const documentGrantTagMarkdown = "Optional version tag ID. Changing this forces a new resource. Tagged workbook and report grants use generic `POST /v2/grants`, `GET /v2/grants/{grantId}`, and `DELETE /v2/grants/{grantId}` because the dedicated POST returns `{}` and list/get responses omit `tagId`. Terraform preserves the configured `tag_id` after refresh; import cannot reconstruct it from the API."

const workspaceGrantTagMarkdown = "Optional version tag ID. Not supported for workspace grants."

func grantSchema(description, inodeDescription, permissionDescription, tagDescription string) schema.Schema {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	return schema.Schema{
		MarkdownDescription: description,
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Grant ID."},
			"inode_id":   schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: inodeDescription},
			"member_id":  schema.StringAttribute{Optional: true, PlanModifiers: replace, MarkdownDescription: "Member ID receiving the grant. Exactly one of `member_id` or `team_id` must be set."},
			"team_id":    schema.StringAttribute{Optional: true, PlanModifiers: replace, MarkdownDescription: "Team ID receiving the grant. Exactly one of `member_id` or `team_id` must be set."},
			"permission": schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: permissionDescription},
			"tag_id": schema.StringAttribute{
				Optional:            true,
				PlanModifiers:       replace,
				MarkdownDescription: tagDescription,
			},
			"organization_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Organization ID."},
			"inode_type":      schema.StringAttribute{Computed: true, MarkdownDescription: "Inode type."},
			"created_by":      schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the member who created the grant."},
			"updated_by":      schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the member who last updated the grant."},
			"created_at":      schema.StringAttribute{Computed: true, MarkdownDescription: "Grant creation timestamp."},
			"updated_at":      schema.StringAttribute{Computed: true, MarkdownDescription: "Grant update timestamp."},
		},
	}
}

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
		// List-by-inode also returns inherited parent grants (workspace view
		// plus document view to the same team). Prefer the grant on the
		// configured inode when that uniquely identifies it.
		desired := model.InodeID.ValueString()
		if desired != "" {
			var exact []sigma.Grant
			for _, match := range matches {
				if match.InodeID == desired {
					exact = append(exact, match)
				}
			}
			if len(exact) == 1 {
				return &exact[0], nil
			}
		}
		return nil, fmt.Errorf("multiple grants matched inode, grantee, and permission; refusing to select the first match")
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

func taggedDocumentGrant(model *grantModel) bool {
	return !model.TagID.IsNull() && !model.TagID.IsUnknown() && model.TagID.ValueString() != ""
}

func grantGrantee(model *grantModel) sigma.Grantee {
	var grantee sigma.Grantee
	if !model.MemberID.IsNull() && !model.MemberID.IsUnknown() {
		grantee.MemberID = model.MemberID.ValueString()
	}
	if !model.TeamID.IsNull() && !model.TeamID.IsUnknown() {
		grantee.TeamID = model.TeamID.ValueString()
	}
	return grantee
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
	priorInode := state.InodeID
	state.ID = types.StringValue(value.GrantID)
	// List/get may return a URL id while configuration used the UUID.
	if !priorInode.IsNull() && !priorInode.IsUnknown() && priorInode.ValueString() != "" {
		state.InodeID = priorInode
	} else {
		state.InodeID = types.StringValue(value.InodeID)
	}
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
