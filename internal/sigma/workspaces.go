package sigma

import (
	"context"
	"fmt"
	"net/http"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma/openapi"
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
	body, err := encodeBody(input)
	if err != nil {
		return nil, err
	}
	var value Workspace
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.CreateWorkspaceWithBody(ctx, nil, jsonContentType, body)
	}, &value)
	return &value, err
}

func (c *Client) GetWorkspace(ctx context.Context, id string) (*Workspace, error) {
	var value Workspace
	err := c.doDecode(func() (*http.Response, error) {
		return c.api.GetWorkspace(ctx, id, nil)
	}, &value)
	return &value, err
}

func (c *Client) UpdateWorkspace(ctx context.Context, id string, input UpdateWorkspaceInput) (*Workspace, error) {
	body, err := encodeBody(input)
	if err != nil {
		return nil, err
	}
	var value Workspace
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.UpdateWorkspaceWithBody(ctx, id, nil, jsonContentType, body)
	}, &value)
	return &value, err
}

func (c *Client) DeleteWorkspace(ctx context.Context, id string) error {
	return c.doVoid(func() (*http.Response, error) {
		return c.api.DeleteWorkspace(ctx, id, nil)
	})
}

// ListWorkspacesOptions are documented v2.1 list workspace filters except pagination.
type ListWorkspacesOptions struct {
	Name      *string
	ExactName *string
}

func (c *Client) ListWorkspaces(ctx context.Context, opts ListWorkspacesOptions) ([]Workspace, error) {
	base := &openapi.V21ListWorkspacesParams{Name: opts.Name, ExactName: opts.ExactName}
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]Workspace, *string, error) {
		params := *base
		params.Page = page
		return fetchPage[Workspace](c, func() (*http.Response, error) {
			return c.api.V21ListWorkspaces(ctx, &params)
		})
	})
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
	body, err := encodeBody(input)
	if err != nil {
		return nil, err
	}
	var value File
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.CreateWithBody(ctx, nil, jsonContentType, body)
	}, &value)
	return &value, err
}

func (c *Client) GetFile(ctx context.Context, id string) (*File, error) {
	var value File
	err := c.doDecode(func() (*http.Response, error) {
		return c.api.Get(ctx, id, nil)
	}, &value)
	return &value, err
}

func (c *Client) UpdateFile(ctx context.Context, id string, input UpdateFileInput) (*File, error) {
	body, err := encodeBody(input)
	if err != nil {
		return nil, err
	}
	var value File
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.UpdateWithBody(ctx, id, nil, jsonContentType, body)
	}, &value)
	return &value, err
}

func (c *Client) DeleteFile(ctx context.Context, id string) error {
	return c.doVoid(func() (*http.Response, error) {
		return c.api.Delete(ctx, id, nil)
	})
}

// ListFilesOptions controls file list filters. Pointers preserve null versus explicit empty/false.
type ListFilesOptions struct {
	Name               *string
	Permission         *string
	TypeFilters        *[]string
	ParentID           *string
	DirectChildrenOnly *bool
}

func (c *Client) ListFiles(ctx context.Context, options ListFilesOptions) ([]File, error) {
	base, err := listFilesParams(options)
	if err != nil {
		return nil, err
	}
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]File, *string, error) {
		params := *base
		params.Page = page
		return fetchPage[File](c, func() (*http.Response, error) {
			return c.api.List(ctx, &params)
		})
	})
}

func listFilesParams(options ListFilesOptions) (*openapi.ListParams, error) {
	params := &openapi.ListParams{
		Name:              options.Name,
		ParentId:          options.ParentID,
		DirectChildFilter: options.DirectChildrenOnly,
	}
	if options.Permission != nil {
		var filter openapi.V2FilesGetParametersPermissionFilter
		if err := filter.FromV2FilesGetParametersPermissionFilter0(openapi.V2FilesGetParametersPermissionFilter0(*options.Permission)); err != nil {
			return nil, fmt.Errorf("encode file permissionFilter: %w", err)
		}
		params.PermissionFilter = &filter
	}
	if options.TypeFilters != nil {
		items := make(openapi.V2FilesGetParametersTypeFilters0, len(*options.TypeFilters))
		for i, fileType := range *options.TypeFilters {
			items[i] = openapi.V2FilesGetParametersTypeFiltersSchemaOneOf0Items(fileType)
		}
		var filters openapi.V2FilesGetParametersTypeFilters
		if err := filters.FromV2FilesGetParametersTypeFilters0(items); err != nil {
			return nil, fmt.Errorf("encode file typeFilters: %w", err)
		}
		params.TypeFilters = &filters
	}
	return params, nil
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
	body, err := encodeBody(input)
	if err != nil {
		return nil, err
	}
	var value Grant
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.CreateGrantWithBody(ctx, nil, jsonContentType, body)
	}, &value)
	return &value, err
}

