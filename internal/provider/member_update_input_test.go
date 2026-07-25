package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMemberUpdateInputOmitsNullAccountFields(t *testing.T) {
	t.Parallel()

	input := memberUpdateInput(&memberModel{
		FirstName:  types.StringValue("Ada"),
		LastName:   types.StringValue("Lovelace"),
		Email:      types.StringValue("ada@example.com"),
		MemberType: types.StringNull(),
		UserKind:   types.StringNull(),
	})
	if input.FirstName == nil || *input.FirstName != "Ada" {
		t.Fatalf("FirstName = %#v", input.FirstName)
	}
	if input.MemberType != nil {
		t.Fatalf("MemberType = %#v, want nil", input.MemberType)
	}
	if input.UserKind != nil {
		t.Fatalf("UserKind = %#v, want nil", input.UserKind)
	}

	input = memberUpdateInput(&memberModel{
		FirstName:  types.StringValue("Ada"),
		LastName:   types.StringValue("Lovelace"),
		Email:      types.StringValue("ada@example.com"),
		MemberType: types.StringValue("Creator"),
		UserKind:   types.StringValue("internal"),
	})
	if input.MemberType == nil || *input.MemberType != "Creator" {
		t.Fatalf("MemberType = %#v", input.MemberType)
	}
	if input.UserKind == nil || *input.UserKind != "internal" {
		t.Fatalf("UserKind = %#v", input.UserKind)
	}
}
