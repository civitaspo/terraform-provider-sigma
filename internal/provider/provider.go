package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = (*SigmaProvider)(nil)

// SigmaProvider implements the Terraform provider for Sigma Computing.
type SigmaProvider struct {
	version string
}

type providerModel struct {
	BaseURL      types.String `tfsdk:"base_url"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

// New returns a provider factory with the supplied release version.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SigmaProvider{version: version}
	}
}

func (p *SigmaProvider) Metadata(_ context.Context, _ provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "sigma"
	response.Version = p.version
}

func (p *SigmaProvider) Schema(_ context.Context, _ provider.SchemaRequest, response *provider.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Configure the Sigma Computing API client.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Optional:    true,
				Description: "Sigma API base URL. May also be set with SIGMA_BASE_URL.",
			},
			"client_id": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Sigma API client ID. May also be set with SIGMA_CLIENT_ID.",
			},
			"client_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Sigma API client secret. May also be set with SIGMA_CLIENT_SECRET.",
			},
		},
	}
}

func (p *SigmaProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	var config providerModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	baseURL := configuredValue(config.BaseURL, "SIGMA_BASE_URL", "base_url", &response.Diagnostics)
	clientID := configuredValue(config.ClientID, "SIGMA_CLIENT_ID", "client_id", &response.Diagnostics)
	clientSecret := configuredValue(config.ClientSecret, "SIGMA_CLIENT_SECRET", "client_secret", &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	client, err := sigma.NewClient(baseURL, clientID, clientSecret, sigma.WithUserAgent("terraform-provider-sigma/"+p.version))
	if err != nil {
		response.Diagnostics.AddError("Invalid Sigma provider configuration", err.Error())
		return
	}
	response.DataSourceData = client
	response.ResourceData = client
}

func (p *SigmaProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewWhoamiDataSource,
		NewMemberDataSource,
		NewMembersDataSource,
		NewTeamDataSource,
		NewTeamsDataSource,
		NewAccountTypesDataSource,
		NewUserAttributesDataSource,
		NewWorkspaceDataSource,
		NewWorkspacesDataSource,
		NewFilesDataSource,
		NewConnectionDataSource,
		NewConnectionsDataSource,
		NewConnectionPathsDataSource,
		NewConnectionPathDataSource,
		NewWorkbookDataSource,
		NewWorkbooksDataSource,
		NewReportDataSource,
		NewReportsDataSource,
		NewDataModelDataSource,
		NewDataModelsDataSource,
		NewDatasetDataSource,
		NewDatasetsDataSource,
		NewTemplatesDataSource,
		NewTagsDataSource,
		NewTenantsDataSource,
		NewDeploymentPoliciesDataSource,
	}
}

func (p *SigmaProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewMemberResource,
		NewTeamResource,
		NewTeamMemberResource,
		NewAccountTypeResource,
		NewUserAttributeResource,
		NewUserAttributeTeamAssignmentResource,
		NewUserAttributeUserAssignmentResource,
		NewUserAttributeTenantAssignmentResource,
		NewWorkspaceResource,
		NewWorkspaceGrantResource,
		NewFolderResource,
		NewWorkbookGrantResource,
		NewReportGrantResource,
		NewConnectionResource,
		NewConnectionGrantResource,
		NewConnectionPathGrantResource,
		NewAPIConnectorResource,
		NewAPICredentialResource,
		NewTagResource,
		NewWorkbookScheduleResource,
		NewReportScheduleResource,
		NewWorkbookEmbedResource,
		NewTranslationResource,
		NewTenantResource,
		NewTenantDeploymentCapabilityResource,
		NewDeploymentPolicyResource,
		NewDeploymentPolicyDocumentResource,
		NewDeploymentPolicyTenantResource,
		NewSourceSwapPolicyResource,
	}
}
