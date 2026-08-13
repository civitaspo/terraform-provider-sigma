package sigma

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma/openapi"
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
	body, err := encodeBody(in)
	if err != nil {
		return nil, err
	}
	var v Member
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.CreateMemberWithBody(ctx, &openapi.CreateMemberParams{SendInvite: in.SendInvite}, jsonContentType, body)
	}, &v)
	return &v, err
}
func (c *Client) GetMember(ctx context.Context, id string) (*Member, error) {
	var v Member
	err := c.doDecode(func() (*http.Response, error) {
		return c.api.GetMember(ctx, id, nil)
	}, &v)
	return &v, err
}
func (c *Client) UpdateMember(ctx context.Context, id string, in UpdateMemberInput) (*Member, error) {
	body, err := encodeBody(in)
	if err != nil {
		return nil, err
	}
	var v Member
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.UpdateMemberWithBody(ctx, id, nil, jsonContentType, body)
	}, &v)
	return &v, err
}
func (c *Client) DeleteMember(ctx context.Context, id string) error {
	return c.doVoid(func() (*http.Response, error) {
		return c.api.DeleteMember(ctx, id, nil)
	})
}
func (c *Client) ListMembers(ctx context.Context) ([]Member, error) {
	return c.listMembers(ctx, &openapi.ListMembersParams{})
}

// FindMemberByEmail looks up a member by email. When includeArchived is true, archived
// members are included in the list filter.
func (c *Client) FindMemberByEmail(ctx context.Context, email string, includeArchived bool) (*Member, error) {
	params := &openapi.ListMembersParams{Email: &email}
	if includeArchived {
		include := true
		params.IncludeArchived = &include
	}
	members, err := c.listMembers(ctx, params)
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

func (c *Client) listMembers(ctx context.Context, base *openapi.ListMembersParams) ([]Member, error) {
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]Member, *string, error) {
		params := *base
		params.Page = page
		return fetchPage[Member](c, func() (*http.Response, error) {
			return c.api.ListMembers(ctx, &params)
		})
	})
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
	body, err := encodeBody(in)
	if err != nil {
		return nil, err
	}
	var v Team
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.CreateTeamWithBody(ctx, nil, jsonContentType, body)
	}, &v)
	return &v, err
}
func (c *Client) GetTeam(ctx context.Context, id string) (*Team, error) {
	var v Team
	err := c.doDecode(func() (*http.Response, error) {
		return c.api.GetTeam(ctx, id, nil)
	}, &v)
	return &v, err
}
func (c *Client) UpdateTeam(ctx context.Context, id string, in UpdateTeamInput) (*Team, error) {
	body, err := encodeBody(in)
	if err != nil {
		return nil, err
	}
	var v Team
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.UpdateTeamWithBody(ctx, id, nil, jsonContentType, body)
	}, &v)
	return &v, err
}
func (c *Client) DeleteTeam(ctx context.Context, id string) error {
	return c.doVoid(func() (*http.Response, error) {
		return c.api.DeleteTeam(ctx, id, nil)
	})
}
func (c *Client) ListTeams(ctx context.Context) ([]Team, error) {
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]Team, *string, error) {
		return fetchPage[Team](c, func() (*http.Response, error) {
			return c.api.V21ListTeams(ctx, &openapi.V21ListTeamsParams{Page: page})
		})
	})
}

type TeamMember struct {
	UserID      string `json:"userId"`
	IsTeamAdmin bool   `json:"isTeamAdmin"`
	AddedBy     string `json:"addedBy"`
}

func (c *Client) ListTeamMembers(ctx context.Context, id string) ([]TeamMember, error) {
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]TeamMember, *string, error) {
		return fetchPage[TeamMember](c, func() (*http.Response, error) {
			return c.api.GetTeamMembers(ctx, id, &openapi.GetTeamMembersParams{Page: page})
		})
	})
}
func (c *Client) UpdateTeamMembers(ctx context.Context, id string, add, remove []string) error {
	payload := map[string]any{"add": add, "remove": remove}
	body, err := encodeBody(payload)
	if err != nil {
		return err
	}
	return c.doVoid(func() (*http.Response, error) {
		return c.api.UpdateTeamMembersWithBody(ctx, id, nil, jsonContentType, body)
	})
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
	return listAllByPageToken(ctx, func(ctx context.Context, pageToken *string) ([]AccountType, *string, error) {
		return fetchPageToken[AccountType](c, func() (*http.Response, error) {
			return c.api.ListAccountTypes(ctx, &openapi.ListAccountTypesParams{PageToken: pageToken})
		})
	})
}
func (c *Client) CreateAccountType(ctx context.Context, n, d string, p []string) (*AccountType, error) {
	body, err := encodeBody(map[string]any{"name": n, "description": d, "permissions": p})
	if err != nil {
		return nil, err
	}
	var v AccountType
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.CreateAccountTypeWithBody(ctx, nil, jsonContentType, body)
	}, &v)
	return &v, err
}
func (c *Client) DeleteAccountType(ctx context.Context, id, reassign string) error {
	params := &openapi.DeleteAccountTypeParams{}
	if reassign != "" {
		params.ReassignToAccountTypeId = &reassign
	}
	return c.doVoid(func() (*http.Response, error) {
		return c.api.DeleteAccountType(ctx, id, params)
	})
}
func (c *Client) ListAccountTypePermissions(ctx context.Context, id string) ([]AccountTypePermission, error) {
	var v []AccountTypePermission
	err := c.doDecode(func() (*http.Response, error) {
		return c.api.ListAccountTypePermissions(ctx, id, nil)
	}, &v)
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
	body, err := encodeBody(in)
	if err != nil {
		return nil, err
	}
	var v UserAttribute
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.CreateUserAttributeWithBody(ctx, nil, jsonContentType, body)
	}, &v)
	return &v, err
}
func (c *Client) GetUserAttribute(ctx context.Context, id string) (*UserAttribute, error) {
	var v UserAttribute
	err := c.doDecode(func() (*http.Response, error) {
		return c.api.GetUserAttribute(ctx, id, nil)
	}, &v)
	return &v, err
}
func (c *Client) DeleteUserAttribute(ctx context.Context, id string) error {
	return c.doVoid(func() (*http.Response, error) {
		return c.api.DeleteUserAttribute(ctx, id, nil)
	})
}
func (c *Client) ListUserAttributes(ctx context.Context) ([]UserAttribute, error) {
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]UserAttribute, *string, error) {
		return fetchPage[UserAttribute](c, func() (*http.Response, error) {
			return c.api.ListUserAttributes(ctx, &openapi.ListUserAttributesParams{Page: page})
		})
	})
}

