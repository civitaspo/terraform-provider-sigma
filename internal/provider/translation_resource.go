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

type translationResource struct{ configuredResource }

var (
	_ resource.Resource                = (*translationResource)(nil)
	_ resource.ResourceWithConfigure   = (*translationResource)(nil)
	_ resource.ResourceWithImportState = (*translationResource)(nil)
)

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
