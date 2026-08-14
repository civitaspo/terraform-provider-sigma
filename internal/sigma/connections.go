package sigma

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma/openapi"
)

// ConnectionTimeout is the GET response timeout object (default plus optional
// worksheet/dashboard/download overrides). Request bodies use TimeoutSecs.
type ConnectionTimeout struct {
	Default   float64  `json:"default"`
	Dashboard *float64 `json:"dashboard,omitempty"`
	Download  *float64 `json:"download,omitempty"`
	Worksheet *float64 `json:"worksheet,omitempty"`
}

// Connection is the stable projection returned by Sigma's connection APIs.
// Warehouse-specific configuration is intentionally represented by the request
// Details field because the get endpoint does not return that configuration.
// UseOauth is returned by GET as useOauth and is not a request field on PUT.
// TimeoutSecs is mapped from response timeout.default (and still accepted from
// legacy timeoutSecs JSON). FriendlyName comes from response friendlyName.
type Connection struct {
	ConnectionID             string             `json:"connectionId"`
	Name                     string             `json:"name"`
	Type                     string             `json:"type"`
	Description              json.RawMessage    `json:"description"`
	PoolSizes                json.RawMessage    `json:"poolSizes"`
	TimeoutSecs              *float64           `json:"timeoutSecs"`
	FriendlyName             bool               `json:"friendlyName"`
	UseOauth                 *bool              `json:"useOauth"`
	Timeout                  *ConnectionTimeout `json:"timeout,omitempty"`
	OrganizationID           string             `json:"organizationId"`
	IsSample                 *bool              `json:"isSample,omitempty"`
	IsAuditLog               *bool              `json:"isAuditLog,omitempty"`
	LastActiveAt             string             `json:"lastActiveAt"`
	CreatedBy                string             `json:"createdBy"`
	UpdatedBy                string             `json:"updatedBy"`
	CreatedAt                string             `json:"createdAt"`
	UpdatedAt                string             `json:"updatedAt"`
	IsArchived               *bool              `json:"isArchived,omitempty"`
	Account                  string             `json:"account"`
	Warehouse                string             `json:"warehouse"`
	User                     string             `json:"user"`
	Role                     string             `json:"role"`
	WriteAccess              *bool              `json:"writeAccess,omitempty"`
	Writebacks               json.RawMessage    `json:"writebacks"`
	WritebackSchemas         json.RawMessage    `json:"writebackSchemas"`
	InputTableAuditLogSchema json.RawMessage    `json:"inputTableAuditLogSchema"`
	MaterializationWarehouse string             `json:"materializationWarehouse"`
	ExportsWarehouse         string             `json:"exportsWarehouse"`
	OauthMetadataURL         string             `json:"oauthMetadataUrl"`
	OauthClientID            string             `json:"oauthClientId"`
	OauthScopes              []string           `json:"oauthScopes"`
	OauthIdpType             string             `json:"oauthIdpType"`
	OauthUsePkce             *bool              `json:"oauthUsePkce,omitempty"`
	OauthUseJwt              *bool              `json:"oauthUseJwt,omitempty"`
	OauthAudience            string             `json:"oauthAudience"`
	IsIndependentOAuth       *bool              `json:"isIndependentOAuth,omitempty"`
	UserAttributes           json.RawMessage    `json:"userAttributes"`
	RoleSwitching            string             `json:"roleSwitching"`
}

