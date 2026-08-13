package sigma

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Workspace is a Sigma workspace.
type Workspace struct {
	WorkspaceID    string `json:"workspaceId"`
	WorkspaceURLID string `json:"workspaceUrlId"`
	Name           string `json:"name"`
	CreatedBy      string `json:"createdBy"`
	UpdatedBy      string `json:"updatedBy"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

// CreateWorkspaceInput contains fields accepted by the create workspace API.
type CreateWorkspaceInput struct {
	Name         string `json:"name"`
	NoDuplicates bool   `json:"noDuplicates,omitempty"`
}

// UpdateWorkspaceInput contains fields accepted by the update workspace API.
type UpdateWorkspaceInput struct {
	Name         string `json:"name"`
	NoDuplicates bool   `json:"noDuplicates,omitempty"`
}

func (c *Client) CreateWorkspace(ctx context.Context, input CreateWorkspaceInput) (*Workspace, error) {
	var value Workspace
	err := c.sendJSON(ctx, http.MethodPost, "/v2/workspaces", input, &value)
	return &value, err
}

func (c *Client) GetWorkspace(ctx context.Context, id string) (*Workspace, error) {
	var value Workspace
	err := c.getJSON(ctx, "/v2/workspaces/"+url.PathEscape(id), &value)
	return &value, err
}

func (c *Client) UpdateWorkspace(ctx context.Context, id string, input UpdateWorkspaceInput) (*Workspace, error) {
	var value Workspace
	err := c.sendJSON(ctx, http.MethodPatch, "/v2/workspaces/"+url.PathEscape(id), input, &value)
	return &value, err
}

func (c *Client) DeleteWorkspace(ctx context.Context, id string) error {
	return c.sendJSON(ctx, http.MethodDelete, "/v2/workspaces/"+url.PathEscape(id), nil, nil)
}

func (c *Client) ListWorkspaces(ctx context.Context) ([]Workspace, error) {
	return ListAll[Workspace](ctx, c, "/v2.1/workspaces")
}

// File is a Sigma inode returned by the files API.
type File struct {
	ID                string  `json:"id"`
	URLID             string  `json:"urlId"`
	Name              string  `json:"name"`
	Type              string  `json:"type"`
	ParentID          string  `json:"parentId"`
	ParentURLID       string  `json:"parentUrlId"`
	Permission        string  `json:"permission"`
	Path              string  `json:"path"`
	Badge             *string `json:"badge"`
	IsArchived        bool    `json:"isArchived"`
	Description       string  `json:"description"`
	OwnerID           *string `json:"ownerId"`
	ParentSourceURLID string  `json:"parentSourceUrlId"`
	CreatedBy         string  `json:"createdBy"`
	UpdatedBy         string  `json:"updatedBy"`
	CreatedAt         string  `json:"createdAt"`
	UpdatedAt         string  `json:"updatedAt"`
}

// CreateFileInput contains fields common to creatable Sigma file types.
type CreateFileInput struct {
	Type        string           `json:"type"`
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	OwnerID     string           `json:"ownerId,omitempty"`
	ParentID    string           `json:"parentId,omitempty"`
	Source      *FileSourceInput `json:"source,omitempty"`
}

// FileSourceInput copies a workbook from an existing inode version on create.
type FileSourceInput struct {
	InodeID string `json:"inodeId"`
	Version int64  `json:"version"`
}

// UpdateFileInput contains mutable file fields.
type UpdateFileInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	OwnerID     *string `json:"ownerId,omitempty"`
	ParentID    *string `json:"parentId,omitempty"`
	Restore     *bool   `json:"restore,omitempty"`
}

func (c *Client) CreateFile(ctx context.Context, input CreateFileInput) (*File, error) {
	var value File
	err := c.sendJSON(ctx, http.MethodPost, "/v2/files", input, &value)
	return &value, err
}

func (c *Client) GetFile(ctx context.Context, id string) (*File, error) {
	var value File
	err := c.getJSON(ctx, "/v2/files/"+url.PathEscape(id), &value)
	return &value, err
}

func (c *Client) UpdateFile(ctx context.Context, id string, input UpdateFileInput) (*File, error) {
	var value File
	err := c.sendJSON(ctx, http.MethodPatch, "/v2/files/"+url.PathEscape(id), input, &value)
	return &value, err
}

func (c *Client) DeleteFile(ctx context.Context, id string) error {
	return c.sendJSON(ctx, http.MethodDelete, "/v2/files/"+url.PathEscape(id), nil, nil)
}

// ListFilesOptions controls file list filters.
type ListFilesOptions struct {
	Name               string
	Permission         string
	TypeFilters        []string
	ParentID           string
	DirectChildrenOnly *bool
}

func (c *Client) ListFiles(ctx context.Context, options ListFilesOptions) ([]File, error) {
	query := url.Values{}
	if options.Name != "" {
		query.Set("name", options.Name)
	}
	if options.Permission != "" {
		query.Set("permissionFilter", options.Permission)
	}
	for _, fileType := range options.TypeFilters {
		query.Add("typeFilters", fileType)
	}
	if options.ParentID != "" {
		query.Set("parentId", options.ParentID)
	}
	if options.DirectChildrenOnly != nil {
		query.Set("directChildFilter", fmt.Sprintf("%t", *options.DirectChildrenOnly))
	}
	path := "/v2/files"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	return ListAll[File](ctx, c, path)
}

// Grant is a permission grant on a Sigma inode.
type Grant struct {
	GrantID        string  `json:"grantId"`
	InodeID        string  `json:"inodeId"`
	OrganizationID string  `json:"organizationId"`
	MemberID       *string `json:"memberId"`
	TeamID         *string `json:"teamId"`
	Permission     string  `json:"permission"`
	TagID          *string `json:"tagId"`
	CreatedBy      string  `json:"createdBy"`
	UpdatedBy      string  `json:"updatedBy"`
	CreatedAt      string  `json:"createdAt"`
	UpdatedAt      string  `json:"updatedAt"`
	InodeType      string  `json:"inodeType"`
}

// Grantee identifies exactly one member or team.
type Grantee struct {
	MemberID string `json:"memberId,omitempty"`
	TeamID   string `json:"teamId,omitempty"`
}

// CreateGrantInput contains fields accepted by the generic grants API.
type CreateGrantInput struct {
	Grantee    Grantee `json:"grantee"`
	Permission string  `json:"permission"`
	InodeID    string  `json:"inodeId"`
	TagID      string  `json:"tagId,omitempty"`
}

func (c *Client) CreateGrant(ctx context.Context, input CreateGrantInput) (*Grant, error) {
	var value Grant
	err := c.sendJSON(ctx, http.MethodPost, "/v2/grants", input, &value)
	return &value, err
}

func (c *Client) GetGrant(ctx context.Context, id string) (*Grant, error) {
	var value Grant
	err := c.getJSON(ctx, "/v2/grants/"+url.PathEscape(id), &value)
	return &value, err
}

func (c *Client) DeleteGrant(ctx context.Context, id string) error {
	return c.sendJSON(ctx, http.MethodDelete, "/v2/grants/"+url.PathEscape(id), nil, nil)
}

func (c *Client) ListGrants(ctx context.Context, inodeID string) ([]Grant, error) {
	path := "/v2/grants"
	if inodeID != "" {
		path += "?inodeId=" + url.QueryEscape(inodeID)
	}
	return ListAll[Grant](ctx, c, path)
}

func (c *Client) ListWorkspaceGrants(ctx context.Context, workspaceID string) ([]Grant, error) {
	return ListAll[Grant](ctx, c, "/v2/workspaces/"+url.PathEscape(workspaceID)+"/grants")
}

func (c *Client) CreateWorkspaceGrant(ctx context.Context, workspaceID string, grantee Grantee, permission string) error {
	payload := map[string]any{"grants": []any{map[string]any{"grantee": grantee, "permission": permission}}}
	return c.sendJSON(ctx, http.MethodPost, "/v2/workspaces/"+url.PathEscape(workspaceID)+"/grants", payload, nil)
}

func (c *Client) DeleteWorkspaceGrant(ctx context.Context, workspaceID, grantID string) error {
	path := "/v2/workspaces/" + url.PathEscape(workspaceID) + "/grants/" + url.PathEscape(grantID)
	return c.sendJSON(ctx, http.MethodDelete, path, nil, nil)
}

func (c *Client) CreateDocumentGrant(ctx context.Context, kind, inodeID string, grantee Grantee, permission, tagID string) error {
	grant := map[string]any{"grantee": grantee, "permission": permission}
	if tagID != "" {
		grant["tagId"] = tagID
	}
	payload := map[string]any{"grants": []any{grant}}
	path := "/v2/" + url.PathEscape(kind) + "/" + url.PathEscape(inodeID) + "/grants"
	return c.sendJSON(ctx, http.MethodPost, path, payload, nil)
}

func (c *Client) DeleteDocumentGrant(ctx context.Context, kind, inodeID, grantID string) error {
	path := "/v2/" + url.PathEscape(kind) + "/" + url.PathEscape(inodeID) + "/grants/" + url.PathEscape(grantID)
	return c.sendJSON(ctx, http.MethodDelete, path, nil, nil)
}
