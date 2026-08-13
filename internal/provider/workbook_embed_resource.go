package provider

import (
	"context"
	"strings"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type workbookEmbedResource struct{ configuredResource }

var (
	_ resource.Resource                = (*workbookEmbedResource)(nil)
	_ resource.ResourceWithConfigure   = (*workbookEmbedResource)(nil)
	_ resource.ResourceWithImportState = (*workbookEmbedResource)(nil)
)

type workbookEmbedModel struct {
	ID         types.String `tfsdk:"id"`
	WorkbookID types.String `tfsdk:"workbook_id"`
	EmbedType  types.String `tfsdk:"embed_type"`
	SourceType types.String `tfsdk:"source_type"`
	SourceID   types.String `tfsdk:"source_id"`
	EmbedURL   types.String `tfsdk:"embed_url"`
}

func NewWorkbookEmbedResource() resource.Resource { return &workbookEmbedResource{} }

func (r *workbookEmbedResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workbook_embed"
}
func (r *workbookEmbedResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *workbookEmbedResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Sigma workbook embed. Embeds are immutable after create; Sigma's list API returns `public` instead of `embed_type`.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Embed ID."},
			"workbook_id": schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Workbook ID."},
			"embed_type":  schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Embed type: `public`, `secure`, or `application`. `secure` and `application` are deprecated by Sigma."},
			"source_type": schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Source scope: `workbook`, `page`, or `element`."},
			"source_id":   schema.StringAttribute{Optional: true, PlanModifiers: replace, MarkdownDescription: "Page or element ID when `source_type` is `page` or `element`."},
			"embed_url":   schema.StringAttribute{Computed: true, MarkdownDescription: "Embed URL returned by Sigma."},
		},
	}
}
func setWorkbookEmbed(state *workbookEmbedModel, value *sigma.WorkbookEmbed) {
	state.ID = types.StringValue(value.EmbedID)
	state.EmbedURL = types.StringValue(value.EmbedURL)
	state.SourceType = types.StringValue(value.SourceType)
	if value.SourceID != nil && *value.SourceID != "" {
		state.SourceID = types.StringValue(*value.SourceID)
	} else if state.SourceID.IsUnknown() {
		state.SourceID = types.StringNull()
	}
	if value.Public {
		state.EmbedType = types.StringValue("public")
	}
}
func (r *workbookEmbedResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workbookEmbedModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.CreateWorkbookEmbed(ctx, plan.WorkbookID.ValueString(), sigma.CreateWorkbookEmbedInput{
		EmbedType: plan.EmbedType.ValueString(), SourceType: plan.SourceType.ValueString(), SourceID: plan.SourceID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma workbook embed", err.Error())
		return
	}
	if refreshed, err := r.client.GetWorkbookEmbed(ctx, plan.WorkbookID.ValueString(), value.EmbedID); err == nil {
		value = refreshed
	}
	setWorkbookEmbed(&plan, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *workbookEmbedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workbookEmbedModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.GetWorkbookEmbed(ctx, state.WorkbookID.ValueString(), state.ID.ValueString())
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma workbook embed", err.Error())
		return
	}
	setWorkbookEmbed(&state, value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *workbookEmbedResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
}
func (r *workbookEmbedResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workbookEmbedModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteWorkbookEmbed(ctx, state.WorkbookID.ValueString(), state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma workbook embed", err.Error())
	}
}
func (r *workbookEmbedResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Use `workbookId/embedId`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workbook_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
