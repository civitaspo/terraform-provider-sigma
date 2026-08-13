package sigma_test

import (
	"encoding/json"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
)

func TestConnectionUseOauthJSON(t *testing.T) {
	t.Parallel()
	encoded := []byte(`{"connectionId":"connection-1","name":"warehouse","type":"postgres","useOauth":true,"friendlyName":false}`)
	var value sigma.Connection
	if err := json.Unmarshal(encoded, &value); err != nil {
		t.Fatal(err)
	}
	if value.UseOauth == nil || !*value.UseOauth {
		t.Fatalf("useOauth = %#v", value.UseOauth)
	}

	omitted := []byte(`{"connectionId":"connection-1","name":"warehouse","type":"postgres"}`)
	var missing sigma.Connection
	if err := json.Unmarshal(omitted, &missing); err != nil {
		t.Fatal(err)
	}
	if missing.UseOauth != nil {
		t.Fatalf("useOauth unexpectedly present: %#v", missing.UseOauth)
	}
}

func TestConnectionInputRestoreJSON(t *testing.T) {
	t.Parallel()
	restore := true
	encoded, err := json.Marshal(sigma.ConnectionInput{Name: "warehouse", Details: json.RawMessage(`{"type":"postgres"}`), Restore: &restore})
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(encoded, &body); err != nil {
		t.Fatal(err)
	}
	if body["restore"] != true {
		t.Fatalf("restore = %#v", body["restore"])
	}

	encoded, err = json.Marshal(sigma.ConnectionInput{Name: "warehouse", Details: json.RawMessage(`{"type":"postgres"}`)})
	if err != nil {
		t.Fatal(err)
	}
	var omitted map[string]any
	if err := json.Unmarshal(encoded, &omitted); err != nil {
		t.Fatal(err)
	}
	if _, ok := omitted["restore"]; ok {
		t.Fatalf("restore unexpectedly present: %#v", omitted)
	}
}