func (value *Connection) UnmarshalJSON(data []byte) error {
	var raw struct {
		ConnectionID             string             `json:"connectionId"`
		Name                     string             `json:"name"`
		Type                     string             `json:"type"`
		Description              json.RawMessage    `json:"description"`
		PoolSizes                json.RawMessage    `json:"poolSizes"`
		TimeoutSecs              *float64           `json:"timeoutSecs"`
		FriendlyName             *bool              `json:"friendlyName"`
		UseFriendlyNames         *bool              `json:"useFriendlyNames"`
		UseOauth                 *bool              `json:"useOauth"`
		Timeout                  *ConnectionTimeout `json:"timeout"`
		OrganizationID           string             `json:"organizationId"`
		IsSample                 *bool              `json:"isSample"`
		IsAuditLog               *bool              `json:"isAuditLog"`
		LastActiveAt             string             `json:"lastActiveAt"`
		CreatedBy                string             `json:"createdBy"`
		UpdatedBy                string             `json:"updatedBy"`
		CreatedAt                string             `json:"createdAt"`
		UpdatedAt                string             `json:"updatedAt"`
		IsArchived               *bool              `json:"isArchived"`
		Account                  string             `json:"account"`
		Warehouse                string             `json:"warehouse"`
		User                     string             `json:"user"`
		Role                     string             `json:"role"`
		WriteAccess              *bool              `json:"writeAccess"`
		Writebacks               json.RawMessage    `json:"writebacks"`
		WritebackSchemas         json.RawMessage    `json:"writebackSchemas"`
		InputTableAuditLogSchema json.RawMessage    `json:"inputTableAuditLogSchema"`
		MaterializationWarehouse string             `json:"materializationWarehouse"`
		ExportsWarehouse         string             `json:"exportsWarehouse"`
		OauthMetadataURL         string             `json:"oauthMetadataUrl"`
		OauthClientID            string             `json:"oauthClientId"`
		OauthScopes              []string           `json:"oauthScopes"`
		OauthIdpType             string             `json:"oauthIdpType"`
		OauthUsePkce             *bool              `json:"oauthUsePkce"`
		OauthUseJwt              *bool              `json:"oauthUseJwt"`
		OauthAudience            string             `json:"oauthAudience"`
		IsIndependentOAuth       *bool              `json:"isIndependentOAuth"`
		UserAttributes           json.RawMessage    `json:"userAttributes"`
		RoleSwitching            string             `json:"roleSwitching"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	value.ConnectionID = raw.ConnectionID
	value.Name = raw.Name
	value.Type = raw.Type
	value.Description = raw.Description
	value.PoolSizes = raw.PoolSizes
	value.UseOauth = raw.UseOauth
	value.Timeout = raw.Timeout
	value.OrganizationID = raw.OrganizationID
	value.IsSample = raw.IsSample
	value.IsAuditLog = raw.IsAuditLog
	value.LastActiveAt = raw.LastActiveAt
	value.CreatedBy = raw.CreatedBy
	value.UpdatedBy = raw.UpdatedBy
	value.CreatedAt = raw.CreatedAt
	value.UpdatedAt = raw.UpdatedAt
	value.IsArchived = raw.IsArchived
	value.Account = raw.Account
	value.Warehouse = raw.Warehouse
	value.User = raw.User
	value.Role = raw.Role
	value.WriteAccess = raw.WriteAccess
	value.Writebacks = raw.Writebacks
	value.WritebackSchemas = raw.WritebackSchemas
	value.InputTableAuditLogSchema = raw.InputTableAuditLogSchema
	value.MaterializationWarehouse = raw.MaterializationWarehouse
	value.ExportsWarehouse = raw.ExportsWarehouse
	value.OauthMetadataURL = raw.OauthMetadataURL
	value.OauthClientID = raw.OauthClientID
	value.OauthScopes = raw.OauthScopes
	value.OauthIdpType = raw.OauthIdpType
	value.OauthUsePkce = raw.OauthUsePkce
	value.OauthUseJwt = raw.OauthUseJwt
	value.OauthAudience = raw.OauthAudience
	value.IsIndependentOAuth = raw.IsIndependentOAuth
	value.UserAttributes = raw.UserAttributes
	value.RoleSwitching = raw.RoleSwitching
	switch {
	case raw.TimeoutSecs != nil:
		value.TimeoutSecs = raw.TimeoutSecs
	case raw.Timeout != nil:
		seconds := raw.Timeout.Default
		value.TimeoutSecs = &seconds
	}
	switch {
	case raw.FriendlyName != nil:
		value.FriendlyName = *raw.FriendlyName
	case raw.UseFriendlyNames != nil:
		value.FriendlyName = *raw.UseFriendlyNames
	}
	return nil
}

// ConnectionInput is the create/update request body. TimeoutSecs and
// UseFriendlyNames are request field names; they are not the GET response shape.
// Restore is intentionally omitted from this façade; Terraform does not expose it.
type ConnectionInput struct {
	Details          json.RawMessage `json:"details"`
	Name             string          `json:"name"`
	Description      json.RawMessage `json:"description,omitempty"`
	PoolSizes        json.RawMessage `json:"poolSizes,omitempty"`
	TimeoutSecs      *float64        `json:"timeoutSecs,omitempty"`
	UseFriendlyNames *bool           `json:"useFriendlyNames,omitempty"`
}

type ConnectionTest struct {
	Read  string `json:"read"`
	Write string `json:"write"`
}

func (c *Client) CreateConnection(ctx context.Context, in ConnectionInput) (*Connection, error) {
	body, err := encodeBody(in)
	if err != nil {
		return nil, err
	}
	var value Connection
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.CreateConnectionWithBody(ctx, nil, jsonContentType, body)
	}, &value)
	return &value, err
}

func (c *Client) GetConnection(ctx context.Context, id string) (*Connection, error) {
	var value Connection
	err := c.doDecode(func() (*http.Response, error) {
		return c.api.GetConnection(ctx, id, nil)
	}, &value)
	return &value, err
}

func (c *Client) UpdateConnection(ctx context.Context, id string, in ConnectionInput) (*Connection, error) {
	body, err := encodeBody(in)
	if err != nil {
		return nil, err
	}
	var value Connection
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.UpdateConnectionWithBody(ctx, id, nil, jsonContentType, body)
	}, &value)
	return &value, err
}

func (c *Client) DeleteConnection(ctx context.Context, id string) error {
	return c.doVoid(func() (*http.Response, error) {
		return c.api.DeleteConnection(ctx, id, nil)
	})
}

func (c *Client) TestConnection(ctx context.Context, id string) (*ConnectionTest, error) {
	var value ConnectionTest
	err := c.doDecode(func() (*http.Response, error) {
		return c.api.TestConnection(ctx, id, nil)
	}, &value)
	return &value, err
}

func (c *Client) ListConnections(ctx context.Context, opts ListConnectionsOptions) ([]Connection, error) {
	base := &openapi.ListConnectionsParams{IncludeArchived: opts.IncludeArchived, Search: opts.Search}
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]Connection, *string, error) {
		params := *base
		params.Page = page
		return fetchPage[Connection](c, func() (*http.Response, error) {
			return c.api.ListConnections(ctx, &params)
		})
	})
}

type ListConnectionsOptions struct {
	IncludeArchived *bool
	Search          *string
}

func (c *Client) CreateConnectionGrant(ctx context.Context, connectionID, memberID, teamID, permission string) error {
	grantee := map[string]string{}
	if memberID != "" {
		grantee["memberId"] = memberID
	} else {
		grantee["teamId"] = teamID
	}
	body, err := encodeBody(map[string]any{"grants": []any{map[string]any{"grantee": grantee, "permission": permission}}})
	if err != nil {
		return err
	}
	return c.doVoid(func() (*http.Response, error) {
		return c.api.CreateConnectionGrantWithBody(ctx, connectionID, nil, jsonContentType, body)
	})
}

func (c *Client) ListConnectionGrants(ctx context.Context, connectionID string) ([]Grant, error) {
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]Grant, *string, error) {
		return fetchPage[Grant](c, func() (*http.Response, error) {
			return c.api.ListConnectionGrants(ctx, connectionID, &openapi.ListConnectionGrantsParams{Page: page})
		})
	})
}

func (c *Client) DeleteConnectionGrant(ctx context.Context, connectionID, grantID string) error {
	return c.doVoid(func() (*http.Response, error) {
		return c.api.DeleteConnectionGrant(ctx, connectionID, grantID, nil)
	})
}

func (c *Client) CreateConnectionPathGrant(ctx context.Context, pathID, memberID, teamID, permission string) error {
	grantee := map[string]string{}
	if memberID != "" {
		grantee["memberId"] = memberID
	} else {
		grantee["teamId"] = teamID
	}
	body, err := encodeBody(map[string]any{"grants": []any{map[string]any{"grantee": grantee, "permission": permission}}})
	if err != nil {
		return err
	}
	return c.doVoid(func() (*http.Response, error) {
		return c.api.CreateConnectionPathGrantWithBody(ctx, pathID, nil, jsonContentType, body)
	})
}

func (c *Client) ListConnectionPathGrants(ctx context.Context, pathID string) ([]Grant, error) {
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]Grant, *string, error) {
		return fetchPage[Grant](c, func() (*http.Response, error) {
			return c.api.ListConnectionPathGrants(ctx, pathID, &openapi.ListConnectionPathGrantsParams{Page: page})
		})
	})
}

func (c *Client) DeleteConnectionPathGrant(ctx context.Context, pathID, grantID string) error {
	return c.doVoid(func() (*http.Response, error) {
		return c.api.DeleteConnectionPathGrant(ctx, pathID, grantID, nil)
	})
}

type ConnectionPath struct {
	ConnectionID string   `json:"connectionId"`
	URLID        string   `json:"urlId"`
	Path         []string `json:"path"`
}

type ListConnectionPathsOptions struct {
	ConnectionID *string
}

func (c *Client) ListConnectionPaths(ctx context.Context, opts ListConnectionPathsOptions) ([]ConnectionPath, error) {
	base := &openapi.ListConnectionPathsParams{ConnectionId: opts.ConnectionID}
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]ConnectionPath, *string, error) {
		params := *base
		params.Page = page
		return fetchPage[ConnectionPath](c, func() (*http.Response, error) {
			return c.api.ListConnectionPaths(ctx, &params)
		})
	})
}

type ConnectionLookup struct {
	Kind    string `json:"kind"`
	InodeID string `json:"inodeId"`
	URL     string `json:"url"`
}

func (c *Client) LookupConnection(ctx context.Context, connectionID string, path []string) (*ConnectionLookup, error) {
	body, err := encodeBody(map[string]any{"path": path})
	if err != nil {
		return nil, err
	}
	var value ConnectionLookup
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.LookupConnectionWithBody(ctx, connectionID, nil, jsonContentType, body)
	}, &value)
	return &value, err
}

type APIConnector struct {
	APIConnectorID string          `json:"apiConnectorId"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	Params         json.RawMessage `json:"params"`
	Config         json.RawMessage `json:"config"`
	AuthID         *string         `json:"authId"`
}

type APIConnectorInput struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Params      json.RawMessage `json:"params,omitempty"`
	Config      json.RawMessage `json:"config,omitempty"`
	AuthID      *string         `json:"authId,omitempty"`
}

