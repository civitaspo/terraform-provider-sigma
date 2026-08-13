package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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

func changedStringPtr(plan, state types.String) *string {
	if plan.IsUnknown() || plan.IsNull() || plan.Equal(state) {
		return nil
	}
	value := plan.ValueString()
	return &value
}

func changedBoolPtr(plan, state types.Bool) *bool {
	if plan.IsUnknown() || plan.IsNull() || plan.Equal(state) {
		return nil
	}
	value := plan.ValueBool()
	return &value
}

func knownTrue(value types.Bool) bool {
	return !value.IsNull() && !value.IsUnknown() && value.ValueBool()
}

func knownString(value types.String, attribute string) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() || value.IsUnknown() {
		diags.AddError("Invalid "+attribute, attribute+" must be a known, non-null value.")
		return "", diags
	}
	return value.ValueString(), diags
}

func knownStringSet(ctx context.Context, value types.Set, attribute string) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() || value.IsUnknown() {
		diags.AddError("Invalid "+attribute, attribute+" must be a known, non-null set.")
		return nil, diags
	}
	var items []string
	diags.Append(value.ElementsAs(ctx, &items, false)...)
	if diags.HasError() {
		return nil, diags
	}
	return items, diags
}

func stringSetValue(ctx context.Context, values []string) (types.Set, diag.Diagnostics) {
	if values == nil {
		values = make([]string, 0)
	}
	return types.SetValueFrom(ctx, types.StringType, values)
}

func knownStringMap(ctx context.Context, value types.Map, attribute string) (map[string]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() || value.IsUnknown() {
		diags.AddError("Invalid "+attribute, attribute+" must be a known, non-null map.")
		return nil, diags
	}
	items := map[string]string{}
	diags.Append(value.ElementsAs(ctx, &items, false)...)
	if diags.HasError() {
		return nil, diags
	}
	if items == nil {
		items = map[string]string{}
	}
	return items, diags
}

func stringMapValue(ctx context.Context, values map[string]string) (types.Map, diag.Diagnostics) {
	if values == nil {
		values = map[string]string{}
	}
	return types.MapValueFrom(ctx, types.StringType, values)
}
