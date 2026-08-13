package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*tenantsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*tenantsDataSource)(nil)
)

type tenantsDataSource struct{ configuredDataSource }

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
	d.configure(req, resp)
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