func (c *Client) CreateAPIConnector(ctx context.Context, in APIConnectorInput) (*APIConnector, error) {
	body, err := encodeBody(in)
	if err != nil {
		return nil, err
	}
	var value APIConnector
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.CreateApiConnectorWithBody(ctx, nil, jsonContentType, body)
	}, &value)
	return &value, err
}
func (c *Client) GetAPIConnector(ctx context.Context, id string) (*APIConnector, error) {
	var value APIConnector
	err := c.doDecode(func() (*http.Response, error) {
		return c.api.GetApiConnector(ctx, id, nil)
	}, &value)
	return &value, err
}
func (c *Client) UpdateAPIConnector(ctx context.Context, id string, in APIConnectorInput) (*APIConnector, error) {
	body, err := encodeBody(in)
	if err != nil {
		return nil, err
	}
	var value APIConnector
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.UpdateApiConnectorWithBody(ctx, id, nil, jsonContentType, body)
	}, &value)
	return &value, err
}
func (c *Client) DeleteAPIConnector(ctx context.Context, id string) error {
	return c.doVoid(func() (*http.Response, error) {
		return c.api.DeleteApiConnector(ctx, id, nil)
	})
}
func (c *Client) ListAPIConnectors(ctx context.Context) ([]APIConnector, error) {
	return listAllByPageToken(ctx, func(ctx context.Context, pageToken *string) ([]APIConnector, *string, error) {
		return fetchPageToken[APIConnector](c, func() (*http.Response, error) {
			return c.api.ListApiConnectors(ctx, &openapi.ListApiConnectorsParams{PageToken: pageToken})
		})
	})
}

