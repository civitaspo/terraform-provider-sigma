package provider_test

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccountTypePermissionsDataSource(t *testing.T) {
	config := `
data "sigma_account_type_permissions" "all" {
  account_type_id = "type-1"
}
`
	t.Run("read", func(t *testing.T) {
		mock := testutil.NewMockSigma(t)
		mock.Mux.HandleFunc("/v2/accountTypes/type-1/permissions", func(response http.ResponseWriter, request *http.Request) {
			mock.AssertBearer(t, request)
			if request.URL.Path != "/v2/accountTypes/type-1/permissions" {
				t.Errorf("path = %q", request.URL.Path)
			}
			if request.Method != http.MethodGet {
				t.Errorf("method = %s", request.Method)
			}
			assertExactQuery(t, request, map[string]string{})
			writeJSON(response, []any{map[string]any{"permission": "view-worksheet", "description": "View worksheets"}})
		})
		resource.UnitTest(t, identityTestCase([]resource.TestStep{{
			Config: providerConfig(mock) + config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.sigma_account_type_permissions.all", "permissions.#", "1"),
				resource.TestCheckResourceAttr("data.sigma_account_type_permissions.all", "permissions.0.permission", "view-worksheet"),
			),
		}}))
	})
	t.Run("empty", func(t *testing.T) {
		mock := testutil.NewMockSigma(t)
		mock.Mux.HandleFunc("/v2/accountTypes/type-1/permissions", func(response http.ResponseWriter, request *http.Request) {
			mock.AssertBearer(t, request)
			writeJSON(response, []any{})
		})
		resource.UnitTest(t, identityTestCase([]resource.TestStep{{
			Config: providerConfig(mock) + config,
			Check:  resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr("data.sigma_account_type_permissions.all", "permissions.#", "0")),
		}}))
	})
	t.Run("error", func(t *testing.T) {
		mock := testutil.NewMockSigma(t)
		mock.Mux.HandleFunc("/v2/accountTypes/type-1/permissions", func(response http.ResponseWriter, request *http.Request) {
			mock.AssertBearer(t, request)
			writeAPIError(response, http.StatusInternalServerError, "boom")
		})
		resource.UnitTest(t, identityTestCase([]resource.TestStep{{
			Config:      providerConfig(mock) + config,
			ExpectError: regexp.MustCompile(`boom|Unable to list|500`),
		}}))
	})
}

func TestAccAccountTypePermissionsDataSource(t *testing.T) { runAccAccountType(t) }
