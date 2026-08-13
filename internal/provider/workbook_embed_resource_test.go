package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestWorkbookEmbedResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	embed := map[string]any{
		"embedId": "embed-1", "embedUrl": "https://app.sigmacomputing.com/embed/1",
		"public": true, "sourceType": "workbook", "sourceId": nil, "sourceName": nil,
	}
	mock.Mux.HandleFunc("/v2/workbooks/workbook-1/embeds", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodPost:
			_ = json.NewEncoder(response).Encode(map[string]any{"embedId": "embed-1", "embedUrl": embed["embedUrl"]})
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(map[string]any{"entries": []any{embed}, "nextPage": nil})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/workbooks/workbook-1/embeds/embed-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodDelete {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{})
	})
	config := documentProviderConfig(mock) + `
resource "sigma_workbook_embed" "test" {
  workbook_id = "workbook-1"
  embed_type  = "public"
  source_type = "workbook"
}
`
	resource.UnitTest(t, documentTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("sigma_workbook_embed.test", "id", "embed-1"),
			resource.TestCheckResourceAttr("sigma_workbook_embed.test", "embed_url", "https://app.sigmacomputing.com/embed/1"),
		),
	}}))
}

func TestAccWorkbookEmbedResource(t *testing.T) { requireAcceptance(t) }
