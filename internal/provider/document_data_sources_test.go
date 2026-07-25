package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	sigmaprovider "github.com/civitaspo/terraform-provider-sigma/internal/provider"
	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDocumentDataSources(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	workbook := map[string]any{
		"workbookId": "workbook-1", "workbookUrlId": "wb-url", "name": "Revenue", "url": "/workbook/wb-url",
		"path": "Analytics/Revenue", "latestVersion": 3, "ownerId": "member-1",
		"createdBy": "member-1", "updatedBy": "member-1",
		"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-02T00:00:00Z", "isArchived": false,
	}
	report := map[string]any{
		"reportId": "report-1", "reportUrlId": "rp-url", "name": "Weekly", "url": "/report/rp-url",
		"path": "Analytics/Weekly", "latestVersion": 2, "ownerId": "member-1",
		"createdBy": "member-1", "updatedBy": "member-1",
		"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-02T00:00:00Z", "isArchived": false,
	}
	tag := map[string]any{
		"versionTagId": "tag-1", "name": "prod", "color": "#00ff00", "description": "Production",
		"ownerId": "member-1", "createdBy": "member-1", "updatedBy": "member-1",
		"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-02T00:00:00Z", "isArchived": false,
	}
	write := func(response http.ResponseWriter, value any) {
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(value)
	}

	mock.Mux.HandleFunc("/v2/workbooks/workbook-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		write(response, workbook)
	})
	mock.Mux.HandleFunc("/v2/workbooks", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		write(response, map[string]any{"entries": []any{workbook}, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/reports/report-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		write(response, report)
	})
	mock.Mux.HandleFunc("/v2/reports", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		write(response, map[string]any{"entries": []any{report}, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/tags", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		write(response, map[string]any{"entries": []any{tag}, "nextPage": nil})
	})

	config := `
provider "sigma" {
  base_url      = "` + mock.URL() + `"
  client_id     = "` + mock.ClientID + `"
  client_secret = "` + mock.ClientSecret + `"
}

data "sigma_workbook" "one" {
  id = "workbook-1"
}
data "sigma_workbooks" "all" {}
data "sigma_report" "one" {
  id = "report-1"
}
data "sigma_reports" "all" {}
data "sigma_tags" "all" {}
`
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.sigma_workbook.one", "name", "Revenue"),
				resource.TestCheckResourceAttr("data.sigma_workbooks.all", "workbooks.#", "1"),
				resource.TestCheckResourceAttr("data.sigma_report.one", "name", "Weekly"),
				resource.TestCheckResourceAttr("data.sigma_reports.all", "reports.#", "1"),
				resource.TestCheckResourceAttr("data.sigma_tags.all", "tags.0.name", "prod"),
			),
		}},
	})
}
