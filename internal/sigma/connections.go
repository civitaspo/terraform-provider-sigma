package sigma

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// Connection is the stable projection returned by Sigma's connection APIs.
// Warehouse-specific configuration is intentionally represented by the request
// Details field because the get endpoint does not return that configuration.
type Connection struct {
	ConnectionID string          `json:"connectionId"`
	Name         string          `json:"name"`
	Type         string          `json:"type"`
	Description  json.RawMessage `json:"description"`
	PoolSizes    json.RawMessage `json:"poolSizes"`
	TimeoutSecs  *float64        `json:"timeoutSecs"`
	FriendlyName bool            `json:"friendlyName"`
}

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
	var value Connection
	err := c.sendJSON(ctx, http.MethodPost, "/v2/connections", in, &value)
	return &value, err
}

func (c *Client) GetConnection(ctx context.Context, id string) (*Connection, error) {
	var value Connection
	err := c.getJSON(ctx, "/v2/connections/"+url.PathEscape(id), &value)
	return &value, err
}

func (c *Client) UpdateConnection(ctx context.Context, id string, in ConnectionInput) (*Connection, error) {
	var value Connection
	err := c.sendJSON(ctx, http.MethodPut, "/v2/connections/"+url.PathEscape(id), in, &value)
	return &value, err
}

func (c *Client) DeleteConnection(ctx context.Context, id string) error {
	return c.sendJSON(ctx, http.MethodDelete, "/v2/connections/"+url.PathEscape(id), nil, nil)
}

func (c *Client) TestConnection(ctx context.Context, id string) (*ConnectionTest, error) {
	var value ConnectionTest
	err := c.getJSON(ctx, "/v2/connections/"+url.PathEscape(id)+"/test", &value)
	return &value, err
}

func (c *Client) ListConnections(ctx context.Context) ([]Connection, error) {
	return ListAll[Connection](ctx, c, "/v2/connections")
}

type Grant struct {
	GrantID    string  `json:"grantId"`
	InodeID    string  `json:"inodeId"`
	MemberID   *string `json:"memberId"`
	TeamID     *string `json:"teamId"`
	Permission string  `json:"permission"`
}

func (c *Client) CreateConnectionGrant(ctx context.Context, connectionID, memberID, teamID, permission string) error {
	grantee := map[string]string{}
	if memberID != "" {
		grantee["memberId"] = memberID
	} else {
		grantee["teamId"] = teamID
	}
	payload := map[string]any{"grants": []any{map[string]any{"grantee": grantee, "permission": permission}}}
	return c.sendJSON(ctx, http.MethodPost, "/v2/connections/"+url.PathEscape(connectionID)+"/grants", payload, nil)
}

func (c *Client) ListConnectionGrants(ctx context.Context, connectionID string) ([]Grant, error) {
	return ListAll[Grant](ctx, c, "/v2/connections/"+url.PathEscape(connectionID)+"/grants")
}

func (c *Client) DeleteConnectionGrant(ctx context.Context, connectionID, grantID string) error {
	return c.sendJSON(ctx, http.MethodDelete, "/v2/connections/"+url.PathEscape(connectionID)+"/grants/"+url.PathEscape(grantID), nil, nil)
}

func (c *Client) CreateConnectionPathGrant(ctx context.Context, pathID, memberID, teamID, permission string) error {
	grantee := map[string]string{}
	if memberID != "" {
		grantee["memberId"] = memberID
	} else {
		grantee["teamId"] = teamID
	}
	payload := map[string]any{"grants": []any{map[string]any{"grantee": grantee, "permission": permission}}}
	return c.sendJSON(ctx, http.MethodPost, "/v2/connections/paths/"+url.PathEscape(pathID)+"/grants", payload, nil)
}

func (c *Client) ListConnectionPathGrants(ctx context.Context, pathID string) ([]Grant, error) {
	return ListAll[Grant](ctx, c, "/v2/connections/paths/"+url.PathEscape(pathID)+"/grants")
}

func (c *Client) DeleteConnectionPathGrant(ctx context.Context, pathID, grantID string) error {
	return c.sendJSON(ctx, http.MethodDelete, "/v2/connections/paths/"+url.PathEscape(pathID)+"/grants/"+url.PathEscape(grantID), nil, nil)
}

