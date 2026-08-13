package provider

import (
	"encoding/json"
	"strings"

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

func mergeJSON(base types.String, overlay types.String) (json.RawMessage, error) {
	var left map[string]any
	if base.IsNull() || base.IsUnknown() {
		left = map[string]any{}
	} else if err := json.Unmarshal([]byte(base.ValueString()), &left); err != nil {
		return nil, err
	}
	if !overlay.IsNull() && !overlay.IsUnknown() {
		var right map[string]any
		if err := json.Unmarshal([]byte(overlay.ValueString()), &right); err != nil {
			return nil, err
		}
		for key, value := range right {
			left[key] = value
		}
	}
	return json.Marshal(left)
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

func jsonValuesEqual(left, right types.String) bool {
	if left.Equal(right) {
		return true
	}
	if left.IsNull() || right.IsNull() || left.IsUnknown() || right.IsUnknown() {
		return false
	}
	leftCanonical, leftErr := canonicalJSON([]byte(left.ValueString()))
	rightCanonical, rightErr := canonicalJSON([]byte(right.ValueString()))
	if leftErr != nil || rightErr != nil {
		return false
	}
	return leftCanonical == rightCanonical
}