type AttributeAssignment struct {
	TeamID               string         `json:"teamId,omitempty"`
	UserID               string         `json:"userId,omitempty"`
	TenantOrganizationID string         `json:"tenantOrganizationId,omitempty"`
	Value                AttributeValue `json:"value"`
}

func (c *Client) SetUserAttributeTeam(ctx context.Context, a, t, v string) error {
	body, err := encodeBody(map[string]any{"assignments": []AttributeAssignment{{TeamID: t, Value: AttributeValue{Val: v, Type: "string"}}}})
	if err != nil {
		return err
	}
	return c.doVoid(func() (*http.Response, error) {
		return c.api.SetUserAttributeForTeamsWithBody(ctx, a, nil, jsonContentType, body)
	})
}
func (c *Client) DeleteUserAttributeTeam(ctx context.Context, a, t string) error {
	return c.doVoid(func() (*http.Response, error) {
		return c.api.DeleteUserAttributeForTeam(ctx, a, t, nil)
	})
}
func (c *Client) ListUserAttributeTeams(ctx context.Context, a string) ([]AttributeAssignment, error) {
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]AttributeAssignment, *string, error) {
		return fetchPage[AttributeAssignment](c, func() (*http.Response, error) {
			return c.api.GetUserAttributeTeamAssignments(ctx, a, &openapi.GetUserAttributeTeamAssignmentsParams{Page: page})
		})
	})
}
func (c *Client) SetUserAttributeUser(ctx context.Context, a, u, v string) error {
	body, err := encodeBody(map[string]any{"assignments": []AttributeAssignment{{UserID: u, Value: AttributeValue{Val: v, Type: "string"}}}})
	if err != nil {
		return err
	}
	return c.doVoid(func() (*http.Response, error) {
		return c.api.SetUserAttributeForUsersWithBody(ctx, a, nil, jsonContentType, body)
	})
}
func (c *Client) DeleteUserAttributeUser(ctx context.Context, a, u string) error {
	return c.doVoid(func() (*http.Response, error) {
		return c.api.DeleteUserAttributeForUser(ctx, a, u, nil)
	})
}
func (c *Client) ListUserAttributeUsers(ctx context.Context, a string) ([]AttributeAssignment, error) {
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]AttributeAssignment, *string, error) {
		return fetchPage[AttributeAssignment](c, func() (*http.Response, error) {
			return c.api.GetUserAttributeUserAssignments(ctx, a, &openapi.GetUserAttributeUserAssignmentsParams{Page: page})
		})
	})
}
func (c *Client) SetUserAttributeTenant(ctx context.Context, a, t, v string) error {
	body, err := encodeBody(map[string]any{"assignments": []AttributeAssignment{{TenantOrganizationID: t, Value: AttributeValue{Val: v, Type: "string"}}}})
	if err != nil {
		return err
	}
	return c.doVoid(func() (*http.Response, error) {
		return c.api.SetUserAttributeForTenantsWithBody(ctx, a, nil, jsonContentType, body)
	})
}
func (c *Client) DeleteUserAttributeTenant(ctx context.Context, a, t string) error {
	return c.doVoid(func() (*http.Response, error) {
		return c.api.DeleteUserAttributeForTenant(ctx, a, t, nil)
	})
}
func (c *Client) ListUserAttributeTenants(ctx context.Context, a string) ([]AttributeAssignment, error) {
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]AttributeAssignment, *string, error) {
		return fetchPage[AttributeAssignment](c, func() (*http.Response, error) {
			return c.api.GetUserAttributeTenantAssignments(ctx, a, &openapi.GetUserAttributeTenantAssignmentsParams{Page: page})
		})
	})
}