type ConnectionPath struct {
	ConnectionID string   `json:"connectionId"`
	URLID        string   `json:"urlId"`
	Path         []string `json:"path"`
}

func (c *Client) ListConnectionPaths(ctx context.Context, connectionID string) ([]ConnectionPath, error) {
	path := "/v2/connections/paths"
	if connectionID != "" {
		path += "?connectionId=" + url.QueryEscape(connectionID)
	}
	return ListAll[ConnectionPath](ctx, c, path)
}

type ConnectionLookup struct {
	Kind    string `json:"kind"`
	InodeID string `json:"inodeId"`
	URL     string `json:"url"`
}

func (c *Client) LookupConnection(ctx context.Context, connectionID string, path []string) (*ConnectionLookup, error) {
	var value ConnectionLookup
	err := c.sendJSON(ctx, http.MethodPost, "/v2/connection/"+url.PathEscape(connectionID)+"/lookup", map[string]any{"path": path}, &value)
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
	Params      json.RawMessage `json:"params"`
	Config      json.RawMessage `json:"config,omitempty"`
	AuthID      *string         `json:"authId,omitempty"`
}

func (c *Client) CreateAPIConnector(ctx context.Context, in APIConnectorInput) (*APIConnector, error) {
	var value APIConnector
	err := c.sendJSON(ctx, http.MethodPost, "/v2/api-connectors", in, &value)
	return &value, err
}
func (c *Client) GetAPIConnector(ctx context.Context, id string) (*APIConnector, error) {
	var value APIConnector
	err := c.getJSON(ctx, "/v2/api-connectors/"+url.PathEscape(id), &value)
	return &value, err
}
func (c *Client) UpdateAPIConnector(ctx context.Context, id string, in APIConnectorInput) (*APIConnector, error) {
	var value APIConnector
	err := c.sendJSON(ctx, http.MethodPatch, "/v2/api-connectors/"+url.PathEscape(id), in, &value)
	return &value, err
}
func (c *Client) DeleteAPIConnector(ctx context.Context, id string) error {
	return c.sendJSON(ctx, http.MethodDelete, "/v2/api-connectors/"+url.PathEscape(id), nil, nil)
}
func (c *Client) ListAPIConnectors(ctx context.Context) ([]APIConnector, error) {
	var all []APIConnector
	path := "/v2/api-connectors"
	for {
		var page struct {
			Entries []APIConnector `json:"entries"`
			Next    string         `json:"nextPageToken"`
		}
		if err := c.getJSON(ctx, path, &page); err != nil {
			return nil, err
		}
		all = append(all, page.Entries...)
		if page.Next == "" {
			return all, nil
		}
		path = "/v2/api-connectors?pageToken=" + url.QueryEscape(page.Next)
	}
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
	Credential  json.RawMessage `json:"credential"`
	Description string          `json:"description,omitempty"`
}

func (c *Client) CreateAPICredential(ctx context.Context, in APICredentialInput) (*APICredential, error) {
	var value APICredential
	err := c.sendJSON(ctx, http.MethodPost, "/v2/api-credentials", in, &value)
	return &value, err
}
func (c *Client) GetAPICredential(ctx context.Context, id string) (*APICredential, error) {
	var value APICredential
	err := c.getJSON(ctx, "/v2/api-credentials/"+url.PathEscape(id), &value)
	return &value, err
}
func (c *Client) UpdateAPICredential(ctx context.Context, id string, in APICredentialInput) (*APICredential, error) {
	var value APICredential
	err := c.sendJSON(ctx, http.MethodPatch, "/v2/api-credentials/"+url.PathEscape(id), in, &value)
	return &value, err
}
func (c *Client) DeleteAPICredential(ctx context.Context, id string) error {
	return c.sendJSON(ctx, http.MethodDelete, "/v2/api-credentials/"+url.PathEscape(id), nil, nil)
}
func (c *Client) ListAPICredentials(ctx context.Context) ([]APICredential, error) {
	var all []APICredential
	path := "/v2/api-credentials"
	for {
		var page struct {
			Entries []APICredential `json:"entries"`
			Next    string          `json:"nextPageToken"`
		}
		if err := c.getJSON(ctx, path, &page); err != nil {
			return nil, err
		}
		all = append(all, page.Entries...)
		if page.Next == "" {
			return all, nil
		}
		path = "/v2/api-credentials?pageToken=" + url.QueryEscape(page.Next)
	}
}
