package sigma_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
)

func TestDocumentsClientContract(t *testing.T) {
	t.Parallel()
	tag := map[string]any{"versionTagId": "tag-1", "name": "prod", "color": "cyan", "ownerId": "m1", "createdBy": "m1", "updatedBy": "m1", "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z", "isArchived": false}
	schedule := map[string]any{"scheduledNotificationId": "sched-1", "workbookId": "wb-1", "isSuspended": false, "ownerId": "m1", "createdBy": "m1", "updatedBy": "m1", "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z"}
	embed := map[string]any{"embedId": "embed-1", "embedUrl": "https://app.sigmacomputing.com/embed/1", "public": true, "sourceType": "workbook"}
	workbook := map[string]any{"workbookId": "wb-1", "name": "Sales"}
	report := map[string]any{"reportId": "rep-1", "name": "Weekly"}
	model := map[string]any{"dataModelId": "dm-1", "name": "Model"}
	dataset := map[string]any{"datasetId": "ds-1", "name": "Deprecated"}
	template := map[string]any{"templateId": "tpl-1", "name": "Starter"}
	list := func(entry any) map[string]any {
		return map[string]any{"entries": []any{entry}, "nextPage": nil}
	}

	mock := testutil.NewRecordingSigma(t,
		testutil.ExpectedRequest{Method: "POST", Path: "/v2/tags", JSONBody: map[string]any{"name": "prod", "color": "cyan"}, Response: tag},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/tags", Response: list(tag)},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/tags", Response: list(tag)},
		testutil.ExpectedRequest{Method: "PATCH", Path: "/v2/tags/tag-1", JSONBody: map[string]any{"description": "live"}, Response: tag},
		testutil.ExpectedRequest{Method: "DELETE", Path: "/v2/tags/tag-1", Response: map[string]any{}},
		testutil.ExpectedRequest{Method: "POST", Path: "/v2/workbooks/wb-1/schedules", JSONBody: map[string]any{"cron": "0 * * * *"}, Response: schedule},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2.1/workbooks/wb-1/schedules", Response: list(schedule)},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2.1/workbooks/wb-1/schedules", Response: list(schedule)},
		testutil.ExpectedRequest{Method: "PATCH", Path: "/v2/workbooks/wb-1/schedules/sched-1", JSONBody: map[string]any{"isSuspended": true}, Response: schedule},
		testutil.ExpectedRequest{Method: "DELETE", Path: "/v2/workbooks/wb-1/schedules/sched-1", Response: map[string]any{}},
		testutil.ExpectedRequest{Method: "POST", Path: "/v2/reports/rep-1/schedules", JSONBody: map[string]any{"cron": "0 * * * *"}, Response: schedule},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/reports/rep-1/schedules", Response: list(schedule)},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/reports/rep-1/schedules", Response: list(schedule)},
		testutil.ExpectedRequest{Method: "PATCH", Path: "/v2/reports/rep-1/schedules/sched-1", JSONBody: map[string]any{"isSuspended": true}, Response: schedule},
		testutil.ExpectedRequest{Method: "DELETE", Path: "/v2/reports/rep-1/schedules/sched-1", Response: map[string]any{}},
		testutil.ExpectedRequest{Method: "POST", Path: "/v2/workbooks/wb-1/embeds", JSONBody: map[string]any{"embedType": "public", "sourceType": "workbook"}, Response: embed},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/workbooks/wb-1/embeds", Response: list(embed)},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/workbooks/wb-1/embeds", Response: list(embed)},
		testutil.ExpectedRequest{Method: "DELETE", Path: "/v2/workbooks/wb-1/embeds/embed-1", Response: map[string]any{}},
		testutil.ExpectedRequest{Method: "POST", Path: "/v2/translations/organization", JSONBody: map[string]any{"lng": "ja", "translations": map[string]any{"hello": "こんにちは"}}, Response: map[string]any{}},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/translations/organization/ja", Response: map[string]any{"lng": "ja", "translations": map[string]any{"hello": "こんにちは"}}},
		testutil.ExpectedRequest{Method: "PUT", Path: "/v2/translations/organization/ja", JSONBody: map[string]any{"translations": map[string]any{"hello": "やあ"}}, Response: map[string]any{}},
		testutil.ExpectedRequest{Method: "DELETE", Path: "/v2/translations/organization/ja", Response: map[string]any{}},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/workbooks/wb-1", Query: map[string]string{"includeTaggedSourceUrlId": "true"}, Response: workbook},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/workbooks", Query: map[string]string{"excludeTags": "true"}, Response: list(workbook)},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/reports/rep-1", Response: report},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/reports", Response: list(report)},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/dataModels/dm-1", Response: model},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/dataModels", Response: list(model)},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/datasets/ds-1", Response: dataset},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/datasets", Response: list(dataset)},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/templates", Response: list(template)},
	)
	client, err := sigma.NewClient(mock.URL(), mock.ClientID, mock.ClientSecret)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := client.CreateTag(ctx, sigma.CreateTagInput{Name: "prod", Color: "cyan"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListTags(ctx, sigma.ListTagsOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetTag(ctx, "tag-1"); err != nil {
		t.Fatal(err)
	}
	desc := "live"
	if _, err := client.UpdateTag(ctx, "tag-1", sigma.UpdateTagInput{Description: &desc}); err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteTag(ctx, "tag-1"); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{"cron": "0 * * * *"})
	if _, err := client.CreateWorkbookSchedule(ctx, "wb-1", body); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListWorkbookSchedules(ctx, "wb-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetWorkbookSchedule(ctx, "wb-1", "sched-1"); err != nil {
		t.Fatal(err)
	}
	patch, _ := json.Marshal(map[string]any{"isSuspended": true})
	if _, err := client.UpdateWorkbookSchedule(ctx, "wb-1", "sched-1", patch); err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteWorkbookSchedule(ctx, "wb-1", "sched-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateReportSchedule(ctx, "rep-1", body); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListReportSchedules(ctx, "rep-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetReportSchedule(ctx, "rep-1", "sched-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.UpdateReportSchedule(ctx, "rep-1", "sched-1", patch); err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteReportSchedule(ctx, "rep-1", "sched-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateWorkbookEmbed(ctx, "wb-1", sigma.CreateWorkbookEmbedInput{EmbedType: "public", SourceType: "workbook"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListWorkbookEmbeds(ctx, "wb-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetWorkbookEmbed(ctx, "wb-1", "embed-1"); err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteWorkbookEmbed(ctx, "wb-1", "embed-1"); err != nil {
		t.Fatal(err)
	}
	if err := client.CreateOrgTranslation(ctx, sigma.CreateOrgTranslationInput{Lng: "ja", Translations: map[string]string{"hello": "こんにちは"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetOrgTranslation(ctx, "ja", ""); err != nil {
		t.Fatal(err)
	}
	if err := client.UpdateOrgTranslation(ctx, "ja", "", sigma.UpdateOrgTranslationInput{Translations: map[string]string{"hello": "やあ"}}); err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteOrgTranslation(ctx, "ja", ""); err != nil {
		t.Fatal(err)
	}
	include := true
	if _, err := client.GetWorkbook(ctx, "wb-1", &include); err != nil {
		t.Fatal(err)
	}
	exclude := true
	if _, err := client.ListWorkbooks(ctx, sigma.ListWorkbooksOptions{ExcludeTags: &exclude}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetReport(ctx, "rep-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListReports(ctx, sigma.ListReportsOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetDataModel(ctx, "dm-1", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListDataModels(ctx, sigma.ListDataModelsOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetDataset(ctx, "ds-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListDatasets(ctx, sigma.ListDatasetsOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListTemplates(ctx, sigma.ListTemplatesOptions{}); err != nil {
		t.Fatal(err)
	}
}
