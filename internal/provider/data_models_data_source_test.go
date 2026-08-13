package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataModelsDataSource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	dataModel := map[string]any{
		"dataModelId": "data-model-1", "dataModelUrlId": "dm-url", "name": "Sales Model", "url": "/data-model/dm-url",
		"path": "Analytics/Sales Model", "latestVersion": 4, "ownerId": "member-1",
		"createdBy": "member-1", "updatedBy": "member-1",
		"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-02T00:00:00Z", "isArchived": false,
	}
	write := func(response http.ResponseWriter, value any) {
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(value)
	}
	mock.Mux.HandleFunc("/v2/dataModels", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		write(response, map[string]any{"entries": []any{dataModel}, "nextPage": nil})
	})
	config := documentProviderConfig(mock) + `
data "sigma_data_models" "all" {}
`
	resource.UnitTest(t, documentTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.sigma_data_models.all", "data_models.#", "1"),
		),
	}}))
}
