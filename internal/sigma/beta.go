package sigma

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// listByPageToken follows Sigma's nextPageToken / pageToken pagination.
func listByPageToken[T any](ctx context.Context, client *Client, path string) ([]T, error) {
	var all []T
	next := path
	seen := map[string]struct{}{}
	for pageNum := 0; pageNum < maxListPages; pageNum++ {
		var page struct {
			Entries []T    `json:"entries"`
			Next    string `json:"nextPageToken"`
		}
		if err := client.getJSON(ctx, next, &page); err != nil {
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
		parsed, err := url.Parse(path)
		if err != nil {
			return nil, err
		}
		query := parsed.Query()
		query.Set("pageToken", page.Next)
		parsed.RawQuery = query.Encode()
		next = parsed.String()
	}
	return nil, fmt.Errorf("sigma pagination exceeded %d pages for %s", maxListPages, path)
}

// Tenant is a Sigma tenant organization (Beta).
type Tenant struct {
	TenantOrganizationID   string  `json:"tenantOrganizationId"`
	ParentOrganizationID   string  `json:"parentOrganizationId"`
	CreatedBy              string  `json:"createdBy"`
	UpdatedBy              string  `json:"updatedBy"`
	CreatedAt              string  `json:"createdAt"`
	UpdatedAt              string  `json:"updatedAt"`
	TenantOrganizationName string  `json:"tenantOrganizationName"`
	TenantOrganizationSlug string  `json:"tenantOrganizationSlug"`
	SharedAt               *string `json:"sharedAt"`
	TenantCloudProvider    *string `json:"tenantCloudProvider"`
	TenantRegion           *string `json:"tenantRegion"`
	TenantAPIURL           *string `json:"tenantApiUrl"`
}

type CreateTenantInput struct {
	TenantOrganizationName string  `json:"tenantOrganizationName"`
	TenantOrganizationSlug string  `json:"tenantOrganizationSlug"`
	CloudProvider          *string `json:"cloudProvider,omitempty"`
}

type PatchTenantInput struct {
	TenantOrganizationName *string `json:"tenantOrganizationName,omitempty"`
	TenantOrganizationSlug *string `json:"tenantOrganizationSlug,omitempty"`
}

func (c *Client) CreateTenant(ctx context.Context, in CreateTenantInput) (*Tenant, error) {
	var value Tenant
	err := c.sendJSON(ctx, http.MethodPost, "/v2/tenants", in, &value)
	return &value, err
}

func (c *Client) GetTenant(ctx context.Context, id string) (*Tenant, error) {
	var value Tenant
	err := c.getJSON(ctx, "/v2/tenants/"+url.PathEscape(id), &value)
	return &value, err
}

func (c *Client) PatchTenant(ctx context.Context, id string, in PatchTenantInput) (*Tenant, error) {
	var value Tenant
	err := c.sendJSON(ctx, http.MethodPatch, "/v2/tenants/"+url.PathEscape(id), in, &value)
	return &value, err
}

func (c *Client) DeleteTenant(ctx context.Context, id string) error {
	return c.sendJSON(ctx, http.MethodDelete, "/v2/tenants/"+url.PathEscape(id), nil, nil)
}

func (c *Client) ListTenants(ctx context.Context) ([]Tenant, error) {
	return listByPageToken[Tenant](ctx, c, "/v2/tenants")
}

// TenantDeploymentCapability is a deploy-to target for a tenant (Beta).
type TenantDeploymentCapability struct {
	TenantOrganizationID string `json:"tenantOrganizationId"`
}

func (c *Client) ListTenantDeploymentCapabilities(ctx context.Context, tenantOrganizationID string) ([]TenantDeploymentCapability, error) {
	return listByPageToken[TenantDeploymentCapability](ctx, c, "/v2/tenants/"+url.PathEscape(tenantOrganizationID)+"/capabilities/deployments")
}

func (c *Client) AddTenantDeploymentCapabilities(ctx context.Context, tenantOrganizationID string, deployTo []string) error {
	var value struct {
		Capabilities []TenantDeploymentCapability `json:"capabilities"`
	}
	return c.sendJSON(ctx, http.MethodPost, "/v2/tenants/"+url.PathEscape(tenantOrganizationID)+"/capabilities/deployments:batchAdd",
		map[string]any{"deployToTenantOrganizationIds": deployTo}, &value)
}

func (c *Client) RemoveTenantDeploymentCapabilities(ctx context.Context, tenantOrganizationID string, deployTo []string) error {
	return c.sendJSON(ctx, http.MethodPost, "/v2/tenants/"+url.PathEscape(tenantOrganizationID)+"/capabilities/deployments:batchRemove",
		map[string]any{"deployToTenantOrganizationIds": deployTo}, nil)
}

// DeploymentPolicy is a Sigma deployment policy (Beta).
type DeploymentPolicy struct {
	DeploymentPolicyID string   `json:"deploymentPolicyId"`
	Name               string   `json:"name"`
	NameInTenant       string   `json:"nameInTenant"`
	VersionTagID       *string  `json:"versionTagId"`
	SourceSwapPolicies []string `json:"sourceSwapPolicies"`
	CopyInputTableData bool     `json:"copyInputTableData"`
}

type CreateDeploymentPolicyInput struct {
	Name               string   `json:"name"`
	VersionTagID       *string  `json:"versionTagId,omitempty"`
	SourceSwapPolicies []string `json:"sourceSwapPolicies,omitempty"`
	NameInTenant       *string  `json:"nameInTenant,omitempty"`
	CopyInputTableData *bool    `json:"copyInputTableData,omitempty"`
}

type UpdateDeploymentPolicyInput struct {
	Name               *string   `json:"name,omitempty"`
	NameInTenant       *string   `json:"nameInTenant,omitempty"`
	SourceSwapPolicies *[]string `json:"sourceSwapPolicies,omitempty"`
	CopyInputTableData *bool     `json:"copyInputTableData,omitempty"`
}

type createDeploymentPolicyResponse struct {
	DeploymentPolicyID string `json:"deploymentPolicyId"`
}

type DeploymentPolicyFile struct {
	InodeID            string `json:"inodeId"`
	DeploymentPolicyID string `json:"deploymentPolicyId"`
}

func (c *Client) CreateDeploymentPolicy(ctx context.Context, in CreateDeploymentPolicyInput) (*DeploymentPolicy, error) {
	var created createDeploymentPolicyResponse
	if err := c.sendJSON(ctx, http.MethodPost, "/v2/deploymentPolicies", in, &created); err != nil {
		return nil, err
	}
	return c.GetDeploymentPolicy(ctx, created.DeploymentPolicyID)
}

func (c *Client) GetDeploymentPolicy(ctx context.Context, id string) (*DeploymentPolicy, error) {
	var value DeploymentPolicy
	err := c.getJSON(ctx, "/v2/deploymentPolicies/"+url.PathEscape(id), &value)
	return &value, err
}

func (c *Client) UpdateDeploymentPolicy(ctx context.Context, id string, in UpdateDeploymentPolicyInput) (*DeploymentPolicy, error) {
	var value DeploymentPolicy
	err := c.sendJSON(ctx, http.MethodPatch, "/v2/deploymentPolicies/"+url.PathEscape(id), in, &value)
	return &value, err
}

func (c *Client) ArchiveDeploymentPolicy(ctx context.Context, id string) error {
	return c.sendJSON(ctx, http.MethodDelete, "/v2/deploymentPolicies/"+url.PathEscape(id), nil, nil)
}

func (c *Client) ListDeploymentPolicies(ctx context.Context) ([]DeploymentPolicy, error) {
	return listByPageToken[DeploymentPolicy](ctx, c, "/v2/deploymentPolicies")
}

func (c *Client) ListDeploymentPolicyInodes(ctx context.Context, deploymentPolicyID string) ([]DeploymentPolicyFile, error) {
	return listByPageToken[DeploymentPolicyFile](ctx, c, "/v2/deploymentPolicies/"+url.PathEscape(deploymentPolicyID)+"/files")
}

func (c *Client) AddDeploymentPolicyInodes(ctx context.Context, deploymentPolicyID string, inodeIDs []string) error {
	return c.sendJSON(ctx, http.MethodPost, "/v2/deploymentPolicies/"+url.PathEscape(deploymentPolicyID)+"/files",
		map[string]any{"inodeIds": inodeIDs}, nil)
}

func (c *Client) RemoveDeploymentPolicyInode(ctx context.Context, deploymentPolicyID, inodeID string) error {
	return c.sendJSON(ctx, http.MethodDelete, "/v2/deploymentPolicies/"+url.PathEscape(deploymentPolicyID)+"/files/"+url.PathEscape(inodeID), nil, nil)
}

func (c *Client) ListDeploymentPolicyTenants(ctx context.Context, deploymentPolicyID string) ([]string, error) {
	return listByPageToken[string](ctx, c, "/v2/deploymentPolicies/"+url.PathEscape(deploymentPolicyID)+"/tenants")
}

func (c *Client) AddDeploymentPolicyTenant(ctx context.Context, deploymentPolicyID, tenantOrganizationID string) error {
	return c.sendJSON(ctx, http.MethodPost, "/v2/deploymentPolicies/"+url.PathEscape(deploymentPolicyID)+"/tenants",
		map[string]any{"tenantOrganizationId": tenantOrganizationID}, nil)
}

func (c *Client) RemoveDeploymentPolicyTenant(ctx context.Context, deploymentPolicyID, tenantOrganizationID string) error {
	return c.sendJSON(ctx, http.MethodDelete, "/v2/deploymentPolicies/"+url.PathEscape(deploymentPolicyID)+"/tenants/"+url.PathEscape(tenantOrganizationID), nil, nil)
}

// SourceSwapPolicy is a Sigma source swap policy (Beta).
type SourceSwapPolicy struct {
	PolicyID         string          `json:"policyId"`
	Type             string          `json:"type"`
	Name             string          `json:"name"`
	FromConnectionID string          `json:"fromConnectionId"`
	Swaps            json.RawMessage `json:"swaps"`
}

type SourceSwapPolicyInput struct {
	Type             string          `json:"type"`
	Name             string          `json:"name"`
	FromConnectionID string          `json:"fromConnectionId"`
	Swaps            json.RawMessage `json:"swaps"`
}

type createSourceSwapPolicyResponse struct {
	PolicyID string `json:"policyId"`
}

func (c *Client) CreateSourceSwapPolicy(ctx context.Context, in SourceSwapPolicyInput) (*SourceSwapPolicy, error) {
	var created createSourceSwapPolicyResponse
	if err := c.sendJSON(ctx, http.MethodPost, "/v2/sourceSwapPolicies", in, &created); err != nil {
		return nil, err
	}
	return c.GetSourceSwapPolicy(ctx, created.PolicyID)
}

func (c *Client) GetSourceSwapPolicy(ctx context.Context, id string) (*SourceSwapPolicy, error) {
	var value SourceSwapPolicy
	err := c.getJSON(ctx, "/v2/sourceSwapPolicies/"+url.PathEscape(id), &value)
	return &value, err
}

func (c *Client) UpdateSourceSwapPolicy(ctx context.Context, id string, in SourceSwapPolicyInput) (*SourceSwapPolicy, error) {
	var created createSourceSwapPolicyResponse
	if err := c.sendJSON(ctx, http.MethodPatch, "/v2/sourceSwapPolicies/"+url.PathEscape(id), in, &created); err != nil {
		return nil, err
	}
	return c.GetSourceSwapPolicy(ctx, id)
}

func (c *Client) DeleteSourceSwapPolicy(ctx context.Context, id string) error {
	return c.sendJSON(ctx, http.MethodDelete, "/v2/sourceSwapPolicies/"+url.PathEscape(id), nil, nil)
}

func (c *Client) ListSourceSwapPolicies(ctx context.Context) ([]SourceSwapPolicy, error) {
	return listByPageToken[SourceSwapPolicy](ctx, c, "/v2/sourceSwapPolicies")
}
