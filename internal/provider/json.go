package provider

import (
	"encoding/json"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func canonicalJSON(raw []byte) (string, error) {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", err
	}
	encoded, err := json.Marshal(value)
	return string(encoded), err
}

func rawJSON(value types.String) (json.RawMessage, error) {
	if value.IsNull() || value.IsUnknown() || strings.TrimSpace(value.ValueString()) == "" {
		return nil, nil
	}
	canonical, err := canonicalJSON([]byte(value.ValueString()))
	return json.RawMessage(canonical), err
}

func jsonString(value json.RawMessage) types.String {
	if len(value) == 0 || string(value) == "null" {
		return types.StringNull()
	}
	canonical, err := canonicalJSON(value)
	if err != nil {
		return types.StringNull()
	}
	return types.StringValue(canonical)
}

func knownNormalizedObject(value jsontypes.Normalized, attribute string) (map[string]any, json.RawMessage, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() || value.IsUnknown() {
		diags.AddError("Invalid "+attribute, attribute+" must be a known JSON object.")
		return nil, nil, diags
	}
	var decoded any
	if err := json.Unmarshal([]byte(value.ValueString()), &decoded); err != nil {
		diags.AddError("Invalid "+attribute, err.Error())
		return nil, nil, diags
	}
	object, ok := decoded.(map[string]any)
	if !ok {
		diags.AddError("Invalid "+attribute, attribute+" must be a JSON object.")
		return nil, nil, diags
	}
	encoded, err := json.Marshal(object)
	if err != nil {
		diags.AddError("Invalid "+attribute, err.Error())
		return nil, nil, diags
	}
	return object, json.RawMessage(encoded), diags
}

func optionalNormalizedJSON(value jsontypes.Normalized, attribute string) (json.RawMessage, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() || value.IsUnknown() {
		return nil, diags
	}
	_, encoded, objectDiags := knownNormalizedObject(value, attribute)
	diags.Append(objectDiags...)
	return encoded, diags
}

func normalizedFromRaw(value json.RawMessage) jsontypes.Normalized {
	if len(value) == 0 || string(value) == "null" {
		return jsontypes.NewNormalizedNull()
	}
	canonical, err := canonicalJSON(value)
	if err != nil {
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(canonical)
}

func mergeObjectWithWriteOnly(base map[string]any, overlay types.String, overlayName string) (json.RawMessage, diag.Diagnostics) {
	var diags diag.Diagnostics
	merged := map[string]any{}
	for key, value := range base {
		merged[key] = value
	}
	if overlay.IsNull() || overlay.IsUnknown() || strings.TrimSpace(overlay.ValueString()) == "" {
		encoded, err := json.Marshal(merged)
		if err != nil {
			diags.AddError("Invalid "+overlayName, err.Error())
			return nil, diags
		}
		return json.RawMessage(encoded), diags
	}
	var right map[string]any
	if err := json.Unmarshal([]byte(overlay.ValueString()), &right); err != nil {
		diags.AddError("Invalid "+overlayName, err.Error())
		return nil, diags
	}
	for key, value := range right {
		merged[key] = value
	}
	encoded, err := json.Marshal(merged)
	if err != nil {
		diags.AddError("Invalid "+overlayName, err.Error())
		return nil, diags
	}
	return json.RawMessage(encoded), diags
}

func writeOnlyVersionPair(payload types.String, version types.Int64, payloadName, versionName string) diag.Diagnostics {
	var diags diag.Diagnostics
	payloadSet := !payload.IsNull() && !payload.IsUnknown() && strings.TrimSpace(payload.ValueString()) != ""
	versionSet := !version.IsNull() && !version.IsUnknown()
	if payloadSet != versionSet {
		diags.AddError(
			"Invalid write-only secret pairing",
			payloadName+" and "+versionName+" must be supplied together.",
		)
	}
	return diags
}
