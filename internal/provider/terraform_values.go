package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

const betaAPINotice = "This resource uses a Sigma Beta API and may change without notice."

const betaDataSourceNotice = "This resource uses a Sigma Beta API and may change without notice."

func nullableString(value *string) types.String {
	if value == nil {
		return types.StringNull()
	}
	return types.StringValue(*value)
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func stringSetDiff(desired, current []string) (add, remove []string) {
	have, want := map[string]bool{}, map[string]bool{}
	for _, id := range current {
		have[id] = true
	}
	for _, id := range desired {
		want[id] = true
	}
	for id := range want {
		if !have[id] {
			add = append(add, id)
		}
	}
	for id := range have {
		if !want[id] {
			remove = append(remove, id)
		}
	}
	return add, remove
}

func optionalStringPtr(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueString()
	return &v
}

func optionalBoolPtr(value types.Bool) *bool {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueBool()
	return &v
}

func stringOrNull(value *string) types.String {
	if value == nil || *value == "" {
		return types.StringNull()
	}
	return types.StringValue(*value)
}
