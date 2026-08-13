package provider_test

import (
	"net/http"
	"regexp"
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
		switch request.Method {
		case http.MethodPost:
			writeJSON(response, map[string]any{"embedId": "embed-1", "embedUrl": embed["embedUrl"]})
		case http.MethodGet:
			writeJSON(response, map[string]any{"entries": []any{embed}, "nextPage": nil})
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
		writeJSON(response, map[string]any{})
	})
	config := documentProviderConfig(mock) + `
resource "sigma_workbook_embed" "test" {
  workbook_id = "workbook-1"
  embed_type  = "public"
  source_type = "workbook"
}
`
	resource.UnitTest(t, documentTestCase([]resource.TestStep{
		{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_workbook_embed.test", "id", "embed-1"),
				resource.TestCheckResourceAttr("sigma_workbook_embed.test", "embed_type", "public"),
				resource.TestCheckResourceAttr("sigma_workbook_embed.test", "embed_url", "https://app.sigmacomputing.com/embed/1"),
			),
		},
		{
			ResourceName:      "sigma_workbook_embed.test",
			ImportState:       true,
			ImportStateId:     "workbook-1/embed-1",
			ImportStateVerify: true,
		},
	}))
}

func TestWorkbookEmbedResourceRejectsNonPublic(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	embed := map[string]any{
		"embedId": "embed-1", "embedUrl": "https://app.sigmacomputing.com/embed/1",
		"public": false, "sourceType": "workbook", "sourceId": nil, "sourceName": nil,
	}
	mock.Mux.HandleFunc("/v2/workbooks/workbook-1/embeds", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodPost:
			writeJSON(response, map[string]any{"embedId": "embed-1", "embedUrl": embed["embedUrl"]})
		case http.MethodGet:
			writeJSON(response, map[string]any{"entries": []any{embed}, "nextPage": nil})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	config := documentProviderConfig(mock) + `
resource "sigma_workbook_embed" "test" {
  workbook_id = "workbook-1"
  embed_type  = "public"
  source_type = "workbook"
}
`
	resource.UnitTest(t, documentTestCase([]resource.TestStep{{
		Config:      config,
		ExpectError: regexp.MustCompile(`not public`),
	}}))
}

func TestWorkbookEmbedResourceRejectsSecureType(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	config := documentProviderConfig(mock) + `
resource "sigma_workbook_embed" "test" {
  workbook_id = "workbook-1"
  embed_type  = "secure"
  source_type = "workbook"
}
`
	resource.UnitTest(t, documentTestCase([]resource.TestStep{{
		Config:      config,
		ExpectError: regexp.MustCompile(`must be .public`),
	}}))
}

func TestWorkbookEmbedResourceImportRejectsNonPublic(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	public := map[string]any{
		"embedId": "embed-1", "embedUrl": "https://app.sigmacomputing.com/embed/1",
		"public": true, "sourceType": "workbook", "sourceId": nil, "sourceName": nil,
	}
	private := map[string]any{
		"embedId": "embed-2", "embedUrl": "https://app.sigmacomputing.com/embed/2",
		"public": false, "sourceType": "workbook", "sourceId": nil, "sourceName": nil,
	}
	mock.Mux.HandleFunc("/v2/workbooks/workbook-1/embeds", func(response http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodPost:
			writeJSON(response, map[string]any{"embedId": "embed-1", "embedUrl": public["embedUrl"]})
		case http.MethodGet:
			writeJSON(response, map[string]any{"entries": []any{public, private}, "nextPage": nil})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/workbooks/workbook-1/embeds/embed-1", func(response http.ResponseWriter, _ *http.Request) {
		writeJSON(response, map[string]any{})
	})
	config := documentProviderConfig(mock) + `
resource "sigma_workbook_embed" "test" {
  workbook_id = "workbook-1"
  embed_type  = "public"
  source_type = "workbook"
}
`
	resource.UnitTest(t, documentTestCase([]resource.TestStep{
		{Config: config},
		{
			ResourceName:  "sigma_workbook_embed.test",
			ImportState:   true,
			ImportStateId: "workbook-1/embed-2",
			ExpectError:   regexp.MustCompile(`not public`),
		},
	}))
}

func TestAccWorkbookEmbedResource(t *testing.T) { requireAcceptance(t) }
