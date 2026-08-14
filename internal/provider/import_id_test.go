package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestSplitCompositeImportID(t *testing.T) {
	t.Parallel()

	left, right, ok := splitCompositeImportID("team-1/member-1")
	if !ok || left != "team-1" || right != "member-1" {
		t.Fatalf("got (%q, %q, %v)", left, right, ok)
	}

	for _, id := range []string{"", "only", "/member-1", "team-1/", "/"} {
		if _, _, ok := splitCompositeImportID(id); ok {
			t.Fatalf("splitCompositeImportID(%q) = true, want false", id)
		}
	}
}

func TestImportGrantCompositeIDRejectsInvalid(t *testing.T) {
	t.Parallel()
	var response resource.ImportStateResponse
	importGrantCompositeID(context.Background(), resource.ImportStateRequest{ID: "only"}, &response)
	if !response.Diagnostics.HasError() {
		t.Fatal("expected invalid import ID diagnostics")
	}
}
