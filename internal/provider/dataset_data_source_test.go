package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDatasetDataSource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	dataset := map[string]any{
		"datasetId": "dataset-1", "name": "Orders", "description": "Order facts", "url": "/dataset/dataset-1",
		"path": "Analytics/Orders", "owner": "member-1",
		"createdBy": "member-1", "updatedBy": "member-1",
		"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-02T00:00:00Z", "isArchived": false,
	}
	write := func(response http.ResponseWriter, value any) {
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(value)
	}
	mock.Mux.HandleFunc("/v2/datasets/dataset-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		write(response, dataset)
	})
	config := documentProviderConfig(mock) + `
data "sigma_dataset" "one" {
  id = "dataset-1"
}
`
	resource.UnitTest(t, documentTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.sigma_dataset.one", "name", "Orders"),
		),
	}}))
}
