package sigma

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma/openapi"
)

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
	body, err := encodeBody(in)
	if err != nil {
		return nil, err
	}
	var value Tenant
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.CreateTenantWithBody(ctx, nil, jsonContentType, body)
	}, &value)
	return &value, err
}

func (c *Client) GetTenant(ctx context.Context, id string) (*Tenant, error) {
	var value Tenant
	err := c.doDecode(func() (*http.Response, error) {
		return c.api.GetTenant(ctx, id, nil)
	}, &value)
	return &value, err
}

func (c *Client) PatchTenant(ctx context.Context, id string, in PatchTenantInput) (*Tenant, error) {
	body, err := encodeBody(in)
	if err != nil {
		return nil, err
	}
	var value Tenant
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.PatchTenantWithBody(ctx, id, nil, jsonContentType, body)
	}, &value)
	return &value, err
}

func (c *Client) DeleteTenant(ctx context.Context, id string) error {
	return c.doVoid(func() (*http.Response, error) {
		return c.api.DeleteTenant(ctx, id, nil)
	})
}

type ListTenantsOptions struct {
	Key    *string
	Order  *string
	Search *string
}

func (c *Client) ListTenants(ctx context.Context, opts ListTenantsOptions) ([]Tenant, error) {
	base := &openapi.ListTenantsParams{Search: opts.Search}
	if opts.Key != nil {
		key := openapi.V2TenantsGetParametersKey(*opts.Key)
		base.Key = &key
	}
	if opts.Order != nil {
		order := openapi.V2TenantsGetParametersOrder(*opts.Order)
		base.Order = &order
	}
	return listAllByPageToken(ctx, func(ctx context.Context, pageToken *string) ([]Tenant, *string, error) {
		params := *base
		params.PageToken = pageToken
		return fetchPageToken[Tenant](c, func() (*http.Response, error) {
			return c.api.ListTenants(ctx, &params)
		})
	})
}

// TenantDeploymentCapability is a deploy-to target for a tenant (Beta).
type TenantDeploymentCapability struct {
	TenantOrganizationID string `json:"tenantOrganizationId"`
}

func (c *Client) ListTenantDeploymentCapabilities(ctx context.Context, tenantOrganizationID string) ([]TenantDeploymentCapability, error) {
	return listAllByPageToken(ctx, func(ctx context.Context, pageToken *string) ([]TenantDeploymentCapability, *string, error) {
		return fetchPageToken[TenantDeploymentCapability](c, func() (*http.Response, error) {
			return c.api.ListTenantDeploymentCapabilities(ctx, tenantOrganizationID, &openapi.ListTenantDeploymentCapabilitiesParams{PageToken: pageToken})
		})
	})
}

func (c *Client) AddTenantDeploymentCapabilities(ctx context.Context, tenantOrganizationID string, deployTo []string) error {
	body, err := encodeBody(map[string]any{"deployToTenantOrganizationIds": deployTo})
	if err != nil {
		return err
	}
	return c.doVoid(func() (*http.Response, error) {
		return c.api.AddTenantDeploymentCapabilitiesWithBody(ctx, tenantOrganizationID, nil, jsonContentType, body)
	})
}

func (c *Client) RemoveTenantDeploymentCapabilities(ctx context.Context, tenantOrganizationID string, deployTo []string) error {
	body, err := encodeBody(map[string]any{"deployToTenantOrganizationIds": deployTo})
	if err != nil {
		return err
	}
	return c.doVoid(func() (*http.Response, error) {
		return c.api.RemoveTenantDeploymentCapabilitiesWithBody(ctx, tenantOrganizationID, nil, jsonContentType, body)
	})
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
	body, err := encodeBody(in)
	if err != nil {
		return nil, err
	}
	var created createDeploymentPolicyResponse
	if err := c.doDecode(func() (*http.Response, error) {
		return c.api.CreateDeploymentWithBody(ctx, nil, jsonContentType, body)
	}, &created); err != nil {
		return nil, err
	}
	return c.GetDeploymentPolicy(ctx, created.DeploymentPolicyID)
}

func (c *Client) GetDeploymentPolicy(ctx context.Context, id string) (*DeploymentPolicy, error) {
	var value DeploymentPolicy
	err := c.doDecode(func() (*http.Response, error) {
		return c.api.GetDeployment(ctx, id, nil)
	}, &value)
	return &value, err
}

func (c *Client) UpdateDeploymentPolicy(ctx context.Context, id string, in UpdateDeploymentPolicyInput) (*DeploymentPolicy, error) {
	body, err := encodeBody(in)
	if err != nil {
		return nil, err
	}
	var value DeploymentPolicy
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.UpdateDeploymentWithBody(ctx, id, nil, jsonContentType, body)
	}, &value)
	return &value, err
}

