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

var (
	_ resource.Resource                = (*tagResource)(nil)
	_ resource.ResourceWithImportState = (*tagResource)(nil)
	_ resource.Resource                = (*workbookScheduleResource)(nil)
	_ resource.ResourceWithImportState = (*workbookScheduleResource)(nil)
	_ resource.Resource                = (*reportScheduleResource)(nil)
	_ resource.ResourceWithImportState = (*reportScheduleResource)(nil)
	_ resource.Resource                = (*workbookEmbedResource)(nil)
	_ resource.ResourceWithImportState = (*workbookEmbedResource)(nil)
	_ resource.Resource                = (*translationResource)(nil)
	_ resource.ResourceWithImportState = (*translationResource)(nil)
)

type tagResource struct{ configuredResource }

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

type workbookScheduleResource struct{ configuredResource }

type workbookScheduleModel struct {
	ID         types.String `tfsdk:"id"`
	WorkbookID types.String `tfsdk:"workbook_id"`
	ConfigJSON types.String `tfsdk:"config_json"`
}

func NewWorkbookScheduleResource() resource.Resource { return &workbookScheduleResource{} }

func (r *workbookScheduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workbook_schedule"
}
func (r *workbookScheduleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *workbookScheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Sigma workbook export schedule. Schedule payloads are polymorphic (`target`, `schedule`, `configV2`, and optional fields), so configure them with `config_json`. Sigma list responses omit `target`, so Terraform preserves `config_json` from configuration after create.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true, MarkdownDescription: "Scheduled notification ID."},
			"workbook_id": schema.StringAttribute{
				Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "Workbook ID.",
			},
			"config_json": schema.StringAttribute{Required: true, MarkdownDescription: "JSON body accepted by the workbook schedule create/update API."},
		},
	}
}
func (r *workbookScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workbookScheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, err := rawJSON(plan.ConfigJSON)
	if err != nil {
		resp.Diagnostics.AddError("Invalid workbook schedule config_json", err.Error())
		return
	}
	value, err := r.client.CreateWorkbookSchedule(ctx, plan.WorkbookID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma workbook schedule", err.Error())
		return
	}
	plan.ID = types.StringValue(value.ScheduledNotificationID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *workbookScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workbookScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.GetWorkbookSchedule(ctx, state.WorkbookID.ValueString(), state.ID.ValueString())
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma workbook schedule", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *workbookScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan workbookScheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, err := rawJSON(plan.ConfigJSON)
	if err != nil {
		resp.Diagnostics.AddError("Invalid workbook schedule config_json", err.Error())
		return
	}
	value, err := r.client.UpdateWorkbookSchedule(ctx, plan.WorkbookID.ValueString(), plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Sigma workbook schedule", err.Error())
		return
	}
	plan.ID = types.StringValue(value.ScheduledNotificationID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *workbookScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workbookScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteWorkbookSchedule(ctx, state.WorkbookID.ValueString(), state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma workbook schedule", err.Error())
	}
}
func (r *workbookScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Use `workbookId/scheduleId`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workbook_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

type reportScheduleResource struct{ configuredResource }

type reportScheduleModel struct {
	ID         types.String `tfsdk:"id"`
	ReportID   types.String `tfsdk:"report_id"`
	ConfigJSON types.String `tfsdk:"config_json"`
}

func NewReportScheduleResource() resource.Resource { return &reportScheduleResource{} }

func (r *reportScheduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_report_schedule"
}
func (r *reportScheduleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *reportScheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Sigma report export schedule. Schedule payloads are polymorphic, so configure them with `config_json`. Sigma list responses omit `target`, so Terraform preserves `config_json` from configuration after create.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true, MarkdownDescription: "Scheduled notification ID."},
			"report_id": schema.StringAttribute{
				Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "Report ID.",
			},
			"config_json": schema.StringAttribute{Required: true, MarkdownDescription: "JSON body accepted by the report schedule create/update API."},
		},
	}
}
func (r *reportScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan reportScheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, err := rawJSON(plan.ConfigJSON)
	if err != nil {
		resp.Diagnostics.AddError("Invalid report schedule config_json", err.Error())
		return
	}
	value, err := r.client.CreateReportSchedule(ctx, plan.ReportID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma report schedule", err.Error())
		return
	}
	plan.ID = types.StringValue(value.ScheduledNotificationID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *reportScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state reportScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.GetReportSchedule(ctx, state.ReportID.ValueString(), state.ID.ValueString())
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma report schedule", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *reportScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan reportScheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, err := rawJSON(plan.ConfigJSON)
	if err != nil {
		resp.Diagnostics.AddError("Invalid report schedule config_json", err.Error())
		return
	}
	value, err := r.client.UpdateReportSchedule(ctx, plan.ReportID.ValueString(), plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Sigma report schedule", err.Error())
		return
	}
	plan.ID = types.StringValue(value.ScheduledNotificationID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *reportScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state reportScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteReportSchedule(ctx, state.ReportID.ValueString(), state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma report schedule", err.Error())
	}
}
func (r *reportScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Use `reportId/scheduleId`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("report_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

type workbookEmbedResource struct{ configuredResource }

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

type translationResource struct{ configuredResource }

type translationModel struct {
	ID           types.String `tfsdk:"id"`
	Lng          types.String `tfsdk:"lng"`
	LngVariant   types.String `tfsdk:"lng_variant"`
	Translations types.Map    `tfsdk:"translations"`
}

func NewTranslationResource() resource.Resource { return &translationResource{} }

func (r *translationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_translation"
}
func (r *translationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *translationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Sigma organization translation file. When `lng_variant` is set, the provider uses the variant translation endpoints.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Translation ID (`lng` or `lng/variant`)."},
			"lng":         schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Locale identifier."},
			"lng_variant": schema.StringAttribute{Optional: true, PlanModifiers: replace, MarkdownDescription: "Optional custom translation variant."},
			"translations": schema.MapAttribute{
				Optional: true, Computed: true, ElementType: types.StringType,
				MarkdownDescription: "Map of phrases to translated strings.",
			},
		},
	}
}
func translationID(lng, variant string) string {
	if variant == "" {
		return lng
	}
	return lng + "/" + variant
}
func (r *translationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan translationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	translations := map[string]string{}
	if !plan.Translations.IsNull() && !plan.Translations.IsUnknown() {
		resp.Diagnostics.Append(plan.Translations.ElementsAs(ctx, &translations, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if err := r.client.CreateOrgTranslation(ctx, sigma.CreateOrgTranslationInput{
		Lng: plan.Lng.ValueString(), LngVariant: plan.LngVariant.ValueString(), Translations: translations,
	}); err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma translation", err.Error())
		return
	}
	value, err := r.client.GetOrgTranslation(ctx, plan.Lng.ValueString(), plan.LngVariant.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma translation after create", err.Error())
		return
	}
	plan.ID = types.StringValue(translationID(plan.Lng.ValueString(), plan.LngVariant.ValueString()))
	if value.Translations == nil {
		value.Translations = map[string]string{}
	}
	plan.Translations, _ = types.MapValueFrom(ctx, types.StringType, value.Translations)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *translationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state translationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.GetOrgTranslation(ctx, state.Lng.ValueString(), state.LngVariant.ValueString())
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma translation", err.Error())
		return
	}
	state.ID = types.StringValue(translationID(state.Lng.ValueString(), state.LngVariant.ValueString()))
	if value.Translations == nil {
		value.Translations = map[string]string{}
	}
	state.Translations, _ = types.MapValueFrom(ctx, types.StringType, value.Translations)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *translationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan translationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	translations := map[string]string{}
	if !plan.Translations.IsNull() && !plan.Translations.IsUnknown() {
		resp.Diagnostics.Append(plan.Translations.ElementsAs(ctx, &translations, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if err := r.client.UpdateOrgTranslation(ctx, plan.Lng.ValueString(), plan.LngVariant.ValueString(), sigma.UpdateOrgTranslationInput{
		Translations: translations,
	}); err != nil {
		resp.Diagnostics.AddError("Unable to update Sigma translation", err.Error())
		return
	}
	value, err := r.client.GetOrgTranslation(ctx, plan.Lng.ValueString(), plan.LngVariant.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma translation after update", err.Error())
		return
	}
	plan.ID = types.StringValue(translationID(plan.Lng.ValueString(), plan.LngVariant.ValueString()))
	if value.Translations == nil {
		value.Translations = map[string]string{}
	}
	plan.Translations, _ = types.MapValueFrom(ctx, types.StringType, value.Translations)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *translationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state translationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteOrgTranslation(ctx, state.Lng.ValueString(), state.LngVariant.ValueString()); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma translation", err.Error())
	}
}
func (r *translationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if parts[0] == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Use `lng` or `lng/variant`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("lng"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	if len(parts) == 2 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("lng_variant"), parts[1])...)
	}
}
