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

type workbookEmbedResource struct{ configuredResource }

var (
	_ resource.Resource                   = (*workbookEmbedResource)(nil)
	_ resource.ResourceWithConfigure      = (*workbookEmbedResource)(nil)
	_ resource.ResourceWithImportState    = (*workbookEmbedResource)(nil)
	_ resource.ResourceWithValidateConfig = (*workbookEmbedResource)(nil)
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
		MarkdownDescription: "Manages a public Sigma workbook embed. Only `embed_type = public` is supported. Deprecated `secure` and `application` embeds cannot round-trip because list responses expose `public` rather than `embed_type`.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Embed ID."},
			"workbook_id": schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Workbook ID."},
			"embed_type":  schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Embed type. Only `public` is supported."},
			"source_type": schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Source scope: `workbook`, `page`, or `element`."},
			"source_id":   schema.StringAttribute{Optional: true, PlanModifiers: replace, MarkdownDescription: "Page or element ID when `source_type` is `page` or `element`."},
			"embed_url":   schema.StringAttribute{Computed: true, MarkdownDescription: "Embed URL returned by Sigma."},
		},
	}
}

func (r *workbookEmbedResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config workbookEmbedModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if config.EmbedType.IsNull() || config.EmbedType.IsUnknown() {
		return
	}
	if config.EmbedType.ValueString() != "public" {
		resp.Diagnostics.AddAttributeError(
			path.Root("embed_type"),
			"Unsupported Sigma workbook embed type",
			"`embed_type` must be `public`. Deprecated `secure` and `application` embeds are not supported.",
		)
	}
}

func applyPublicWorkbookEmbed(state *workbookEmbedModel, value *sigma.WorkbookEmbed) error {
	if value == nil {
		return fmt.Errorf("empty workbook embed response")
	}
	if !value.Public {
		return fmt.Errorf("workbook embed %s is not public; sigma_workbook_embed only manages public embeds", value.EmbedID)
	}
	state.ID = types.StringValue(value.EmbedID)
	state.EmbedURL = types.StringValue(value.EmbedURL)
	state.SourceType = types.StringValue(value.SourceType)
	state.EmbedType = types.StringValue("public")
	if !state.SourceID.IsNull() && !state.SourceID.IsUnknown() {
		if value.SourceID != nil && *value.SourceID != "" {
			state.SourceID = types.StringValue(*value.SourceID)
		}
	} else if state.SourceID.IsUnknown() {
		if value.SourceID != nil && *value.SourceID != "" {
			state.SourceID = types.StringValue(*value.SourceID)
		} else {
			state.SourceID = types.StringNull()
		}
	}
	return nil
}

func (r *workbookEmbedResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workbookEmbedModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	workbookID, diags := knownString(plan.WorkbookID, "workbook_id")
	resp.Diagnostics.Append(diags...)
	embedType, diags := knownString(plan.EmbedType, "embed_type")
	resp.Diagnostics.Append(diags...)
	sourceType, diags := knownString(plan.SourceType, "source_type")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if embedType != "public" {
		resp.Diagnostics.AddError("Unsupported Sigma workbook embed type", "`embed_type` must be `public`.")
		return
	}
	input := sigma.CreateWorkbookEmbedInput{EmbedType: embedType, SourceType: sourceType}
	if sourceID := optionalStringPtr(plan.SourceID); sourceID != nil && *sourceID != "" {
		input.SourceID = *sourceID
	}
	value, err := r.client.CreateWorkbookEmbed(ctx, workbookID, input)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma workbook embed", err.Error())
		return
	}
	if refreshed, refreshErr := r.client.GetWorkbookEmbed(ctx, workbookID, value.EmbedID); refreshErr == nil {
		value = refreshed
	}
	if err := applyPublicWorkbookEmbed(&plan, value); err != nil {
		resp.Diagnostics.AddError("Created Sigma workbook embed is not public", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workbookEmbedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workbookEmbedModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	workbookID, diags := knownString(state.WorkbookID, "workbook_id")
	resp.Diagnostics.Append(diags...)
	embedID, diags := knownString(state.ID, "id")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.GetWorkbookEmbed(ctx, workbookID, embedID)
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma workbook embed", err.Error())
		return
	}
	if err := applyPublicWorkbookEmbed(&state, value); err != nil {
		resp.Diagnostics.AddError("Sigma workbook embed is not public", err.Error())
		return
	}
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
	workbookID, diags := knownString(state.WorkbookID, "workbook_id")
	resp.Diagnostics.Append(diags...)
	embedID, diags := knownString(state.ID, "id")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteWorkbookEmbed(ctx, workbookID, embedID); err != nil && !sigma.IsNotFound(err) {
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
