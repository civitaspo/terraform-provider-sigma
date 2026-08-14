package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func reportScheduleFixture(cron string, suspended bool) map[string]any {
	return map[string]any{
		"scheduledNotificationId": "schedule-1", "reportId": "report-1",
		"schedule": map[string]any{"cronSpec": cron}, "configV2": map[string]any{"title": "Weekly"},
		"isSuspended": suspended, "ownerId": "owner-1", "createdBy": "owner-1", "updatedBy": "owner-1",
		"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
	}
}

func TestReportScheduleResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	schedule := reportScheduleFixture("0 9 * * 1", false)
	mock.Mux.HandleFunc("/v2/reports/report-1/schedules", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodPost:
			writeJSON(response, schedule)
		case http.MethodGet:
			writeJSON(response, map[string]any{"entries": []any{schedule}, "nextPage": nil})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/reports/report-1/schedules/schedule-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodPatch:
			schedule["isSuspended"] = true
			writeJSON(response, schedule)
		case http.MethodDelete:
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	config := documentProviderConfig(mock) + `
resource "sigma_report_schedule" "test" {
  report_id = "report-1"
  config_json = jsonencode({
    target   = [{ type = "email", recipient = "user@example.com" }]
    schedule = { cronSpec = "0 9 * * 1" }
    configV2 = { title = "Weekly" }
  })
}
`
	resource.UnitTest(t, documentTestCase([]resource.TestStep{
		{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_report_schedule.test", "id", "schedule-1"),
				resource.TestCheckResourceAttr("sigma_report_schedule.test", "is_suspended", "false"),
				resource.TestCheckResourceAttrWith("sigma_report_schedule.test", "config_json", func(value string) error {
					var object map[string]any
					if err := json.Unmarshal([]byte(value), &object); err != nil {
						return err
					}
					if _, ok := object["target"]; !ok {
						return fmt.Errorf("target missing from merged config_json")
					}
					return nil
				}),
			),
		},
		{
			Config: documentProviderConfig(mock) + `
resource "sigma_report_schedule" "test" {
  report_id     = "report-1"
  is_suspended  = true
  config_json = jsonencode({
    target   = [{ type = "email", recipient = "user@example.com" }]
    schedule = { cronSpec = "0 9 * * 1" }
    configV2 = { title = "Weekly" }
  })
}
`,
			Check: resource.TestCheckResourceAttr("sigma_report_schedule.test", "is_suspended", "true"),
		},
	}))
}

func TestReportScheduleResourceImportUnsupported(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	schedule := reportScheduleFixture("0 9 * * 1", false)
	mock.Mux.HandleFunc("/v2/reports/report-1/schedules", func(response http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodPost:
			writeJSON(response, schedule)
		case http.MethodGet:
			writeJSON(response, map[string]any{"entries": []any{schedule}, "nextPage": nil})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/reports/report-1/schedules/schedule-1", func(response http.ResponseWriter, _ *http.Request) {
		writeJSON(response, map[string]any{})
	})
	config := documentProviderConfig(mock) + `
resource "sigma_report_schedule" "test" {
  report_id = "report-1"
  config_json = jsonencode({
    target   = [{ type = "email", recipient = "user@example.com" }]
    schedule = { cronSpec = "0 9 * * 1" }
    configV2 = { title = "Weekly" }
  })
}
`
	resource.UnitTest(t, documentTestCase([]resource.TestStep{
		{Config: config},
		{
			ResourceName:  "sigma_report_schedule.test",
			ImportState:   true,
			ImportStateId: "report-1/schedule-1",
			ExpectError:   regexp.MustCompile(`(?i)import`),
		},
	}))
}

func TestAccReportScheduleResource(t *testing.T) { requireAcceptance(t) }