func (c *Client) ArchiveDeploymentPolicy(ctx context.Context, id string) error {
	return c.doVoid(func() (*http.Response, error) {
		return c.api.ArchiveDeployment(ctx, id, nil)
	})
}

func (c *Client) ListDeploymentPolicies(ctx context.Context) ([]DeploymentPolicy, error) {
	return listAllByPageToken(ctx, func(ctx context.Context, pageToken *string) ([]DeploymentPolicy, *string, error) {
		return fetchPageToken[DeploymentPolicy](c, func() (*http.Response, error) {
			return c.api.ListDeployments(ctx, &openapi.ListDeploymentsParams{PageToken: pageToken})
		})
	})
}

func (c *Client) ListDeploymentPolicyInodes(ctx context.Context, deploymentPolicyID string) ([]DeploymentPolicyFile, error) {
	return listAllByPageToken(ctx, func(ctx context.Context, pageToken *string) ([]DeploymentPolicyFile, *string, error) {
		return fetchPageToken[DeploymentPolicyFile](c, func() (*http.Response, error) {
			return c.api.ListInodesForDeployment(ctx, deploymentPolicyID, &openapi.ListInodesForDeploymentParams{PageToken: pageToken})
		})
	})
}

func (c *Client) AddDeploymentPolicyInodes(ctx context.Context, deploymentPolicyID string, inodeIDs []string) error {
	body, err := encodeBody(map[string]any{"inodeIds": inodeIDs})
	if err != nil {
		return err
	}
	return c.doVoid(func() (*http.Response, error) {
		return c.api.AddInodesToDeploymentWithBody(ctx, deploymentPolicyID, nil, jsonContentType, body)
	})
}

func (c *Client) RemoveDeploymentPolicyInode(ctx context.Context, deploymentPolicyID, inodeID string) error {
	return c.doVoid(func() (*http.Response, error) {
		return c.api.RemoveInodesFromDeployment(ctx, deploymentPolicyID, inodeID, nil)
	})
}

func (c *Client) ListDeploymentPolicyTenants(ctx context.Context, deploymentPolicyID string) ([]string, error) {
	return listAllByPageToken(ctx, func(ctx context.Context, pageToken *string) ([]string, *string, error) {
		return fetchPageToken[string](c, func() (*http.Response, error) {
			return c.api.ListTenantsForDeployment(ctx, deploymentPolicyID, &openapi.ListTenantsForDeploymentParams{PageToken: pageToken})
		})
	})
}

func (c *Client) AddDeploymentPolicyTenant(ctx context.Context, deploymentPolicyID, tenantOrganizationID string) error {
	body, err := encodeBody(map[string]any{"tenantOrganizationId": tenantOrganizationID})
	if err != nil {
		return err
	}
	return c.doVoid(func() (*http.Response, error) {
		return c.api.AddTenantToDeploymentWithBody(ctx, deploymentPolicyID, nil, jsonContentType, body)
	})
}

func (c *Client) RemoveDeploymentPolicyTenant(ctx context.Context, deploymentPolicyID, tenantOrganizationID string) error {
	return c.doVoid(func() (*http.Response, error) {
		return c.api.RemoveTenantFromDeployment(ctx, deploymentPolicyID, tenantOrganizationID, nil)
	})
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
	body, err := encodeBody(in)
	if err != nil {
		return nil, err
	}
	var created createSourceSwapPolicyResponse
	if err := c.doDecode(func() (*http.Response, error) {
		return c.api.CreateSourceSwapPolicyWithBody(ctx, nil, jsonContentType, body)
	}, &created); err != nil {
		return nil, err
	}
	return c.GetSourceSwapPolicy(ctx, created.PolicyID)
}

func (c *Client) GetSourceSwapPolicy(ctx context.Context, id string) (*SourceSwapPolicy, error) {
	var value SourceSwapPolicy
	err := c.doDecode(func() (*http.Response, error) {
		return c.api.GetSourceSwapPolicy(ctx, id, nil)
	}, &value)
	return &value, err
}

func (c *Client) UpdateSourceSwapPolicy(ctx context.Context, id string, in SourceSwapPolicyInput) (*SourceSwapPolicy, error) {
	body, err := encodeBody(in)
	if err != nil {
		return nil, err
	}
	var created createSourceSwapPolicyResponse
	if err := c.doDecode(func() (*http.Response, error) {
		return c.api.UpdateSourceSwapPolicyWithBody(ctx, id, nil, jsonContentType, body)
	}, &created); err != nil {
		return nil, err
	}
	return c.GetSourceSwapPolicy(ctx, id)
}

func (c *Client) DeleteSourceSwapPolicy(ctx context.Context, id string) error {
	return c.doVoid(func() (*http.Response, error) {
		return c.api.DeleteSourceSwapPolicy(ctx, id, nil)
	})
}
