package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestUserAttributesDataSource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	attribute := map[string]any{
		"userAttributeId": "attribute-1", "name": "Region", "description": "Sales region",
		"defaultValue": map[string]string{"val": "global", "type": "string"},
	}
	write := func(response http.ResponseWriter, value any) {
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(value)
	}
	mock.Mux.HandleFunc("/v2/user-attributes", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		write(response, map[string]any{"entries": []any{attribute}, "nextPage": nil})
	})
	config := identityProviderConfig(mock) + `
data "sigma_user_attributes" "all" {}
`
	resource.UnitTest(t, identityTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.sigma_user_attributes.all", "user_attributes.0.name", "Region"),
		),
	}}))
}
