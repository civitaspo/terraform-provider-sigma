package provider

import (
	"fmt"
	"os"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type configuredResource struct{ client *sigma.Client }

func (r *configuredResource) configure(request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}
	client, ok := request.ProviderData.(*sigma.Client)
	if !ok {
		response.Diagnostics.AddError("Unexpected resource configuration type", "The Sigma provider returned an unexpected client type.")
		return
	}
	r.client = client
}

type configuredDataSource struct{ client *sigma.Client }

func (d *configuredDataSource) configure(request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}
	client, ok := request.ProviderData.(*sigma.Client)
	if !ok {
		response.Diagnostics.AddError("Unexpected data source configuration type", "The Sigma provider returned an unexpected client type.")
		return
	}
	d.client = client
}

func configuredValue(value types.String, environment, attribute string, diagnostics *diag.Diagnostics) string {
	if value.IsUnknown() {
		diagnostics.AddError(
			fmt.Sprintf("Unknown %s", attribute),
			fmt.Sprintf("The %s provider attribute must be known during configuration.", attribute),
		)
		return ""
	}
	if !value.IsNull() && value.ValueString() != "" {
		return value.ValueString()
	}
	if environmentValue := os.Getenv(environment); environmentValue != "" {
		return environmentValue
	}
	diagnostics.AddError(
		fmt.Sprintf("Missing %s", attribute),
		fmt.Sprintf("Set the %s provider attribute or the %s environment variable.", attribute, environment),
	)
	return ""
}
