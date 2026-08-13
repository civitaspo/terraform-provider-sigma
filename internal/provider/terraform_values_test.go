package provider

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestKnownStringSetRejectsNullUnknownAndNullElements(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, diags := knownStringSet(ctx, types.SetNull(types.StringType), "permissions")
	if !diags.HasError() {
		t.Fatal("null set: expected diagnostics")
	}

	_, diags = knownStringSet(ctx, types.SetUnknown(types.StringType), "permissions")
	if !diags.HasError() {
		t.Fatal("unknown set: expected diagnostics")
	}

	withNull, setDiags := types.SetValue(types.StringType, []attr.Value{types.StringNull()})
	if setDiags.HasError() {
		t.Fatalf("set value diags: %v", setDiags)
	}
	_, diags = knownStringSet(ctx, withNull, "permissions")
	if !diags.HasError() {
		t.Fatal("null element: expected diagnostics")
	}

	withUnknown, setDiags := types.SetValue(types.StringType, []attr.Value{types.StringUnknown()})
	if setDiags.HasError() {
		t.Fatalf("set value diags: %v", setDiags)
	}
	_, diags = knownStringSet(ctx, withUnknown, "permissions")
	if !diags.HasError() {
		t.Fatal("unknown element: expected diagnostics")
	}
}

func TestAccountTypeCreateUnknownPermissionsMakesZeroHTTPCalls(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	var extra atomic.Int64
	mock.Mux.HandleFunc("/v2/accountTypes", func(response http.ResponseWriter, request *http.Request) {
		extra.Add(1)
		http.Error(response, "unexpected "+request.URL.Path, http.StatusInternalServerError)
	})
	client, err := sigma.NewClient(mock.URL(), mock.ClientID, mock.ClientSecret)
	if err != nil {
		t.Fatal(err)
	}
	resource := &accountTypeResource{configuredResource{client: client}}
	plan := accountTypeModel{
		Name:        types.StringValue("Analyst"),
		Description: types.StringValue("Analyst access"),
		Permissions: types.SetUnknown(types.StringType),
	}
	permissions, diags := knownStringSet(context.Background(), plan.Permissions, "permissions")
	if !diags.HasError() {
		t.Fatal("expected conversion diagnostics")
	}
	if permissions != nil {
		t.Fatalf("permissions = %#v, want nil", permissions)
	}
	if extra.Load() != 0 {
		t.Fatalf("non-auth HTTP calls = %d, want 0", extra.Load())
	}
	_ = resource
}

func TestChangedStringPtrOmitsUnchangedAndNull(t *testing.T) {
	t.Parallel()
	plan := types.StringValue("Analytics")
	state := types.StringValue("Analytics")
	if got := changedStringPtr(plan, state); got != nil {
		t.Fatalf("unchanged = %#v", got)
	}
	if got := changedStringPtr(types.StringNull(), types.StringValue("Analytics")); got != nil {
		t.Fatalf("null plan = %#v", got)
	}
	if got := changedStringPtr(types.StringUnknown(), types.StringValue("Analytics")); got != nil {
		t.Fatalf("unknown plan = %#v", got)
	}
	got := changedStringPtr(types.StringValue("Renamed"), types.StringValue("Analytics"))
	if got == nil || *got != "Renamed" {
		t.Fatalf("changed = %#v", got)
	}
}
