package provider

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestScheduleConfigForRequestRejectsSuspensionAction(t *testing.T) {
	t.Parallel()

	_, diags := scheduleConfigForRequest(jsontypes.NewNormalizedValue(`{"schedule":{"cronSpec":"0 9 * * 1"},"suspensionAction":"pause"}`))
	if !diags.HasError() {
		t.Fatal("expected suspensionAction to be rejected")
	}
}

func TestScheduleConfigForRequestRejectsNonObject(t *testing.T) {
	t.Parallel()

	_, diags := scheduleConfigForRequest(jsontypes.NewNormalizedValue(`["not","an","object"]`))
	if !diags.HasError() {
		t.Fatal("expected non-object JSON to be rejected")
	}
}

func TestMergeScheduleConfigPreservesTargetAndRemovesSuspensionAction(t *testing.T) {
	t.Parallel()

	prior := jsontypes.NewNormalizedValue(`{"target":[{"type":"email","recipient":"user@example.com"}],"schedule":{"cronSpec":"0 9 * * 1"},"configV2":{"title":"Weekly"},"suspensionAction":"pause"}`)
	merged, diags := mergeScheduleConfig(prior, scheduleRefresh{
		Schedule: json.RawMessage(`{"cronSpec":"0 10 * * 1"}`),
		ConfigV2: json.RawMessage(`{"title":"Daily"}`),
	})
	if diags.HasError() {
		t.Fatalf("merge diagnostics: %v", diags)
	}
	var object map[string]any
	if err := json.Unmarshal([]byte(merged.ValueString()), &object); err != nil {
		t.Fatal(err)
	}
	if _, ok := object["target"]; !ok {
		t.Fatal("expected target to be retained")
	}
	if _, ok := object["suspensionAction"]; ok {
		t.Fatal("expected suspensionAction to be removed")
	}
	schedule, _ := object["schedule"].(map[string]any)
	if schedule["cronSpec"] != "0 10 * * 1" {
		t.Fatalf("schedule = %#v", object["schedule"])
	}
}

func TestScheduleUpdateBodySendsPauseOnlyWhenChanged(t *testing.T) {
	t.Parallel()

	config := jsontypes.NewNormalizedValue(`{"schedule":{"cronSpec":"0 9 * * 1"},"target":[]}`)
	body, diags := scheduleUpdateBody(config, types.BoolValue(true), types.BoolValue(false))
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}
	var object map[string]any
	if err := json.Unmarshal(body, &object); err != nil {
		t.Fatal(err)
	}
	if object["suspensionAction"] != "pause" {
		t.Fatalf("suspensionAction = %#v", object["suspensionAction"])
	}

	unchanged, diags := scheduleUpdateBody(config, types.BoolValue(false), types.BoolValue(false))
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}
	object = map[string]any{}
	if err := json.Unmarshal(unchanged, &object); err != nil {
		t.Fatal(err)
	}
	if _, ok := object["suspensionAction"]; ok {
		t.Fatal("expected unchanged is_suspended to omit suspensionAction")
	}
}

func TestCanonicalJSONIgnoresWhitespaceAndKeyOrder(t *testing.T) {
	t.Parallel()

	left, err := canonicalJSON([]byte(`{ "b": 1, "a": 2 }`))
	if err != nil {
		t.Fatal(err)
	}
	right, err := canonicalJSON([]byte(`{"a":2,"b":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if left != right {
		t.Fatalf("canonical JSON mismatch: %s vs %s", left, right)
	}
}
