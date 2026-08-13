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

func TestLookupGrantAmbiguityNeverSelectsFirstMatch(t *testing.T) {
	t.Parallel()

	member := "member-1"
	first := sigma.Grant{GrantID: "grant-1", Permission: "explore", MemberID: &member}
	second := sigma.Grant{GrantID: "grant-2", Permission: "explore", MemberID: &member}
	model := &grantModel{
		Permission: types.StringValue("explore"),
		MemberID:   types.StringValue("member-1"),
	}
	_, err := lookupGrant([]sigma.Grant{first, second}, model, "")
	if err == nil {
		t.Fatal("expected ambiguity error")
	}
	if got := err.Error(); got != "multiple grants matched inode, grantee, and permission; refusing to select the first match" {
		t.Fatalf("error = %q", got)
	}
}

func TestLookupGrantZeroMatchesIsNotFound(t *testing.T) {
	t.Parallel()

	member := "member-1"
	model := &grantModel{
		Permission: types.StringValue("explore"),
		MemberID:   types.StringValue("member-1"),
	}
	_, err := lookupGrant([]sigma.Grant{{GrantID: "grant-1", Permission: "view", MemberID: &member}}, model, "")
	if err == nil {
		t.Fatal("expected not-found")
	}
	if !sigma.IsNotFound(err) {
		t.Fatalf("error = %v, want structured not-found", err)
	}
}
