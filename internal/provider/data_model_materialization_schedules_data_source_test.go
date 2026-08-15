package provider_test

import "testing"

func TestDataModelMaterializationSchedulesDataSource(t *testing.T) {
	runListDataSourceCases(t, listDataSourceCase{
		path: "/v2/dataModels/data-model-1/materializationSchedules",
		config: `
data "sigma_data_model_materialization_schedules" "all" {
  data_model_id = "data-model-1"
}
`,
		wantQuery:  map[string]string{},
		cursorKey:  "pageToken",
		tokenPaged: true,
		entry: map[string]any{
			"sheetId": "sheet-1", "elementId": "element-1", "elementName": "Sales",
			"schedule":     map[string]any{"cronSpec": "0 0 * * *", "timezone": "UTC"},
			"configuredAt": "2026-01-01T00:00:00Z", "paused": false,
		},
		address:   "data.sigma_data_model_materialization_schedules.all",
		countAttr: "schedules.#",
	})
}

func TestAccDataModelMaterializationSchedulesDataSource(t *testing.T) {
	requireAcceptance(t)
	t.Skip("materialization schedule lookup needs a dedicated data model ID")
}
