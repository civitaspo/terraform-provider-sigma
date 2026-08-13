package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccountTypesDataSource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	accountType := map[string]any{
		"accountTypeId": "type-1", "accountTypeName": "Analyst", "description": "Analyst access", "isCustom": true,
	}
	write := func(response http.ResponseWriter, value any) {
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(value)
	}
	mock.Mux.HandleFunc("/v2/accountTypes", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		write(response, map[string]any{"entries": []any{accountType}})
	})
	config := identityProviderConfig(mock) + `
data "sigma_account_types" "all" {}
`
	resource.UnitTest(t, identityTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.sigma_account_types.all", "account_types.#", "1"),
		),
	}}))
}
