package provider

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const scheduleConfigJSONDescription = "JSON body accepted by the Sigma schedule create/update API. Must be a JSON object. Do not set top-level `suspensionAction`; use `is_suspended` instead. List/get responses omit `target` and echo extra `configV2` fields (for example `includeLink` or `notificationAttachments`), so Terraform retains the configured object shape from prior state and overlays API values only onto keys that were already set."

type scheduleRefresh struct {
	Schedule    json.RawMessage
	ConfigV2    json.RawMessage
	IsSuspended bool
}

func scheduleSchema(description, parentName, parentDescription string) schema.Schema {
	return schema.Schema{
		MarkdownDescription: description,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true, MarkdownDescription: "Scheduled notification ID."},
			parentName: schema.StringAttribute{
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: parentDescription,
			},
			"config_json": schema.StringAttribute{
				Required:            true,
				CustomType:          jsontypes.NormalizedType{},
				MarkdownDescription: scheduleConfigJSONDescription,
			},
			"is_suspended": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Whether the schedule is paused. On update, Terraform sends `pause` or `resume` only when this value changes.",
			},
		},
	}
}

func decodeJSONObject(raw string) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	var decoded any
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		diags.AddError("Invalid config_json", err.Error())
		return nil, diags
	}
	object, ok := decoded.(map[string]any)
	if !ok {
		diags.AddError("Invalid config_json", "config_json must be a JSON object.")
		return nil, diags
	}
	return object, diags
}

func scheduleConfigForRequest(value jsontypes.Normalized) (json.RawMessage, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() || value.IsUnknown() {
		diags.AddError("Invalid config_json", "config_json must be a known JSON object.")
		return nil, diags
	}
	object, objectDiags := decodeJSONObject(value.ValueString())
	diags.Append(objectDiags...)
	if diags.HasError() {
		return nil, diags
	}
	if _, exists := object["suspensionAction"]; exists {
		diags.AddError("Invalid config_json", "Top-level `suspensionAction` is not allowed in `config_json`; set `is_suspended` instead.")
		return nil, diags
	}
	return json.RawMessage(value.ValueString()), diags
}

func scheduleUpdateBody(config jsontypes.Normalized, planSuspended, stateSuspended types.Bool) (json.RawMessage, diag.Diagnostics) {
	body, diags := scheduleConfigForRequest(config)
	if diags.HasError() {
		return nil, diags
	}
	if planSuspended.IsUnknown() || planSuspended.Equal(stateSuspended) {
		return body, diags
	}
	if planSuspended.IsNull() {
		return body, diags
	}
	object, objectDiags := decodeJSONObject(string(body))
	diags.Append(objectDiags...)
	if diags.HasError() {
		return nil, diags
	}
	if planSuspended.ValueBool() {
		object["suspensionAction"] = "pause"
	} else {
		object["suspensionAction"] = "resume"
	}
	encoded, err := json.Marshal(object)
	if err != nil {
		diags.AddError("Invalid schedule update body", err.Error())
		return nil, diags
	}
	return encoded, diags
}

func mergeScheduleConfig(prior jsontypes.Normalized, refresh scheduleRefresh) (jsontypes.Normalized, diag.Diagnostics) {
	var diags diag.Diagnostics
	object := map[string]any{}
	if !prior.IsNull() && !prior.IsUnknown() {
		decoded, decodeDiags := decodeJSONObject(prior.ValueString())
		diags.Append(decodeDiags...)
		if diags.HasError() {
			return jsontypes.NewNormalizedNull(), diags
		}
		object = decoded
	}
	delete(object, "suspensionAction")
	if overlay, ok := rawJSONValue(refresh.Schedule); ok {
		object["schedule"] = overlayConfiguredJSON(object["schedule"], overlay)
	}
	if overlay, ok := rawJSONValue(refresh.ConfigV2); ok {
		object["configV2"] = overlayConfiguredJSON(object["configV2"], overlay)
	}
	encoded, err := json.Marshal(object)
	if err != nil {
		diags.AddError("Unable to canonicalize schedule config_json", err.Error())
		return jsontypes.NewNormalizedNull(), diags
	}
	return jsontypes.NewNormalizedValue(string(encoded)), diags
}

// overlayConfiguredJSON keeps the configured JSON shape. When both values are
// objects, API keys that were not configured are dropped so live responses
// (defaults, renamed attachment fields) do not rewrite config_json.
func overlayConfiguredJSON(prior, api any) any {
	priorObject, priorOK := prior.(map[string]any)
	apiObject, apiOK := api.(map[string]any)
	if !priorOK || len(priorObject) == 0 {
		return api
	}
	if !apiOK {
		return prior
	}
	merged := make(map[string]any, len(priorObject))
	for key, priorValue := range priorObject {
		apiValue, exists := apiObject[key]
		if !exists {
			merged[key] = priorValue
			continue
		}
		merged[key] = overlayConfiguredJSON(priorValue, apiValue)
	}
	return merged
}

func scheduleIsSuspended(plan types.Bool, api bool) types.Bool {
	if plan.IsNull() || plan.IsUnknown() {
		return types.BoolValue(api)
	}
	return plan
}

func rawJSONValue(raw json.RawMessage) (any, bool) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, false
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, false
	}
	return value, true
}

func suspensionMismatchBody(isSuspended bool) (json.RawMessage, error) {
	action := "resume"
	if isSuspended {
		action = "pause"
	}
	return json.Marshal(map[string]any{"suspensionAction": action})
}

func applyScheduleCreateSuspension(config jsontypes.Normalized, planSuspended types.Bool, apiSuspended bool) (json.RawMessage, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	if planSuspended.IsNull() || planSuspended.IsUnknown() {
		return nil, false, diags
	}
	if planSuspended.ValueBool() == apiSuspended {
		return nil, false, diags
	}
	// Live PATCH ignores a bare suspensionAction; send the configured body with it.
	body, bodyDiags := scheduleUpdateBody(config, planSuspended, types.BoolValue(apiSuspended))
	diags.Append(bodyDiags...)
	return body, !diags.HasError(), diags
}

func validateScheduleConfigJSON(value jsontypes.Normalized) diag.Diagnostics {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	_, diags := scheduleConfigForRequest(value)
	return diags
}
