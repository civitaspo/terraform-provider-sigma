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
	member := sigma.Member{MemberID: "member-1", OrganizationID: "org-1", Email: "ada@example.com", FirstName: "Ada", LastName: "Lovelace", MemberType: "Creator", UserKind: "internal", HomeFolderID: "folder-1"}
	teamDescription := "Analytics"
	workspaceID := "workspace-1"
	team := sigma.Team{TeamID: "team-1", Name: "Analytics", Description: &teamDescription, Visibility: "private", WorkspaceID: &workspaceID}
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
		var body sigma.CreateTeamInput
		_ = json.NewDecoder(request.Body).Decode(&body)
		if body.Name != team.Name {
			t.Errorf("create team name = %q", body.Name)
		}
		if body.CreateTeamFolder == nil || !*body.CreateTeamFolder {
			t.Errorf("createTeamFolder = %#v", body.CreateTeamFolder)
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
	gotMember, err := client.GetMember(ctx, member.MemberID)
	if err != nil {
		t.Fatal(err)
	}
	if gotMember.HomeFolderID != "folder-1" {
		t.Fatalf("GetMember homeFolderId = %q", gotMember.HomeFolderID)
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

	createFolder := true
	created, err := client.CreateTeam(ctx, sigma.CreateTeamInput{Name: team.Name, CreateTeamFolder: &createFolder})
	if err != nil {
		t.Fatal(err)
	}
	if created.WorkspaceID == nil || *created.WorkspaceID != "workspace-1" {
		t.Fatalf("CreateTeam workspaceId = %v", created.WorkspaceID)
	}
	got, err := client.GetTeam(ctx, team.TeamID)
	if err != nil {
		t.Fatal(err)
	}
	if got.WorkspaceID == nil || *got.WorkspaceID != "workspace-1" {
		t.Fatalf("GetTeam workspaceId = %v", got.WorkspaceID)
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

func TestCreateTeamInputJSON(t *testing.T) {
	t.Parallel()
	trueVal := true
	encoded, err := json.Marshal(sigma.CreateTeamInput{Name: "Analytics", CreateTeamFolder: &trueVal})
	if err != nil {
		t.Fatal(err)
	}
	var withFolder map[string]any
	if err := json.Unmarshal(encoded, &withFolder); err != nil {
		t.Fatal(err)
	}
	if withFolder["createTeamFolder"] != true {
		t.Fatalf("createTeamFolder = %#v", withFolder["createTeamFolder"])
	}

	falseVal := false
	encoded, err = json.Marshal(sigma.CreateTeamInput{Name: "Analytics", CreateTeamFolder: &falseVal})
	if err != nil {
		t.Fatal(err)
	}
	var withFalse map[string]any
	if err := json.Unmarshal(encoded, &withFalse); err != nil {
		t.Fatal(err)
	}
	if withFalse["createTeamFolder"] != false {
		t.Fatalf("createTeamFolder false = %#v", withFalse["createTeamFolder"])
	}

	encoded, err = json.Marshal(sigma.CreateTeamInput{Name: "Analytics"})
	if err != nil {
		t.Fatal(err)
	}
	var omitted map[string]any
	if err := json.Unmarshal(encoded, &omitted); err != nil {
		t.Fatal(err)
	}
	if _, ok := omitted["createTeamFolder"]; ok {
		t.Fatalf("createTeamFolder unexpectedly present: %#v", omitted)
	}
}

func TestTeamWorkspaceIDJSON(t *testing.T) {
	t.Parallel()
	var team sigma.Team
	if err := json.Unmarshal([]byte(`{"teamId":"team-1","workspaceId":"workspace-1"}`), &team); err != nil {
		t.Fatal(err)
	}
	if team.WorkspaceID == nil || *team.WorkspaceID != "workspace-1" {
		t.Fatalf("workspaceId = %v", team.WorkspaceID)
	}
	var nullTeam sigma.Team
	if err := json.Unmarshal([]byte(`{"teamId":"team-1","workspaceId":null}`), &nullTeam); err != nil {
		t.Fatal(err)
	}
	if nullTeam.WorkspaceID != nil {
		t.Fatalf("null workspaceId = %v", nullTeam.WorkspaceID)
	}
}

func TestCreateMemberInputJSON(t *testing.T) {
	t.Parallel()
	admin := true
	encoded, err := json.Marshal(sigma.CreateMemberInput{
		Email: "ada@example.com", FirstName: "Ada", LastName: "Lovelace",
		AddToTeams: []sigma.AddToTeamInput{{TeamID: "team-1", IsTeamAdmin: &admin}},
		SendInvite: &admin,
	})
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(encoded, &body); err != nil {
		t.Fatal(err)
	}
	if _, ok := body["sendInvite"]; ok {
		t.Fatalf("sendInvite should be a query parameter, not JSON: %#v", body)
	}
	teams, _ := body["addToTeams"].([]any)
	if len(teams) != 1 {
		t.Fatalf("addToTeams = %#v", body["addToTeams"])
	}
	team, _ := teams[0].(map[string]any)
	if team["teamId"] != "team-1" || team["isTeamAdmin"] != true {
		t.Fatalf("addToTeams[0] = %#v", team)
	}

	encoded, err = json.Marshal(sigma.CreateMemberInput{Email: "ada@example.com", FirstName: "Ada", LastName: "Lovelace"})
	if err != nil {
		t.Fatal(err)
	}
	var omitted map[string]any
	if err := json.Unmarshal(encoded, &omitted); err != nil {
		t.Fatal(err)
	}
	if _, ok := omitted["addToTeams"]; ok {
		t.Fatalf("addToTeams unexpectedly present: %#v", omitted)
	}
}

func TestCreateMemberSendInviteQuery(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	var gotQuery string
	var gotBody map[string]any
	mock.Mux.HandleFunc("/v2/members", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			t.Errorf("method = %s", request.Method)
		}
		gotQuery = request.URL.Query().Get("sendInvite")
		_ = json.NewDecoder(request.Body).Decode(&gotBody)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(sigma.Member{MemberID: "member-1", Email: "ada@example.com", HomeFolderID: "folder-1"})
	})
	client, err := sigma.NewClient(mock.URL(), mock.ClientID, mock.ClientSecret)
	if err != nil {
		t.Fatal(err)
	}
	sendInvite := false
	admin := true
	created, err := client.CreateMember(context.Background(), sigma.CreateMemberInput{
		Email: "ada@example.com", FirstName: "Ada", LastName: "Lovelace",
		SendInvite: &sendInvite,
		AddToTeams: []sigma.AddToTeamInput{{TeamID: "team-1", IsTeamAdmin: &admin}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.HomeFolderID != "folder-1" {
		t.Fatalf("homeFolderId = %q", created.HomeFolderID)
	}
	if gotQuery != "false" {
		t.Fatalf("sendInvite query = %q", gotQuery)
	}
	if gotBody["email"] != "ada@example.com" {
		t.Fatalf("body = %#v", gotBody)
	}
	teams, _ := gotBody["addToTeams"].([]any)
	if len(teams) != 1 {
		t.Fatalf("addToTeams = %#v", gotBody["addToTeams"])
	}
}

func TestUpdateMemberDeactivateJSON(t *testing.T) {
	t.Parallel()
	archived := true
	owner := "member-admin"
	archiveDocs := true
	encoded, err := json.Marshal(sigma.UpdateMemberInput{
		IsArchived: &archived, NewOwnerID: &owner, ArchiveDocuments: &archiveDocs,
	})
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(encoded, &body); err != nil {
		t.Fatal(err)
	}
	if body["isArchived"] != true || body["newOwnerId"] != "member-admin" || body["archiveDocuments"] != true {
		t.Fatalf("deactivate PATCH = %#v", body)
	}
	if _, ok := body["archiveScheduledExports"]; ok {
		t.Fatalf("archiveScheduledExports unexpectedly present: %#v", body)
	}
}
