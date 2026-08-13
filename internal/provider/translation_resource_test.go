package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestTranslationResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	translations := map[string]string{"Hello": "Bonjour"}
	mock.Mux.HandleFunc("/v2/translations/organization", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{})
	})
	mock.Mux.HandleFunc("/v2/translations/organization/fr", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(map[string]any{"translations": translations})
		case http.MethodPut:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		case http.MethodDelete:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	config := documentProviderConfig(mock) + `
resource "sigma_translation" "test" {
  lng = "fr"
  translations = {
    Hello = "Bonjour"
  }
}
`
	resource.UnitTest(t, documentTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("sigma_translation.test", "id", "fr"),
			resource.TestCheckResourceAttr("sigma_translation.test", "translations.Hello", "Bonjour"),
		),
	}}))
}

func TestAccTranslationResource(t *testing.T) { requireAcceptance(t) }
