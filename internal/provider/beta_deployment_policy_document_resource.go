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

type deploymentPolicyDocumentResource struct{ configuredResource }

var (
	_ resource.Resource                = (*deploymentPolicyDocumentResource)(nil)
	_ resource.ResourceWithConfigure   = (*deploymentPolicyDocumentResource)(nil)
	_ resource.ResourceWithImportState = (*deploymentPolicyDocumentResource)(nil)
)

type deploymentPolicyDocumentModel struct {
	ID                 types.String `tfsdk:"id"`
	DeploymentPolicyID types.String `tfsdk:"deployment_policy_id"`
	InodeID            types.String `tfsdk:"inode_id"`
}

func NewDeploymentPolicyDocumentResource() resource.Resource {
	return &deploymentPolicyDocumentResource{}
}

func (r *deploymentPolicyDocumentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment_policy_document"
}
func (r *deploymentPolicyDocumentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *deploymentPolicyDocumentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Attaches exactly one document inode to an existing Sigma deployment policy. This resource never creates the parent policy. " + betaAPINotice,
		Attributes: map[string]schema.Attribute{
			"id":                   schema.StringAttribute{Computed: true, MarkdownDescription: "Composite ID `deploymentPolicyId/inodeId`."},
			"deployment_policy_id": schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Existing deployment policy ID."},
			"inode_id":             schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Document inode ID to attach."},
		},
	}
}
func (r *deploymentPolicyDocumentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan deploymentPolicyDocumentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	policyID, policyDiags := knownString(plan.DeploymentPolicyID, "deployment_policy_id")
	inodeID, inodeDiags := knownString(plan.InodeID, "inode_id")
	resp.Diagnostics.Append(policyDiags...)
	resp.Diagnostics.Append(inodeDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.AddDeploymentPolicyInodes(ctx, policyID, []string{inodeID}); err != nil {
		resp.Diagnostics.AddError("Unable to add Sigma deployment policy document", err.Error())
		return
	}
	plan.ID = types.StringValue(policyID + "/" + inodeID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *deploymentPolicyDocumentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state deploymentPolicyDocumentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	policyID, policyDiags := knownString(state.DeploymentPolicyID, "deployment_policy_id")
	inodeID, inodeDiags := knownString(state.InodeID, "inode_id")
	resp.Diagnostics.Append(policyDiags...)
	resp.Diagnostics.Append(inodeDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	files, err := r.client.ListDeploymentPolicyInodes(ctx, policyID)
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma deployment policy documents", err.Error())
		return
	}
	found := false
	for _, file := range files {
		if file.InodeID == inodeID {
			found = true
			break
		}
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}
	state.ID = types.StringValue(policyID + "/" + inodeID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *deploymentPolicyDocumentResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Unexpected update", "sigma_deployment_policy_document has no mutable attributes.")
}
func (r *deploymentPolicyDocumentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state deploymentPolicyDocumentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	policyID, policyDiags := knownString(state.DeploymentPolicyID, "deployment_policy_id")
	inodeID, inodeDiags := knownString(state.InodeID, "inode_id")
	resp.Diagnostics.Append(policyDiags...)
	resp.Diagnostics.Append(inodeDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.RemoveDeploymentPolicyInode(ctx, policyID, inodeID); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to remove Sigma deployment policy document", err.Error())
	}
}
func (r *deploymentPolicyDocumentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	policyID, inodeID, ok := splitCompositeImportID(req.ID)
	if !ok {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected deploymentPolicyId/inodeId, got %q.", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("deployment_policy_id"), policyID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("inode_id"), inodeID)...)
}
