package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*whoamiDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*whoamiDataSource)(nil)
)

type whoamiDataSource struct {
	client *sigma.Client
}

type whoamiDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	UserID         types.String `tfsdk:"user_id"`
	OrganizationID types.String `tfsdk:"organization_id"`
}

// NewWhoamiDataSource returns the current-user data source.
func NewWhoamiDataSource() datasource.DataSource {
	return &whoamiDataSource{}
}

func (d *whoamiDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_whoami"
}

func (d *whoamiDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Retrieves the identity of the authenticated Sigma API principal.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Stable identifier for this data source, equal to user_id.",
			},
			"user_id": schema.StringAttribute{
				Computed:    true,
				Description: "Sigma user ID returned by the whoami API.",
			},
			"organization_id": schema.StringAttribute{
				Computed:    true,
				Description: "Sigma organization ID returned by the whoami API.",
			},
		},
	}
}

func (d *whoamiDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}
	client, ok := request.ProviderData.(*sigma.Client)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected data source configuration type",
			"The Sigma provider returned an unexpected client type. Please report this issue to the provider maintainers.",
		)
		return
	}
	d.client = client
}

func (d *whoamiDataSource) Read(ctx context.Context, _ datasource.ReadRequest, response *datasource.ReadResponse) {
	if d.client == nil {
		response.Diagnostics.AddError("Sigma client is not configured", "Configure the Sigma provider before reading this data source.")
		return
	}

	identity, err := d.client.Whoami(ctx)
	if err != nil {
		response.Diagnostics.AddError("Unable to read Sigma identity", err.Error())
		return
	}
	state := whoamiDataSourceModel{
		ID:             types.StringValue(identity.UserID),
		UserID:         types.StringValue(identity.UserID),
		OrganizationID: types.StringValue(identity.OrganizationID),
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