func (c *Client) GetGrant(ctx context.Context, id string) (*Grant, error) {
	var value Grant
	err := c.doDecode(func() (*http.Response, error) {
		return c.api.GetGrant(ctx, id, nil)
	}, &value)
	return &value, err
}

func (c *Client) DeleteGrant(ctx context.Context, id string) error {
	return c.doVoid(func() (*http.Response, error) {
		return c.api.DeleteGrant(ctx, id, nil)
	})
}

func (c *Client) ListGrants(ctx context.Context, inodeID string) ([]Grant, error) {
	base := &openapi.ListGrantsParams{}
	if inodeID != "" {
		base.InodeId = &inodeID
	}
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]Grant, *string, error) {
		params := *base
		params.Page = page
		return fetchPage[Grant](c, func() (*http.Response, error) {
			return c.api.ListGrants(ctx, &params)
		})
	})
}

func (c *Client) ListWorkspaceGrants(ctx context.Context, workspaceID string) ([]Grant, error) {
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]Grant, *string, error) {
		return fetchPage[Grant](c, func() (*http.Response, error) {
			return c.api.ListWorkspaceGrants(ctx, workspaceID, &openapi.ListWorkspaceGrantsParams{Page: page})
		})
	})
}

func (c *Client) CreateWorkspaceGrant(ctx context.Context, workspaceID string, grantee Grantee, permission string) error {
	payload := map[string]any{"grants": []any{map[string]any{"grantee": grantee, "permission": permission}}}
	body, err := encodeBody(payload)
	if err != nil {
		return err
	}
	return c.doVoid(func() (*http.Response, error) {
		return c.api.CreateWorkspaceGrantWithBody(ctx, workspaceID, nil, jsonContentType, body)
	})
}

func (c *Client) DeleteWorkspaceGrant(ctx context.Context, workspaceID, grantID string) error {
	return c.doVoid(func() (*http.Response, error) {
		return c.api.DeleteWorkspaceGrant(ctx, workspaceID, grantID, nil)
	})
}

func (c *Client) CreateDocumentGrant(ctx context.Context, kind, inodeID string, grantee Grantee, permission, tagID string) error {
	grant := map[string]any{"grantee": grantee, "permission": permission}
	if tagID != "" {
		grant["tagId"] = tagID
	}
	body, err := encodeBody(map[string]any{"grants": []any{grant}})
	if err != nil {
		return err
	}
	switch kind {
	case "workbooks":
		return c.doVoid(func() (*http.Response, error) {
			return c.api.CreateWorkbookGrantWithBody(ctx, inodeID, nil, jsonContentType, body)
		})
	case "reports":
		return c.doVoid(func() (*http.Response, error) {
			return c.api.CreateReportGrantWithBody(ctx, inodeID, nil, jsonContentType, body)
		})
	default:
		return fmt.Errorf("unsupported document grant kind %q", kind)
	}
}

func (c *Client) DeleteDocumentGrant(ctx context.Context, kind, inodeID, grantID string) error {
	switch kind {
	case "workbooks":
		return c.doVoid(func() (*http.Response, error) {
			return c.api.DeleteWorkbookGrant(ctx, inodeID, grantID, nil)
		})
	case "reports":
		return c.doVoid(func() (*http.Response, error) {
			return c.api.DeleteReportGrant(ctx, inodeID, grantID, nil)
		})
	default:
		return fmt.Errorf("unsupported document grant kind %q", kind)
	}
}