type APICredential struct {
	APICredentialID string          `json:"apiCredentialId"`
	Name            string          `json:"name"`
	AuthMethod      string          `json:"authMethod"`
	Allowlist       []string        `json:"allowlist"`
	Description     string          `json:"description"`
	Credential      json.RawMessage `json:"credential"`
}

type APICredentialInput struct {
	Name        string          `json:"name"`
	Allowlist   []string        `json:"allowlist"`
	Credential  json.RawMessage `json:"credential,omitempty"`
	Description string          `json:"description,omitempty"`
}

func (c *Client) CreateAPICredential(ctx context.Context, in APICredentialInput) (*APICredential, error) {
	body, err := encodeBody(in)
	if err != nil {
		return nil, err
	}
	var value APICredential
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.CreateApiCredentialWithBody(ctx, nil, jsonContentType, body)
	}, &value)
	return &value, err
}
func (c *Client) GetAPICredential(ctx context.Context, id string) (*APICredential, error) {
	var value APICredential
	err := c.doDecode(func() (*http.Response, error) {
		return c.api.GetApiCredential(ctx, id, nil)
	}, &value)
	return &value, err
}
func (c *Client) UpdateAPICredential(ctx context.Context, id string, in APICredentialInput) (*APICredential, error) {
	body, err := encodeBody(in)
	if err != nil {
		return nil, err
	}
	var value APICredential
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.UpdateApiCredentialWithBody(ctx, id, nil, jsonContentType, body)
	}, &value)
	return &value, err
}
func (c *Client) DeleteAPICredential(ctx context.Context, id string) error {
	return c.doVoid(func() (*http.Response, error) {
		return c.api.DeleteApiCredential(ctx, id, nil)
	})
}
func (c *Client) ListAPICredentials(ctx context.Context) ([]APICredential, error) {
	return listAllByPageToken(ctx, func(ctx context.Context, pageToken *string) ([]APICredential, *string, error) {
		return fetchPageToken[APICredential](c, func() (*http.Response, error) {
			return c.api.ListApiCredentials(ctx, &openapi.ListApiCredentialsParams{PageToken: pageToken})
		})
	})
}
