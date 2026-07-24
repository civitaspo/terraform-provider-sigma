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

func documentProviderConfig(mock *testutil.MockSigma) string {
	return `
provider "sigma" {
  base_url      = "` + mock.URL() + `"
  client_id     = "` + mock.ClientID + `"
  client_secret = "` + mock.ClientSecret + `"
}
`
}

func documentTestCase(steps []resource.TestStep) resource.TestCase {
	return resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: steps,
	}
}

func TestTagResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	tag := map[string]any{
		"versionTagId": "tag-1", "name": "prod", "color": "cyan", "description": "Production",
		"ownerId": "owner-1", "createdBy": "owner-1", "updatedBy": "owner-1",
		"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z", "isArchived": false,
	}
	mock.Mux.HandleFunc("/v2/tags", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodPost:
			_ = json.NewEncoder(response).Encode(map[string]any{"versionTagId": "tag-1", "name": "prod"})
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(map[string]any{"entries": []any{tag}, "nextPage": nil})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/tags/tag-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodPatch:
			_ = json.NewEncoder(response).Encode(tag)
		case http.MethodDelete:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	config := documentProviderConfig(mock) + `
resource "sigma_tag" "test" {
  name        = "prod"
  color       = "cyan"
  description = "Production"
}
`
	resource.UnitTest(t, documentTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("sigma_tag.test", "id", "tag-1"),
			resource.TestCheckResourceAttr("sigma_tag.test", "color", "cyan"),
		),
	}}))
}

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

func TestReportScheduleResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	schedule := map[string]any{
		"scheduledNotificationId": "schedule-1", "reportId": "report-1",
		"schedule": map[string]any{"cronSpec": "0 9 * * 1"}, "configV2": map[string]any{"title": "Weekly"},
		"isSuspended": false, "ownerId": "owner-1", "createdBy": "owner-1", "updatedBy": "owner-1",
		"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
	}
	mock.Mux.HandleFunc("/v2/reports/report-1/schedules", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodPost:
			_ = json.NewEncoder(response).Encode(schedule)
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(map[string]any{"entries": []any{schedule}, "nextPage": nil})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/reports/report-1/schedules/schedule-1", func(response http.ResponseWriter, request *http.Request) {
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
resource "sigma_report_schedule" "test" {
  report_id = "report-1"
  config_json = jsonencode({
    target   = [{ type = "email", recipient = "user@example.com" }]
    schedule = { cronSpec = "0 9 * * 1" }
    configV2 = { title = "Weekly" }
  })
}
`
	resource.UnitTest(t, documentTestCase([]resource.TestStep{{
		Config: config,
		Check:  resource.TestCheckResourceAttr("sigma_report_schedule.test", "id", "schedule-1"),
	}}))
}

func TestWorkbookEmbedResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	embed := map[string]any{
		"embedId": "embed-1", "embedUrl": "https://app.sigmacomputing.com/embed/1",
		"public": true, "sourceType": "workbook", "sourceId": nil, "sourceName": nil,
	}
	mock.Mux.HandleFunc("/v2/workbooks/workbook-1/embeds", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodPost:
			_ = json.NewEncoder(response).Encode(map[string]any{"embedId": "embed-1", "embedUrl": embed["embedUrl"]})
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(map[string]any{"entries": []any{embed}, "nextPage": nil})
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
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{})
	})
	config := documentProviderConfig(mock) + `
resource "sigma_workbook_embed" "test" {
  workbook_id = "workbook-1"
  embed_type  = "public"
  source_type = "workbook"
}
`
	resource.UnitTest(t, documentTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("sigma_workbook_embed.test", "id", "embed-1"),
			resource.TestCheckResourceAttr("sigma_workbook_embed.test", "embed_url", "https://app.sigmacomputing.com/embed/1"),
		),
	}}))
}

func TestTranslationResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	translations := map[string]string{"Hello": "Bonjour"}
	mock.Mux.HandleFunc("/v2/translations/organization", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		if request.Method != http.MethodPost {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{})
	})
	mock.Mux.HandleFunc("/v2/translations/organization/fr", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(map[string]any{"translations": translations})
		case http.MethodPut:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		case http.MethodDelete:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	config := documentProviderConfig(mock) + `
resource "sigma_translation" "test" {
  lng = "fr"
  translations = {
    Hello = "Bonjour"
  }
}
`
	resource.UnitTest(t, documentTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("sigma_translation.test", "id", "fr"),
			resource.TestCheckResourceAttr("sigma_translation.test", "translations.Hello", "Bonjour"),
		),
	}}))
}
