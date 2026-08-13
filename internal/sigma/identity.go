package sigma

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Member struct {
	OrganizationID string `json:"organizationId"`
	MemberID       string `json:"memberId"`
	MemberType     string `json:"memberType"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	Email          string `json:"email"`
	UserKind       string `json:"userKind"`
	HomeFolderID   string `json:"homeFolderId"`
	IsArchived     bool   `json:"isArchived"`
	IsInactive     bool   `json:"isInactive"`
}
type AddToTeamInput struct {
	TeamID      string `json:"teamId"`
	IsTeamAdmin *bool  `json:"isTeamAdmin,omitempty"`
}
type CreateMemberInput struct {
	Email      string           `json:"email"`
	FirstName  string           `json:"firstName"`
	LastName   string           `json:"lastName"`
	MemberType string           `json:"memberType,omitempty"`
	UserKind   string           `json:"userKind,omitempty"`
	AddToTeams []AddToTeamInput `json:"addToTeams,omitempty"`
	SendInvite *bool            `json:"-"`
}
type UpdateMemberInput struct {
	FirstName               *string `json:"firstName,omitempty"`
	LastName                *string `json:"lastName,omitempty"`
	Email                   *string `json:"email,omitempty"`
	MemberType              *string `json:"memberType,omitempty"`
	UserKind                *string `json:"userKind,omitempty"`
	IsArchived              *bool   `json:"isArchived,omitempty"`
	NewOwnerID              *string `json:"newOwnerId,omitempty"`
	ArchiveDocuments        *bool   `json:"archiveDocuments,omitempty"`
	ArchiveScheduledExports *bool   `json:"archiveScheduledExports,omitempty"`
}

func (c *Client) CreateMember(ctx context.Context, in CreateMemberInput) (*Member, error) {
	path := "/v2/members"
	if in.SendInvite != nil {
		path += "?sendInvite=" + strconv.FormatBool(*in.SendInvite)
	}
	var v Member
	err := c.sendJSON(ctx, http.MethodPost, path, in, &v)
	return &v, err
}
func (c *Client) GetMember(ctx context.Context, id string) (*Member, error) {
	var v Member
	err := c.getJSON(ctx, "/v2/members/"+url.PathEscape(id), &v)
	return &v, err
}
func (c *Client) UpdateMember(ctx context.Context, id string, in UpdateMemberInput) (*Member, error) {
	var v Member
	err := c.sendJSON(ctx, http.MethodPatch, "/v2/members/"+url.PathEscape(id), in, &v)
	return &v, err
}
func (c *Client) DeleteMember(ctx context.Context, id string) error {
	return c.sendJSON(ctx, http.MethodDelete, "/v2/members/"+url.PathEscape(id), nil, nil)
}
func (c *Client) ListMembers(ctx context.Context) ([]Member, error) {
	return ListAll[Member](ctx, c, "/v2/members")
}

// FindMemberByEmail looks up a member by email. When includeArchived is true, archived
// members are included so callers can reactivate deactivated accounts.
func (c *Client) FindMemberByEmail(ctx context.Context, email string, includeArchived bool) (*Member, error) {
	path := "/v2/members?email=" + url.QueryEscape(email)
	if includeArchived {
		path += "&includeArchived=true"
	}
	members, err := ListAll[Member](ctx, c, path)
	if err != nil {
		return nil, err
	}
	for i := range members {
		if strings.EqualFold(members[i].Email, email) {
			return &members[i], nil
		}
	}
	return nil, &APIError{StatusCode: http.StatusNotFound, Message: "member not found"}
}

type Team struct {
	TeamID      string   `json:"teamId"`
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	Visibility  string   `json:"visibility"`
	IsArchived  bool     `json:"isArchived"`
	Members     []string `json:"members"`
	WorkspaceID *string  `json:"workspaceId"`
}
type CreateTeamInput struct {
	Name             string   `json:"name"`
	Description      string   `json:"description,omitempty"`
	Visibility       string   `json:"visibility,omitempty"`
	Members          []string `json:"members,omitempty"`
	CreateTeamFolder *bool    `json:"createTeamFolder,omitempty"`
}
type UpdateTeamInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Visibility  *string `json:"visibility,omitempty"`
}

func (c *Client) CreateTeam(ctx context.Context, in CreateTeamInput) (*Team, error) {
	var v Team
	err := c.sendJSON(ctx, http.MethodPost, "/v2/teams", in, &v)
	return &v, err
}
func (c *Client) GetTeam(ctx context.Context, id string) (*Team, error) {
	var v Team
	err := c.getJSON(ctx, "/v2/teams/"+url.PathEscape(id), &v)
	return &v, err
}
func (c *Client) UpdateTeam(ctx context.Context, id string, in UpdateTeamInput) (*Team, error) {
	var v Team
	err := c.sendJSON(ctx, http.MethodPatch, "/v2/teams/"+url.PathEscape(id), in, &v)
	return &v, err
}
func (c *Client) DeleteTeam(ctx context.Context, id string) error {
	return c.sendJSON(ctx, http.MethodDelete, "/v2/teams/"+url.PathEscape(id), nil, nil)
}
func (c *Client) ListTeams(ctx context.Context) ([]Team, error) {
	return ListAll[Team](ctx, c, "/v2.1/teams")
}

type TeamMember struct {
	UserID      string `json:"userId"`
	IsTeamAdmin bool   `json:"isTeamAdmin"`
	AddedBy     string `json:"addedBy"`
}

func (c *Client) ListTeamMembers(ctx context.Context, id string) ([]TeamMember, error) {
	return ListAll[TeamMember](ctx, c, "/v2/teams/"+url.PathEscape(id)+"/members")
}
func (c *Client) UpdateTeamMembers(ctx context.Context, id string, add, remove []string) error {
	return c.sendJSON(ctx, http.MethodPatch, "/v2/teams/"+url.PathEscape(id)+"/members", map[string]any{"add": add, "remove": remove}, nil)
}

type AccountType struct {
	AccountTypeID   string `json:"accountTypeId"`
	AccountTypeName string `json:"accountTypeName"`
	Description     string `json:"description"`
	IsCustom        bool   `json:"isCustom"`
}
type AccountTypePermission struct {
	Permission  string `json:"permission"`
	Description string `json:"description"`
}

func (c *Client) ListAccountTypes(ctx context.Context) ([]AccountType, error) {
	var all []AccountType
	p := "/v2/accountTypes"
	seen := map[string]struct{}{}
	for pageNum := 0; pageNum < maxListPages; pageNum++ {
		var page struct {
			Entries []AccountType `json:"entries"`
			Next    string        `json:"nextPageToken"`
		}
		if err := c.getJSON(ctx, p, &page); err != nil {
			return nil, err
		}
		all = append(all, page.Entries...)
		if page.Next == "" {
			return all, nil
		}
		if _, ok := seen[page.Next]; ok {
			return nil, fmt.Errorf("sigma pagination cycle detected: nextPageToken %q repeated", page.Next)
		}
		seen[page.Next] = struct{}{}
		p = "/v2/accountTypes?pageToken=" + url.QueryEscape(page.Next)
	}
	return nil, fmt.Errorf("sigma pagination exceeded %d pages for /v2/accountTypes", maxListPages)
}
func (c *Client) CreateAccountType(ctx context.Context, n, d string, p []string) (*AccountType, error) {
	var v AccountType
	err := c.sendJSON(ctx, http.MethodPost, "/v2/accountTypes", map[string]any{"name": n, "description": d, "permissions": p}, &v)
	return &v, err
}
func (c *Client) DeleteAccountType(ctx context.Context, id, reassign string) error {
	p := "/v2/accountTypes/" + url.PathEscape(id)
	if reassign != "" {
		p += "?reassignToAccountTypeId=" + url.QueryEscape(reassign)
	}
	return c.sendJSON(ctx, http.MethodDelete, p, nil, nil)
}
func (c *Client) ListAccountTypePermissions(ctx context.Context, id string) ([]AccountTypePermission, error) {
	var v []AccountTypePermission
	err := c.getJSON(ctx, "/v2/accountTypes/"+url.PathEscape(id)+"/permissions", &v)
	return v, err
}
func (c *Client) FindAccountType(ctx context.Context, key string) (*AccountType, error) {
	v, err := c.ListAccountTypes(ctx)
	if err != nil {
		return nil, err
	}
	for _, x := range v {
		if x.AccountTypeID == key || x.AccountTypeName == key {
			return &x, nil
		}
	}
	return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("account type %q not found", key)}
}

type AttributeValue struct {
	Val  string `json:"val"`
	Type string `json:"type"`
}
type UserAttribute struct {
	UserAttributeID string          `json:"userAttributeId"`
	Name            string          `json:"name"`
	Description     *string         `json:"description"`
	DefaultValue    *AttributeValue `json:"defaultValue"`
}

func (c *Client) CreateUserAttribute(ctx context.Context, n, d string, def *AttributeValue) (*UserAttribute, error) {
	in := map[string]any{"name": n}
	if d != "" {
		in["description"] = d
	}
	if def != nil {
		in["defaultValue"] = def
	}
	var v UserAttribute
	err := c.sendJSON(ctx, http.MethodPost, "/v2/user-attributes", in, &v)
	return &v, err
}
func (c *Client) GetUserAttribute(ctx context.Context, id string) (*UserAttribute, error) {
	var v UserAttribute
	err := c.getJSON(ctx, "/v2/user-attributes/"+url.PathEscape(id), &v)
	return &v, err
}
func (c *Client) DeleteUserAttribute(ctx context.Context, id string) error {
	return c.sendJSON(ctx, http.MethodDelete, "/v2/user-attributes/"+url.PathEscape(id), nil, nil)
}
func (c *Client) ListUserAttributes(ctx context.Context) ([]UserAttribute, error) {
	return ListAll[UserAttribute](ctx, c, "/v2/user-attributes")
}

type AttributeAssignment struct {
	TeamID               string         `json:"teamId,omitempty"`
	UserID               string         `json:"userId,omitempty"`
	TenantOrganizationID string         `json:"tenantOrganizationId,omitempty"`
	Value                AttributeValue `json:"value"`
}

func (c *Client) SetUserAttributeTeam(ctx context.Context, a, t, v string) error {
	return c.sendJSON(ctx, http.MethodPost, "/v2/user-attributes/"+url.PathEscape(a)+"/teams", map[string]any{"assignments": []AttributeAssignment{{TeamID: t, Value: AttributeValue{Val: v, Type: "string"}}}}, nil)
}
func (c *Client) DeleteUserAttributeTeam(ctx context.Context, a, t string) error {
	return c.sendJSON(ctx, http.MethodDelete, "/v2/user-attributes/"+url.PathEscape(a)+"/teams/"+url.PathEscape(t), nil, nil)
}
func (c *Client) ListUserAttributeTeams(ctx context.Context, a string) ([]AttributeAssignment, error) {
	return ListAll[AttributeAssignment](ctx, c, "/v2/user-attributes/"+url.PathEscape(a)+"/teams")
}
func (c *Client) SetUserAttributeUser(ctx context.Context, a, u, v string) error {
	return c.sendJSON(ctx, http.MethodPost, "/v2/user-attributes/"+url.PathEscape(a)+"/users", map[string]any{"assignments": []AttributeAssignment{{UserID: u, Value: AttributeValue{Val: v, Type: "string"}}}}, nil)
}
func (c *Client) DeleteUserAttributeUser(ctx context.Context, a, u string) error {
	return c.sendJSON(ctx, http.MethodDelete, "/v2/user-attributes/"+url.PathEscape(a)+"/users/"+url.PathEscape(u), nil, nil)
}
func (c *Client) ListUserAttributeUsers(ctx context.Context, a string) ([]AttributeAssignment, error) {
	return ListAll[AttributeAssignment](ctx, c, "/v2/user-attributes/"+url.PathEscape(a)+"/users")
}
func (c *Client) SetUserAttributeTenant(ctx context.Context, a, t, v string) error {
	return c.sendJSON(ctx, http.MethodPost, "/v2/user-attributes/"+url.PathEscape(a)+"/tenants", map[string]any{"assignments": []AttributeAssignment{{TenantOrganizationID: t, Value: AttributeValue{Val: v, Type: "string"}}}}, nil)
}
func (c *Client) DeleteUserAttributeTenant(ctx context.Context, a, t string) error {
	return c.sendJSON(ctx, http.MethodDelete, "/v2/user-attributes/"+url.PathEscape(a)+"/tenants/"+url.PathEscape(t), nil, nil)
}
func (c *Client) ListUserAttributeTenants(ctx context.Context, a string) ([]AttributeAssignment, error) {
	return ListAll[AttributeAssignment](ctx, c, "/v2/user-attributes/"+url.PathEscape(a)+"/tenants")
}
