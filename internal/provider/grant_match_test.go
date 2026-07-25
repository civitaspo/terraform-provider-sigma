package provider

import (
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestGrantMatchesConsidersTagID(t *testing.T) {
	t.Parallel()

	member := "member-1"
	tag := "tag-1"
	grant := &sigma.Grant{
		Permission: "explore",
		MemberID:   &member,
		TagID:      &tag,
	}
	model := &grantModel{
		Permission: types.StringValue("explore"),
		MemberID:   types.StringValue("member-1"),
		TagID:      types.StringValue("tag-1"),
	}
	if !grantMatches(grant, model) {
		t.Fatal("expected tagged grant to match")
	}

	model.TagID = types.StringValue("tag-2")
	if grantMatches(grant, model) {
		t.Fatal("expected mismatched tag_id to reject grant")
	}

	untagged := &sigma.Grant{Permission: "explore", MemberID: &member}
	model.TagID = types.StringValue("tag-1")
	if grantMatches(untagged, model) {
		t.Fatal("expected API grant without tagId not to match desired tagged grant")
	}

	model.TagID = types.StringNull()
	if !grantMatches(untagged, model) {
		t.Fatal("expected untagged grant to match null tag_id")
	}
}
