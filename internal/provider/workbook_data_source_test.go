package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestWorkbookDataSource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	workbook := map[string]any{
		"workbookId": "workbook-1", "workbookUrlId": "wb-url", "name": "Revenue", "url": "/workbook/wb-url",
		"path": "Analytics/Revenue", "latestVersion": 3, "ownerId": "member-1",
		"createdBy": "member-1", "updatedBy": "member-1",
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
	config := documentProviderConfig(mock) + `
data "sigma_workbook" "one" {
  id = "workbook-1"
}
`
	resource.UnitTest(t, documentTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.sigma_workbook.one", "name", "Revenue"),
		),
	}}))
}
