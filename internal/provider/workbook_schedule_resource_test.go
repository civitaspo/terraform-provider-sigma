package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestWorkbookScheduleResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	schedule := map[string]any{
		"scheduledNotificationId": "schedule-1", "workbookId": "workbook-1",
		"schedule": map[string]any{"cronSpec": "0 9 * * 1"}, "configV2": map[string]any{"title": "Weekly"},
		"isSuspended": false, "ownerId": "owner-1", "createdBy": "owner-1", "updatedBy": "owner-1",
		"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
	}
	mock.Mux.HandleFunc("/v2/workbooks/workbook-1/schedules", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(schedule)
	})
	mock.Mux.HandleFunc("/v2.1/workbooks/workbook-1/schedules", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{"entries": []any{schedule}, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/workbooks/workbook-1/schedules/schedule-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodPatch:
			_ = json.NewEncoder(response).Encode(schedule)
		case http.MethodDelete:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	config := documentProviderConfig(mock) + `
resource "sigma_workbook_schedule" "test" {
  workbook_id = "workbook-1"
  config_json = jsonencode({
    target   = [{ type = "email", recipient = "user@example.com" }]
    schedule = { cronSpec = "0 9 * * 1" }
    configV2 = { title = "Weekly" }
  })
}
`
	resource.UnitTest(t, documentTestCase([]resource.TestStep{{
		Config: config,
		Check:  resource.TestCheckResourceAttr("sigma_workbook_schedule.test", "id", "schedule-1"),
	}}))
}

func TestAccWorkbookScheduleResource(t *testing.T) { requireAcceptance(t) }
