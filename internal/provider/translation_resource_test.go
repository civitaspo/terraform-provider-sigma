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
		writeJSON(response, map[string]any{})
	})
	mock.Mux.HandleFunc("/v2/translations/organization/fr", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodGet:
			writeJSON(response, map[string]any{"translations": translations})
		case http.MethodPut:
			writeJSON(response, map[string]any{})
		case http.MethodDelete:
			writeJSON(response, map[string]any{})
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

func TestTranslationResourceEmptyMapIntent(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	mock.Mux.HandleFunc("/v2/translations/organization", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		var payload map[string]any
		_ = json.NewDecoder(request.Body).Decode(&payload)
		if _, ok := payload["translations"]; !ok {
			t.Errorf("create omitted translations: %#v", payload)
		}
		writeJSON(response, map[string]any{})
	})
	mock.Mux.HandleFunc("/v2/translations/organization/fr", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodGet:
			writeJSON(response, map[string]any{})
		case http.MethodPut:
			var payload map[string]any
			_ = json.NewDecoder(request.Body).Decode(&payload)
			if payload["translations"] == nil {
				t.Errorf("update omitted translations: %#v", payload)
			}
			writeJSON(response, map[string]any{})
		case http.MethodDelete:
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	config := documentProviderConfig(mock) + `
resource "sigma_translation" "test" {
  lng          = "fr"
  translations = {}
}
`
	resource.UnitTest(t, documentTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("sigma_translation.test", "id", "fr"),
			resource.TestCheckResourceAttr("sigma_translation.test", "translations.%", "0"),
		),
	}}))
}

func TestAccTranslationResource(t *testing.T) { requireAcceptance(t) }
