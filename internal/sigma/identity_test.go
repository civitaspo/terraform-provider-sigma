package sigma_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
)

func TestIdentityClientMethods(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	write := func(response http.ResponseWriter, value any) {
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(value)
	}
	member := sigma.Member{MemberID: "member-1", OrganizationID: "org-1", Email: "ada@example.com", FirstName: "Ada", LastName: "Lovelace", MemberType: "Creator", UserKind: "internal"}
	teamDescription := "Analytics"
	team := sigma.Team{TeamID: "team-1", Name: "Analytics", Description: &teamDescription, Visibility: "private"}
	attributeDescription := "Region"
	attribute := sigma.UserAttribute{UserAttributeID: "attribute-1", Name: "Region", Description: &attributeDescription, DefaultValue: &sigma.AttributeValue{Val: "global", Type: "string"}}

	mock.Mux.HandleFunc("/v2/members", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method == http.MethodPost {
			write(response, member)
			return
		}
		write(response, map[string]any{"entries": []sigma.Member{member}, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/members/member-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method == http.MethodDelete {
			write(response, map[string]any{})
			return
		}
		write(response, member)
	})
	mock.Mux.HandleFunc("/v2/members/missing", func(response http.ResponseWriter, _ *http.Request) {
		http.Error(response, `{"message":"missing"}`, http.StatusNotFound)
	})
	mock.Mux.HandleFunc("/v2/teams", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Errorf("unexpected teams method %s", request.Method)
		}
		write(response, team)
	})
	mock.Mux.HandleFunc("/v2.1/teams", func(response http.ResponseWriter, _ *http.Request) {
		write(response, map[string]any{"entries": []sigma.Team{team}, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/teams/team-1", func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodDelete {
			write(response, map[string]any{})
			return
		}
		write(response, team)
	})
	mock.Mux.HandleFunc("/v2/teams/team-1/members", func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet {
			write(response, map[string]any{"entries": []sigma.TeamMember{{UserID: "member-1"}}, "nextPage": nil})
			return
		}
		write(response, map[string]any{})
	})
	mock.Mux.HandleFunc("/v2/accountTypes", func(response http.ResponseWriter, request *http.Request) {
		value := sigma.AccountType{AccountTypeID: "type-1", AccountTypeName: "Analyst", Description: "Analyst access", IsCustom: true}
		if request.Method == http.MethodPost {
			response.WriteHeader(http.StatusCreated)
			write(response, value)
			return
		}
		write(response, map[string]any{"entries": []sigma.AccountType{value}})
	})
	mock.Mux.HandleFunc("/v2/accountTypes/type-1", func(response http.ResponseWriter, _ *http.Request) { write(response, map[string]any{}) })
	mock.Mux.HandleFunc("/v2/accountTypes/type-1/permissions", func(response http.ResponseWriter, _ *http.Request) {
		write(response, []sigma.AccountTypePermission{{Permission: "view-worksheet"}})
	})
	mock.Mux.HandleFunc("/v2/user-attributes", func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPost {
			write(response, attribute)
			return
		}
		write(response, map[string]any{"entries": []sigma.UserAttribute{attribute}, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/user-attributes/attribute-1", func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodDelete {
			write(response, map[string]any{})
			return
		}
		write(response, attribute)
	})
	mock.Mux.HandleFunc("/v2/user-attributes/attribute-1/teams", func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet {
			write(response, map[string]any{"entries": []sigma.AttributeAssignment{{TeamID: "team-1", Value: sigma.AttributeValue{Val: "americas", Type: "string"}}}, "nextPage": nil})
			return
		}
		write(response, map[string]any{})
	})
	mock.Mux.HandleFunc("/v2/user-attributes/attribute-1/teams/team-1", func(response http.ResponseWriter, _ *http.Request) { write(response, map[string]any{}) })
	mock.Mux.HandleFunc("/v2/user-attributes/attribute-1/users", func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet {
			write(response, map[string]any{"entries": []sigma.AttributeAssignment{{UserID: "member-1", Value: sigma.AttributeValue{Val: "americas", Type: "string"}}}, "nextPage": nil})
			return
		}
		write(response, map[string]any{})
	})
	mock.Mux.HandleFunc("/v2/user-attributes/attribute-1/users/member-1", func(response http.ResponseWriter, _ *http.Request) { write(response, map[string]any{}) })

	client, err := sigma.NewClient(mock.URL(), mock.ClientID, mock.ClientSecret)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err = client.CreateMember(ctx, sigma.CreateMemberInput{Email: member.Email, FirstName: member.FirstName, LastName: member.LastName}); err != nil {
		t.Fatal(err)
	}
	if _, err = client.GetMember(ctx, member.MemberID); err != nil {
		t.Fatal(err)
	}
	first := "Augusta"
	if _, err = client.UpdateMember(ctx, member.MemberID, sigma.UpdateMemberInput{FirstName: &first}); err != nil {
		t.Fatal(err)
	}
	if values, listErr := client.ListMembers(ctx); listErr != nil || len(values) != 1 {
		t.Fatalf("ListMembers() = %v, %v", values, listErr)
	}
	if err = client.DeleteMember(ctx, member.MemberID); err != nil {
		t.Fatal(err)
	}
	if _, err = client.GetMember(ctx, "missing"); !sigma.IsNotFound(err) {
		t.Fatalf("GetMember(missing) error = %v", err)
	}

	if _, err = client.CreateTeam(ctx, sigma.CreateTeamInput{Name: team.Name}); err != nil {
		t.Fatal(err)
	}
	if _, err = client.GetTeam(ctx, team.TeamID); err != nil {
		t.Fatal(err)
	}
	if _, err = client.UpdateTeam(ctx, team.TeamID, sigma.UpdateTeamInput{Name: &team.Name}); err != nil {
		t.Fatal(err)
	}
	if values, listErr := client.ListTeams(ctx); listErr != nil || len(values) != 1 {
		t.Fatalf("ListTeams() = %v, %v", values, listErr)
	}
	if values, listErr := client.ListTeamMembers(ctx, team.TeamID); listErr != nil || len(values) != 1 {
		t.Fatalf("ListTeamMembers() = %v, %v", values, listErr)
	}
	if err = client.UpdateTeamMembers(ctx, team.TeamID, []string{member.MemberID}, nil); err != nil {
		t.Fatal(err)
	}
	if err = client.DeleteTeam(ctx, team.TeamID); err != nil {
		t.Fatal(err)
	}

	accountType, err := client.CreateAccountType(ctx, "Analyst", "Analyst access", []string{"view-worksheet"})
	if err != nil || accountType.AccountTypeID != "type-1" {
		t.Fatalf("CreateAccountType() = %v, %v", accountType, err)
	}
	if _, err = client.FindAccountType(ctx, "Analyst"); err != nil {
		t.Fatal(err)
	}
	if values, permissionsErr := client.ListAccountTypePermissions(ctx, "type-1"); permissionsErr != nil || len(values) != 1 {
		t.Fatalf("permissions = %v, %v", values, permissionsErr)
	}
	if err = client.DeleteAccountType(ctx, "type-1", "default"); err != nil {
		t.Fatal(err)
	}

	if _, err = client.CreateUserAttribute(ctx, attribute.Name, attributeDescription, attribute.DefaultValue); err != nil {
		t.Fatal(err)
	}
	if _, err = client.GetUserAttribute(ctx, attribute.UserAttributeID); err != nil {
		t.Fatal(err)
	}
	if values, listErr := client.ListUserAttributes(ctx); listErr != nil || len(values) != 1 {
		t.Fatalf("attributes = %v, %v", values, listErr)
	}
	if err = client.SetUserAttributeTeam(ctx, attribute.UserAttributeID, team.TeamID, "americas"); err != nil {
		t.Fatal(err)
	}
	if values, listErr := client.ListUserAttributeTeams(ctx, attribute.UserAttributeID); listErr != nil || len(values) != 1 {
		t.Fatalf("team assignments = %v, %v", values, listErr)
	}
	if err = client.DeleteUserAttributeTeam(ctx, attribute.UserAttributeID, team.TeamID); err != nil {
		t.Fatal(err)
	}
	if err = client.SetUserAttributeUser(ctx, attribute.UserAttributeID, member.MemberID, "americas"); err != nil {
		t.Fatal(err)
	}
	if values, listErr := client.ListUserAttributeUsers(ctx, attribute.UserAttributeID); listErr != nil || len(values) != 1 {
		t.Fatalf("user assignments = %v, %v", values, listErr)
	}
	if err = client.DeleteUserAttributeUser(ctx, attribute.UserAttributeID, member.MemberID); err != nil {
		t.Fatal(err)
	}
	if err = client.DeleteUserAttribute(ctx, attribute.UserAttributeID); err != nil {
		t.Fatal(err)
	}
}
