package provider

import (
	"bytes"
	"context"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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

func TestConfiguredValueAndAbortUnknown(t *testing.T) {
	var diags diag.Diagnostics
	if abortUnknownInputs(&diags, types.StringUnknown()) != true || !diags.HasError() {
		t.Fatal("unknown input should abort")
	}
	diags = nil
	if abortUnknownInputs(&diags, types.StringValue("ok")) {
		t.Fatal("known input should not abort")
	}

	t.Setenv("SIGMA_TEST_ATTR", "from-env")
	diags = nil
	if got := configuredValue(types.StringNull(), "SIGMA_TEST_ATTR", "test_attr", &diags); got != "from-env" || diags.HasError() {
		t.Fatalf("env value = %q diags=%v", got, diags)
	}
	diags = nil
	if got := configuredValue(types.StringValue("from-config"), "SIGMA_TEST_ATTR", "test_attr", &diags); got != "from-config" {
		t.Fatalf("config value = %q", got)
	}
	diags = nil
	if got := configuredValue(types.StringUnknown(), "SIGMA_TEST_ATTR", "test_attr", &diags); got != "" || !diags.HasError() {
		t.Fatalf("unknown = %q diags=%v", got, diags)
	}
	t.Setenv("SIGMA_TEST_ATTR", "")
	diags = nil
	if got := configuredValue(types.StringNull(), "SIGMA_TEST_ATTR", "test_attr", &diags); got != "" || !diags.HasError() {
		t.Fatalf("missing = %q diags=%v", got, diags)
	}
}

func TestJSONHelpers(t *testing.T) {
	t.Parallel()
	_, _, diags := knownNormalizedObject(jsontypes.NewNormalizedNull(), "details_json")
	if !diags.HasError() {
		t.Fatal("null object")
	}
	_, _, diags = knownNormalizedObject(jsontypes.NewNormalizedValue(`[]`), "details_json")
	if !diags.HasError() {
		t.Fatal("array object")
	}
	object, raw, diags := knownNormalizedObject(jsontypes.NewNormalizedValue(`{"b":1,"a":2}`), "details_json")
	if diags.HasError() || object["a"].(float64) != 2 || len(raw) == 0 {
		t.Fatalf("object = %#v diags=%v", object, diags)
	}
	merged, diags := mergeObjectWithWriteOnly(object, types.StringValue(`{"secret":"x"}`), "credentials_wo")
	if diags.HasError() || !bytes.Contains(merged, []byte("secret")) {
		t.Fatalf("merged = %s diags=%v", merged, diags)
	}
	_, diags = mergeObjectWithWriteOnly(object, types.StringValue("{"), "credentials_wo")
	if !diags.HasError() {
		t.Fatal("invalid overlay")
	}
	diags = writeOnlyVersionPair(types.StringValue("secret"), types.Int64Null(), "credentials_wo", "credentials_wo_version")
	if !diags.HasError() {
		t.Fatal("unpaired payload")
	}
	diags = writeOnlyVersionPair(types.StringNull(), types.Int64Value(1), "credentials_wo", "credentials_wo_version")
	if !diags.HasError() {
		t.Fatal("unpaired version")
	}
	if diags = writeOnlyVersionPair(types.StringValue("secret"), types.Int64Value(1), "credentials_wo", "credentials_wo_version"); diags.HasError() {
		t.Fatalf("paired: %v", diags)
	}
	body, err := suspensionMismatchBody(true)
	if err != nil || !bytes.Contains(body, []byte("pause")) {
		t.Fatalf("pause body = %s err=%v", body, err)
	}
	body, err = suspensionMismatchBody(false)
	if err != nil || !bytes.Contains(body, []byte("resume")) {
		t.Fatalf("resume body = %s err=%v", body, err)
	}
	if _, diags := knownString(types.StringUnknown(), "id"); !diags.HasError() {
		t.Fatal("unknown string")
	}
	if _, diags := knownString(types.StringNull(), "id"); !diags.HasError() {
		t.Fatal("null string")
	}
	if got, diags := knownString(types.StringValue("id-1"), "id"); diags.HasError() || got != "id-1" {
		t.Fatalf("known string = %q %v", got, diags)
	}
	encoded, diags := optionalNormalizedJSON(jsontypes.NewNormalizedNull(), "swaps_json")
	if diags.HasError() || encoded != nil {
		t.Fatalf("optional null = %s %v", encoded, diags)
	}
	if got := normalizedFromRaw(nil); !got.IsNull() {
		t.Fatal("empty raw should be null")
	}
	if got := stringOrNull(nil); !got.IsNull() {
		t.Fatal("nil optional string should be null")
	}
	empty := ""
	if got := stringOrNull(&empty); !got.IsNull() {
		t.Fatal("empty optional string should be null")
	}
	set, diags := stringSetValue(context.Background(), nil)
	if diags.HasError() || set.IsNull() {
		t.Fatalf("empty set = %#v %v", set, diags)
	}
	_, diags = knownStringMap(context.Background(), types.MapNull(types.StringType), "headers")
	if !diags.HasError() {
		t.Fatal("null map")
	}
}

func TestConfigureRejectsUnexpectedProviderData(t *testing.T) {
	t.Parallel()
	resourceResp := &resource.ConfigureResponse{}
	(&configuredResource{}).configure(resource.ConfigureRequest{ProviderData: "nope"}, resourceResp)
	if !resourceResp.Diagnostics.HasError() {
		t.Fatal("resource configure")
	}
	dataResp := &datasource.ConfigureResponse{}
	(&configuredDataSource{}).configure(datasource.ConfigureRequest{ProviderData: "nope"}, dataResp)
	if !dataResp.Diagnostics.HasError() {
		t.Fatal("data source configure")
	}
	whoamiResp := &datasource.ReadResponse{}
	(&whoamiDataSource{}).Read(context.Background(), datasource.ReadRequest{}, whoamiResp)
	if !whoamiResp.Diagnostics.HasError() {
		t.Fatal("whoami without client")
	}
	whoamiCfg := &datasource.ConfigureResponse{}
	(&whoamiDataSource{}).Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "nope"}, whoamiCfg)
	if !whoamiCfg.Diagnostics.HasError() {
		t.Fatal("whoami configure")
	}
	embedResp := &resource.ImportStateResponse{}
	(&workbookEmbedResource{}).ImportState(context.Background(), resource.ImportStateRequest{ID: "bad"}, embedResp)
	if !embedResp.Diagnostics.HasError() {
		t.Fatal("embed import")
	}
	_, diags := optionalStringSlice(context.Background(), types.SetUnknown(types.StringType), "type_filters")
	if !diags.HasError() {
		t.Fatal("unknown set")
	}
	got, diags := optionalStringSlice(context.Background(), types.SetNull(types.StringType), "type_filters")
	if diags.HasError() || got != nil {
		t.Fatalf("null set = %#v %v", got, diags)
	}
	emptySet, setDiags := types.SetValue(types.StringType, []attr.Value{})
	if setDiags.HasError() {
		t.Fatal(setDiags)
	}
	got, diags = optionalStringSlice(context.Background(), emptySet, "type_filters")
	if diags.HasError() || got == nil || len(*got) != 0 {
		t.Fatalf("empty set = %#v %v", got, diags)
	}
	whoamiNil := &datasource.ConfigureResponse{}
	(&whoamiDataSource{}).Configure(context.Background(), datasource.ConfigureRequest{}, whoamiNil)
	if whoamiNil.Diagnostics.HasError() {
		t.Fatal("nil provider data should be a no-op")
	}
	if contains(nil, "x") || contains([]string{"a"}, "x") || !contains([]string{"x"}, "x") {
		t.Fatal("contains")
	}
	for _, importer := range []func(context.Context, resource.ImportStateRequest, *resource.ImportStateResponse){
		(&deploymentPolicyDocumentResource{}).ImportState,
		(&deploymentPolicyTenantResource{}).ImportState,
		(&tenantDeploymentCapabilityResource{}).ImportState,
	} {
		resp := &resource.ImportStateResponse{}
		importer(context.Background(), resource.ImportStateRequest{ID: "bad"}, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected invalid composite import")
		}
	}
}
