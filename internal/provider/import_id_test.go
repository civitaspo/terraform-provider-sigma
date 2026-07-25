package provider

import "testing"

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
