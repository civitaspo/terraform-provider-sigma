package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sync/atomic"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func workbookScheduleFixture(cron, title string, suspended bool) map[string]any {
	return map[string]any{
		"scheduledNotificationID": "schedule-1", "workbookId": "workbook-1",
		"scheduledNotificationId": "schedule-1",
		"schedule":                map[string]any{"cronSpec": cron},
		"configV2":                map[string]any{"title": title},
		"isSuspended":             suspended, "ownerId": "owner-1", "createdBy": "owner-1", "updatedBy": "owner-1",
		"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
	}
}

func TestWorkbookScheduleResourceTargetPreservationAndJSONEquality(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	schedule := workbookScheduleFixture("0 9 * * 1", "Weekly", false)
	mock.Mux.HandleFunc("/v2/workbooks/workbook-1/schedules", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(response, schedule)
	})
	mock.Mux.HandleFunc("/v2.1/workbooks/workbook-1/schedules", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		writeJSON(response, map[string]any{"entries": []any{schedule}, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/workbooks/workbook-1/schedules/schedule-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodPatch:
			writeJSON(response, schedule)
		case http.MethodDelete:
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	ordered := documentProviderConfig(mock) + `
resource "sigma_workbook_schedule" "test" {
  workbook_id = "workbook-1"
  config_json = jsonencode({
    target   = [{ type = "email", recipient = "user@example.com" }]
    schedule = { cronSpec = "0 9 * * 1" }
    configV2 = { title = "Weekly" }
  })
}
`
	reordered := documentProviderConfig(mock) + `
resource "sigma_workbook_schedule" "test" {
  workbook_id = "workbook-1"
  config_json = <<-EOT
  { "configV2" : { "title" : "Weekly" }, "schedule" : { "cronSpec" : "0 9 * * 1" }, "target" : [{ "type" : "email", "recipient" : "user@example.com" }] }
  EOT
}
`
	resource.UnitTest(t, documentTestCase([]resource.TestStep{
		{
			Config: ordered,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_workbook_schedule.test", "id", "schedule-1"),
				resource.TestCheckResourceAttr("sigma_workbook_schedule.test", "is_suspended", "false"),
				resource.TestCheckResourceAttrWith("sigma_workbook_schedule.test", "config_json", func(value string) error {
					var object map[string]any
					if err := json.Unmarshal([]byte(value), &object); err != nil {
						return err
					}
					if _, ok := object["target"]; !ok {
						return fmt.Errorf("target missing from merged config_json: %s", value)
					}
					return nil
				}),
			),
		},
		{Config: reordered},
	}))
}

func TestWorkbookScheduleResourceExternalDrift(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	current := workbookScheduleFixture("0 9 * * 1", "Weekly", false)
	mock.Mux.HandleFunc("/v2/workbooks/workbook-1/schedules", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		writeJSON(response, current)
	})
	mock.Mux.HandleFunc("/v2.1/workbooks/workbook-1/schedules", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		writeJSON(response, map[string]any{"entries": []any{current}, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/workbooks/workbook-1/schedules/schedule-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodPatch:
			writeJSON(response, current)
		case http.MethodDelete:
			writeJSON(response, map[string]any{})
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
	resource.UnitTest(t, documentTestCase([]resource.TestStep{
		{Config: config},
		{
			PreConfig: func() {
				current = workbookScheduleFixture("0 10 * * 1", "Weekly", false)
			},
			Config:             config,
			PlanOnly:           true,
			ExpectNonEmptyPlan: true,
		},
	}))
}

func TestWorkbookScheduleResourceSuspensionTransition(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	var lastPatch atomic.Value
	schedule := workbookScheduleFixture("0 9 * * 1", "Weekly", false)
	mock.Mux.HandleFunc("/v2/workbooks/workbook-1/schedules", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		writeJSON(response, schedule)
	})
	mock.Mux.HandleFunc("/v2.1/workbooks/workbook-1/schedules", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		writeJSON(response, map[string]any{"entries": []any{schedule}, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/workbooks/workbook-1/schedules/schedule-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodPatch:
			var payload map[string]any
			_ = json.NewDecoder(request.Body).Decode(&payload)
			lastPatch.Store(payload)
			if payload["suspensionAction"] == "pause" {
				schedule["isSuspended"] = true
			}
			writeJSON(response, schedule)
		case http.MethodDelete:
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	base := documentProviderConfig(mock) + `
resource "sigma_workbook_schedule" "test" {
  workbook_id = "workbook-1"
  config_json = jsonencode({
    target   = [{ type = "email", recipient = "user@example.com" }]
    schedule = { cronSpec = "0 9 * * 1" }
    configV2 = { title = "Weekly" }
  })
`
	resource.UnitTest(t, documentTestCase([]resource.TestStep{
		{Config: base + "}\n"},
		{
			Config: base + "  is_suspended = true\n}\n",
			Check:  resource.TestCheckResourceAttr("sigma_workbook_schedule.test", "is_suspended", "true"),
		},
	}))
	payload, _ := lastPatch.Load().(map[string]any)
	if payload["suspensionAction"] != "pause" {
		t.Fatalf("PATCH body = %#v, want suspensionAction=pause", payload)
	}
}

func TestWorkbookScheduleResourceRejectsSuspensionAction(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	config := documentProviderConfig(mock) + `
resource "sigma_workbook_schedule" "test" {
  workbook_id = "workbook-1"
  config_json = jsonencode({
    schedule         = { cronSpec = "0 9 * * 1" }
    suspensionAction = "pause"
  })
}
`
	resource.UnitTest(t, documentTestCase([]resource.TestStep{{
		Config:      config,
		ExpectError: regexp.MustCompile(`suspensionAction`),
	}}))
}

func TestWorkbookScheduleResourceImportUnsupported(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	schedule := workbookScheduleFixture("0 9 * * 1", "Weekly", false)
	mock.Mux.HandleFunc("/v2/workbooks/workbook-1/schedules", func(response http.ResponseWriter, request *http.Request) {
		writeJSON(response, schedule)
	})
	mock.Mux.HandleFunc("/v2.1/workbooks/workbook-1/schedules", func(response http.ResponseWriter, request *http.Request) {
		writeJSON(response, map[string]any{"entries": []any{schedule}, "nextPage": nil})
	})
	mock.Mux.HandleFunc("/v2/workbooks/workbook-1/schedules/schedule-1", func(response http.ResponseWriter, request *http.Request) {
		writeJSON(response, map[string]any{})
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
	resource.UnitTest(t, documentTestCase([]resource.TestStep{
		{Config: config},
		{
			ResourceName:  "sigma_workbook_schedule.test",
			ImportState:   true,
			ImportStateId: "workbook-1/schedule-1",
			ExpectError:   regexp.MustCompile(`(?i)import`),
		},
	}))
}

func TestAccWorkbookScheduleResource(t *testing.T) { requireAcceptance(t) }
