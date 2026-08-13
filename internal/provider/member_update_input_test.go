package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMemberUpdateInputSendsChangedFieldsOnly(t *testing.T) {
	t.Parallel()

	state := &memberModel{
		FirstName:  types.StringValue("Ada"),
		LastName:   types.StringValue("Lovelace"),
		Email:      types.StringValue("ada@example.com"),
		MemberType: types.StringValue("Creator"),
		UserKind:   types.StringValue("internal"),
		IsArchived: types.BoolValue(true),
	}
	plan := &memberModel{
		FirstName:  types.StringValue("Augusta"),
		LastName:   types.StringValue("Lovelace"),
		Email:      types.StringValue("ada@example.com"),
		MemberType: types.StringValue("Explorer"),
		UserKind:   types.StringNull(),
		IsArchived: types.BoolValue(false),
	}
	input := memberUpdateInput(plan, state)
	if input.FirstName == nil || *input.FirstName != "Augusta" {
		t.Fatalf("FirstName = %#v", input.FirstName)
	}
	if input.LastName != nil {
		t.Fatalf("LastName = %#v, want omitted", input.LastName)
	}
	if input.Email != nil {
		t.Fatalf("Email = %#v, want omitted", input.Email)
	}
	if input.MemberType == nil || *input.MemberType != "Explorer" {
		t.Fatalf("MemberType = %#v", input.MemberType)
	}
	if input.UserKind != nil {
		t.Fatalf("UserKind = %#v, want omitted", input.UserKind)
	}
	if input.IsArchived == nil || *input.IsArchived {
		t.Fatalf("IsArchived = %#v, want false", input.IsArchived)
	}
}

func TestMemberUpdateInputOmitsUnchangedArchiveFlag(t *testing.T) {
	t.Parallel()
	state := &memberModel{
		FirstName:  types.StringValue("Ada"),
		LastName:   types.StringValue("Lovelace"),
		Email:      types.StringValue("ada@example.com"),
		IsArchived: types.BoolValue(false),
	}
	plan := &memberModel{
		FirstName:  types.StringValue("Ada"),
		LastName:   types.StringValue("Lovelace"),
		Email:      types.StringValue("ada@example.com"),
		IsArchived: types.BoolValue(false),
	}
	input := memberUpdateInput(plan, state)
	if input.IsArchived != nil {
		t.Fatalf("IsArchived = %#v, want omitted", input.IsArchived)
	}
}
