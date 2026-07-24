package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const betaDataSourceNotice = "This resource uses a Sigma Beta API and may change without notice."

type tenantsDataSource struct {
	client *sigma.Client
}

type tenantDataModel struct {
	ID                     types.String `tfsdk:"id"`
	TenantOrganizationName types.String `tfsdk:"tenant_organization_name"`
	TenantOrganizationSlug types.String `tfsdk:"tenant_organization_slug"`
	ParentOrganizationID   types.String `tfsdk:"parent_organization_id"`
	CreatedBy              types.String `tfsdk:"created_by"`
	UpdatedBy              types.String `tfsdk:"updated_by"`
	CreatedAt              types.String `tfsdk:"created_at"`
	UpdatedAt              types.String `tfsdk:"updated_at"`
	SharedAt               types.String `tfsdk:"shared_at"`
	TenantCloudProvider    types.String `tfsdk:"tenant_cloud_provider"`
	TenantRegion           types.String `tfsdk:"tenant_region"`
	TenantAPIURL           types.String `tfsdk:"tenant_api_url"`
}

type tenantsDataModel struct {
	ID      types.String      `tfsdk:"id"`
	Tenants []tenantDataModel `tfsdk:"tenants"`
}

func NewTenantsDataSource() datasource.DataSource { return &tenantsDataSource{} }

func (d *tenantsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenants"
}
func (d *tenantsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sigma.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected data source configuration type", "The Sigma provider returned an unexpected client type.")
		return
	}
	d.client = client
}
func (d *tenantsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Sigma tenant organizations. " + betaDataSourceNotice,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
			"tenants": schema.ListNestedAttribute{
				Computed: true, MarkdownDescription: "Tenant organizations visible to the caller.",
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id":                       schema.StringAttribute{Computed: true, MarkdownDescription: "Tenant organization ID."},
					"tenant_organization_name": schema.StringAttribute{Computed: true, MarkdownDescription: "Display name of the tenant organization."},
					"tenant_organization_slug": schema.StringAttribute{Computed: true, MarkdownDescription: "URL identifier for the tenant organization."},
					"parent_organization_id":   schema.StringAttribute{Computed: true, MarkdownDescription: "Parent organization ID."},
					"created_by":               schema.StringAttribute{Computed: true, MarkdownDescription: "Member ID that created the tenant."},
					"updated_by":               schema.StringAttribute{Computed: true, MarkdownDescription: "Member ID that last updated the tenant."},
					"created_at":               schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp."},
					"updated_at":               schema.StringAttribute{Computed: true, MarkdownDescription: "Last update timestamp."},
					"shared_at":                schema.StringAttribute{Computed: true, MarkdownDescription: "Share timestamp, when applicable."},
					"tenant_cloud_provider":    schema.StringAttribute{Computed: true, MarkdownDescription: "Cloud provider hosting the tenant."},
					"tenant_region":            schema.StringAttribute{Computed: true, MarkdownDescription: "Region hosting the tenant."},
					"tenant_api_url":           schema.StringAttribute{Computed: true, MarkdownDescription: "Tenant organization API base URL."},
				}},
			},
		},
	}
}
func (d *tenantsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state tenantsDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tenants, err := d.client.ListTenants(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma tenants", err.Error())
		return
	}
	state.ID = types.StringValue("tenants")
	state.Tenants = make([]tenantDataModel, 0, len(tenants))
	for _, tenant := range tenants {
		state.Tenants = append(state.Tenants, tenantDataModel{
			ID:                     types.StringValue(tenant.TenantOrganizationID),
			TenantOrganizationName: types.StringValue(tenant.TenantOrganizationName),
			TenantOrganizationSlug: types.StringValue(tenant.TenantOrganizationSlug),
			ParentOrganizationID:   types.StringValue(tenant.ParentOrganizationID),
			CreatedBy:              types.StringValue(tenant.CreatedBy),
			UpdatedBy:              types.StringValue(tenant.UpdatedBy),
			CreatedAt:              types.StringValue(tenant.CreatedAt),
			UpdatedAt:              types.StringValue(tenant.UpdatedAt),
			SharedAt:               stringOrNull(tenant.SharedAt),
			TenantCloudProvider:    stringOrNull(tenant.TenantCloudProvider),
			TenantRegion:           stringOrNull(tenant.TenantRegion),
			TenantAPIURL:           stringOrNull(tenant.TenantAPIURL),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

type deploymentPoliciesDataSource struct {
	client *sigma.Client
}

type deploymentPolicyDataModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	NameInTenant       types.String `tfsdk:"name_in_tenant"`
	VersionTagID       types.String `tfsdk:"version_tag_id"`
	SourceSwapPolicies types.Set    `tfsdk:"source_swap_policies"`
	CopyInputTableData types.Bool   `tfsdk:"copy_input_table_data"`
}

type deploymentPoliciesDataModel struct {
	ID                 types.String                `tfsdk:"id"`
	DeploymentPolicies []deploymentPolicyDataModel `tfsdk:"deployment_policies"`
}

func NewDeploymentPoliciesDataSource() datasource.DataSource { return &deploymentPoliciesDataSource{} }

func (d *deploymentPoliciesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment_policies"
}
func (d *deploymentPoliciesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sigma.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected data source configuration type", "The Sigma provider returned an unexpected client type.")
		return
	}
	d.client = client
}
func (d *deploymentPoliciesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Sigma deployment policies. " + betaDataSourceNotice,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
			"deployment_policies": schema.ListNestedAttribute{
				Computed: true, MarkdownDescription: "Deployment policies visible to the caller.",
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id":             schema.StringAttribute{Computed: true, MarkdownDescription: "Deployment policy ID."},
					"name":           schema.StringAttribute{Computed: true, MarkdownDescription: "Deployment policy name."},
					"name_in_tenant": schema.StringAttribute{Computed: true, MarkdownDescription: "Workspace name created in receiving tenants."},
					"version_tag_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Version tag ID."},
					"source_swap_policies": schema.SetAttribute{
						Computed: true, ElementType: types.StringType,
						MarkdownDescription: "Source swap policy IDs used when deploying documents.",
					},
					"copy_input_table_data": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether input table data is copied when deploying."},
				}},
			},
		},
	}
}
func (d *deploymentPoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state deploymentPoliciesDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	policies, err := d.client.ListDeploymentPolicies(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma deployment policies", err.Error())
		return
	}
	state.ID = types.StringValue("deployment_policies")
	state.DeploymentPolicies = make([]deploymentPolicyDataModel, 0, len(policies))
	for _, policy := range policies {
		item := deploymentPolicyDataModel{
			ID:                 types.StringValue(policy.DeploymentPolicyID),
			Name:               types.StringValue(policy.Name),
			NameInTenant:       types.StringValue(policy.NameInTenant),
			VersionTagID:       stringOrNull(policy.VersionTagID),
			CopyInputTableData: types.BoolValue(policy.CopyInputTableData),
		}
		if policy.SourceSwapPolicies == nil {
			policy.SourceSwapPolicies = []string{}
		}
		item.SourceSwapPolicies, _ = types.SetValueFrom(ctx, types.StringType, policy.SourceSwapPolicies)
		state.DeploymentPolicies = append(state.DeploymentPolicies, item)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
