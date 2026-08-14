package provider_test

import "testing"

func TestWorkbookMaterializationSchedulesDataSource(t *testing.T) {
	runListDataSourceCases(t, listDataSourceCase{
		path: "/v2.1/workbooks/workbook-1/materialization-schedules",
		config: `
data "sigma_workbook_materialization_schedules" "all" {
  workbook_id = "workbook-1"
}
`,
		wantQuery: map[string]string{},
		entry: map[string]any{
			"sheetId": "sheet-1", "elementId": "element-1", "elementName": "Sales",
			"schedule":     map[string]any{"cronSpec": "0 0 * * *", "timezone": "UTC"},
			"configuredAt": "2026-01-01T00:00:00Z", "paused": false,
		},
		address:   "data.sigma_workbook_materialization_schedules.all",
		countAttr: "schedules.#",
	})
}

func TestAccWorkbookMaterializationSchedulesDataSource(t *testing.T) { requireAcceptance(t) }
