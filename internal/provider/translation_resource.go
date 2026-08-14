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
		MarkdownDescription: "Manages a Sigma organization translation file. `translations` is required and authoritative. When `lng_variant` is set, the provider uses the variant translation endpoints.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Translation ID (`lng` or `lng/variant`)."},
			"lng":         schema.StringAttribute{Required: true, PlanModifiers: replace, MarkdownDescription: "Locale identifier."},
			"lng_variant": schema.StringAttribute{Optional: true, PlanModifiers: replace, MarkdownDescription: "Optional custom translation variant."},
			"translations": schema.MapAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Authoritative map of phrases to translated strings. An empty map clears stored phrases. A missing API map is stored as an empty map, not null.",
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

func translationVariant(value types.String) string {
	if value.IsNull() || value.IsUnknown() {
		return ""
	}
	return value.ValueString()
}

func (r *translationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan translationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	lng, diags := knownString(plan.Lng, "lng")
	resp.Diagnostics.Append(diags...)
	translations, mapDiags := knownStringMap(ctx, plan.Translations, "translations")
	resp.Diagnostics.Append(mapDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	variant := translationVariant(plan.LngVariant)
	if err := r.client.CreateOrgTranslation(ctx, sigma.CreateOrgTranslationInput{
		Lng: lng, LngVariant: variant, Translations: translations,
	}); err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma translation", err.Error())
		return
	}
	value, err := r.client.GetOrgTranslation(ctx, lng, variant)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma translation after create", err.Error())
		return
	}
	plan.ID = types.StringValue(translationID(lng, variant))
	mapped, mapValueDiags := stringMapValue(ctx, value.Translations)
	resp.Diagnostics.Append(mapValueDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Translations = mapped
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *translationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state translationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	lng, diags := knownString(state.Lng, "lng")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	variant := translationVariant(state.LngVariant)
	value, err := r.client.GetOrgTranslation(ctx, lng, variant)
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma translation", err.Error())
		return
	}
	state.ID = types.StringValue(translationID(lng, variant))
	mapped, mapValueDiags := stringMapValue(ctx, value.Translations)
	resp.Diagnostics.Append(mapValueDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Translations = mapped
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *translationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan translationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	lng, diags := knownString(plan.Lng, "lng")
	resp.Diagnostics.Append(diags...)
	translations, mapDiags := knownStringMap(ctx, plan.Translations, "translations")
	resp.Diagnostics.Append(mapDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	variant := translationVariant(plan.LngVariant)
	if err := r.client.UpdateOrgTranslation(ctx, lng, variant, sigma.UpdateOrgTranslationInput{
		Translations: translations,
	}); err != nil {
		resp.Diagnostics.AddError("Unable to update Sigma translation", err.Error())
		return
	}
	value, err := r.client.GetOrgTranslation(ctx, lng, variant)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma translation after update", err.Error())
		return
	}
	plan.ID = types.StringValue(translationID(lng, variant))
	mapped, mapValueDiags := stringMapValue(ctx, value.Translations)
	resp.Diagnostics.Append(mapValueDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Translations = mapped
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *translationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state translationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	lng, diags := knownString(state.Lng, "lng")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteOrgTranslation(ctx, lng, translationVariant(state.LngVariant)); err != nil && !sigma.IsNotFound(err) {
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
