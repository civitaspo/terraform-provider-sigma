package sigma

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// Tag is a Sigma version tag.
type Tag struct {
	VersionTagID string  `json:"versionTagId"`
	Name         string  `json:"name"`
	OwnerID      string  `json:"ownerId"`
	CreatedBy    string  `json:"createdBy"`
	UpdatedBy    string  `json:"updatedBy"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
	IsArchived   bool    `json:"isArchived"`
	Description  *string `json:"description"`
	Color        string  `json:"color"`
}

type CreateTagInput struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description,omitempty"`
}

type UpdateTagInput struct {
	Description string `json:"description"`
}

func (c *Client) CreateTag(ctx context.Context, in CreateTagInput) (*Tag, error) {
	var value Tag
	if err := c.sendJSON(ctx, http.MethodPost, "/v2/tags", in, &value); err != nil {
		return nil, err
	}
	return &value, nil
}

func (c *Client) ListTags(ctx context.Context) ([]Tag, error) {
	return ListAll[Tag](ctx, c, "/v2/tags")
}

func (c *Client) GetTag(ctx context.Context, id string) (*Tag, error) {
	tags, err := c.ListTags(ctx)
	if err != nil {
		return nil, err
	}
	for i := range tags {
		if tags[i].VersionTagID == id {
			return &tags[i], nil
		}
	}
	return nil, &APIError{StatusCode: http.StatusNotFound, Message: fmt.Sprintf("tag %q not found", id)}
}

func (c *Client) UpdateTag(ctx context.Context, id string, in UpdateTagInput) (*Tag, error) {
	var value Tag
	err := c.sendJSON(ctx, http.MethodPatch, "/v2/tags/"+url.PathEscape(id), in, &value)
	return &value, err
}

func (c *Client) DeleteTag(ctx context.Context, id string) error {
	return c.sendJSON(ctx, http.MethodDelete, "/v2/tags/"+url.PathEscape(id), nil, nil)
}

// WorkbookSchedule is a scheduled workbook export.
type WorkbookSchedule struct {
	ScheduledNotificationID string          `json:"scheduledNotificationId"`
	WorkbookID              string          `json:"workbookId"`
	Schedule                json.RawMessage `json:"schedule"`
	ConfigV2                json.RawMessage `json:"configV2"`
	IsSuspended             bool            `json:"isSuspended"`
	OwnerID                 string          `json:"ownerId"`
	CreatedBy               string          `json:"createdBy"`
	UpdatedBy               string          `json:"updatedBy"`
	CreatedAt               string          `json:"createdAt"`
	UpdatedAt               string          `json:"updatedAt"`
}

func (c *Client) CreateWorkbookSchedule(ctx context.Context, workbookID string, body json.RawMessage) (*WorkbookSchedule, error) {
	var value WorkbookSchedule
	err := c.sendJSON(ctx, http.MethodPost, "/v2/workbooks/"+url.PathEscape(workbookID)+"/schedules", body, &value)
	return &value, err
}

func (c *Client) ListWorkbookSchedules(ctx context.Context, workbookID string) ([]WorkbookSchedule, error) {
	return ListAll[WorkbookSchedule](ctx, c, "/v2.1/workbooks/"+url.PathEscape(workbookID)+"/schedules")
}

func (c *Client) GetWorkbookSchedule(ctx context.Context, workbookID, scheduleID string) (*WorkbookSchedule, error) {
	values, err := c.ListWorkbookSchedules(ctx, workbookID)
	if err != nil {
		return nil, err
	}
	for i := range values {
		if values[i].ScheduledNotificationID == scheduleID {
			return &values[i], nil
		}
	}
	return nil, &APIError{StatusCode: http.StatusNotFound, Message: fmt.Sprintf("workbook schedule %q not found", scheduleID)}
}

func (c *Client) UpdateWorkbookSchedule(ctx context.Context, workbookID, scheduleID string, body json.RawMessage) (*WorkbookSchedule, error) {
	var value WorkbookSchedule
	path := "/v2/workbooks/" + url.PathEscape(workbookID) + "/schedules/" + url.PathEscape(scheduleID)
	err := c.sendJSON(ctx, http.MethodPatch, path, body, &value)
	return &value, err
}

func (c *Client) DeleteWorkbookSchedule(ctx context.Context, workbookID, scheduleID string) error {
	path := "/v2/workbooks/" + url.PathEscape(workbookID) + "/schedules/" + url.PathEscape(scheduleID)
	return c.sendJSON(ctx, http.MethodDelete, path, nil, nil)
}

// ReportSchedule is a scheduled report export.
type ReportSchedule struct {
	ScheduledNotificationID string          `json:"scheduledNotificationId"`
	ReportID                string          `json:"reportId"`
	Schedule                json.RawMessage `json:"schedule"`
	ConfigV2                json.RawMessage `json:"configV2"`
	IsSuspended             bool            `json:"isSuspended"`
	OwnerID                 string          `json:"ownerId"`
	CreatedBy               string          `json:"createdBy"`
	UpdatedBy               string          `json:"updatedBy"`
	CreatedAt               string          `json:"createdAt"`
	UpdatedAt               string          `json:"updatedAt"`
}

func (c *Client) CreateReportSchedule(ctx context.Context, reportID string, body json.RawMessage) (*ReportSchedule, error) {
	var value ReportSchedule
	err := c.sendJSON(ctx, http.MethodPost, "/v2/reports/"+url.PathEscape(reportID)+"/schedules", body, &value)
	return &value, err
}

func (c *Client) ListReportSchedules(ctx context.Context, reportID string) ([]ReportSchedule, error) {
	return ListAll[ReportSchedule](ctx, c, "/v2/reports/"+url.PathEscape(reportID)+"/schedules")
}

func (c *Client) GetReportSchedule(ctx context.Context, reportID, scheduleID string) (*ReportSchedule, error) {
	values, err := c.ListReportSchedules(ctx, reportID)
	if err != nil {
		return nil, err
	}
	for i := range values {
		if values[i].ScheduledNotificationID == scheduleID {
			return &values[i], nil
		}
	}
	return nil, &APIError{StatusCode: http.StatusNotFound, Message: fmt.Sprintf("report schedule %q not found", scheduleID)}
}

func (c *Client) UpdateReportSchedule(ctx context.Context, reportID, scheduleID string, body json.RawMessage) (*ReportSchedule, error) {
	var value ReportSchedule
	path := "/v2/reports/" + url.PathEscape(reportID) + "/schedules/" + url.PathEscape(scheduleID)
	err := c.sendJSON(ctx, http.MethodPatch, path, body, &value)
	return &value, err
}

func (c *Client) DeleteReportSchedule(ctx context.Context, reportID, scheduleID string) error {
	path := "/v2/reports/" + url.PathEscape(reportID) + "/schedules/" + url.PathEscape(scheduleID)
	return c.sendJSON(ctx, http.MethodDelete, path, nil, nil)
}

// WorkbookEmbed is a workbook embed URL configuration.
type WorkbookEmbed struct {
	EmbedID    string  `json:"embedId"`
	EmbedURL   string  `json:"embedUrl"`
	Public     bool    `json:"public"`
	SourceType string  `json:"sourceType"`
	SourceID   *string `json:"sourceId"`
	SourceName *string `json:"sourceName"`
}

type CreateWorkbookEmbedInput struct {
	EmbedType  string `json:"embedType"`
	SourceType string `json:"sourceType"`
	SourceID   string `json:"sourceId,omitempty"`
}

func (c *Client) CreateWorkbookEmbed(ctx context.Context, workbookID string, in CreateWorkbookEmbedInput) (*WorkbookEmbed, error) {
	var value WorkbookEmbed
	err := c.sendJSON(ctx, http.MethodPost, "/v2/workbooks/"+url.PathEscape(workbookID)+"/embeds", in, &value)
	return &value, err
}

func (c *Client) ListWorkbookEmbeds(ctx context.Context, workbookID string) ([]WorkbookEmbed, error) {
	return ListAll[WorkbookEmbed](ctx, c, "/v2/workbooks/"+url.PathEscape(workbookID)+"/embeds")
}

func (c *Client) GetWorkbookEmbed(ctx context.Context, workbookID, embedID string) (*WorkbookEmbed, error) {
	values, err := c.ListWorkbookEmbeds(ctx, workbookID)
	if err != nil {
		return nil, err
	}
	for i := range values {
		if values[i].EmbedID == embedID {
			return &values[i], nil
		}
	}
	return nil, &APIError{StatusCode: http.StatusNotFound, Message: fmt.Sprintf("workbook embed %q not found", embedID)}
}

func (c *Client) DeleteWorkbookEmbed(ctx context.Context, workbookID, embedID string) error {
	path := "/v2/workbooks/" + url.PathEscape(workbookID) + "/embeds/" + url.PathEscape(embedID)
	return c.sendJSON(ctx, http.MethodDelete, path, nil, nil)
}

// OrgTranslation is an organization translation file.
type OrgTranslation struct {
	Translations map[string]string `json:"translations"`
}

type CreateOrgTranslationInput struct {
	Lng          string            `json:"lng"`
	LngVariant   string            `json:"lng_variant,omitempty"`
	Translations map[string]string `json:"translations,omitempty"`
}

type UpdateOrgTranslationInput struct {
	Translations map[string]string `json:"translations"`
}

func orgTranslationPath(lng, variant string) string {
	path := "/v2/translations/organization/" + url.PathEscape(lng)
	if variant != "" {
		path += "/" + url.PathEscape(variant)
	}
	return path
}

func (c *Client) CreateOrgTranslation(ctx context.Context, in CreateOrgTranslationInput) error {
	return c.sendJSON(ctx, http.MethodPost, "/v2/translations/organization", in, nil)
}

func (c *Client) GetOrgTranslation(ctx context.Context, lng, variant string) (*OrgTranslation, error) {
	var value OrgTranslation
	err := c.getJSON(ctx, orgTranslationPath(lng, variant), &value)
	return &value, err
}

func (c *Client) UpdateOrgTranslation(ctx context.Context, lng, variant string, in UpdateOrgTranslationInput) error {
	return c.sendJSON(ctx, http.MethodPut, orgTranslationPath(lng, variant), in, nil)
}

func (c *Client) DeleteOrgTranslation(ctx context.Context, lng, variant string) error {
	return c.sendJSON(ctx, http.MethodDelete, orgTranslationPath(lng, variant), nil, nil)
}

// Workbook is a Sigma workbook document.
type Workbook struct {
	WorkbookID    string  `json:"workbookId"`
	WorkbookURLID string  `json:"workbookUrlId"`
	Name          string  `json:"name"`
	URL           string  `json:"url"`
	Path          string  `json:"path"`
	LatestVersion float64 `json:"latestVersion"`
	OwnerID       string  `json:"ownerId"`
	CreatedBy     string  `json:"createdBy"`
	UpdatedBy     string  `json:"updatedBy"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
	IsArchived    bool    `json:"isArchived"`
}

func (c *Client) GetWorkbook(ctx context.Context, id string) (*Workbook, error) {
	var value Workbook
	err := c.getJSON(ctx, "/v2/workbooks/"+url.PathEscape(id), &value)
	return &value, err
}

func (c *Client) ListWorkbooks(ctx context.Context) ([]Workbook, error) {
	return ListAll[Workbook](ctx, c, "/v2/workbooks")
}

// Report is a Sigma report document.
type Report struct {
	ReportID      string  `json:"reportId"`
	ReportURLID   string  `json:"reportUrlId"`
	Name          string  `json:"name"`
	URL           string  `json:"url"`
	Path          string  `json:"path"`
	LatestVersion float64 `json:"latestVersion"`
	OwnerID       string  `json:"ownerId"`
	CreatedBy     string  `json:"createdBy"`
	UpdatedBy     string  `json:"updatedBy"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
	IsArchived    bool    `json:"isArchived"`
}

func (c *Client) GetReport(ctx context.Context, id string) (*Report, error) {
	var value Report
	err := c.getJSON(ctx, "/v2/reports/"+url.PathEscape(id), &value)
	return &value, err
}

func (c *Client) ListReports(ctx context.Context) ([]Report, error) {
	return ListAll[Report](ctx, c, "/v2/reports")
}

// DataModel is a Sigma data model document.
type DataModel struct {
	DataModelID    string  `json:"dataModelId"`
	DataModelURLID string  `json:"dataModelUrlId"`
	Name           string  `json:"name"`
	URL            string  `json:"url"`
	Path           string  `json:"path"`
	LatestVersion  float64 `json:"latestVersion"`
	OwnerID        string  `json:"ownerId"`
	CreatedBy      string  `json:"createdBy"`
	UpdatedBy      string  `json:"updatedBy"`
	CreatedAt      string  `json:"createdAt"`
	UpdatedAt      string  `json:"updatedAt"`
	IsArchived     bool    `json:"isArchived"`
}

func (c *Client) GetDataModel(ctx context.Context, id string) (*DataModel, error) {
	var value DataModel
	err := c.getJSON(ctx, "/v2/dataModels/"+url.PathEscape(id), &value)
	return &value, err
}

func (c *Client) ListDataModels(ctx context.Context) ([]DataModel, error) {
	return ListAll[DataModel](ctx, c, "/v2/dataModels")
}

// Dataset is a deprecated Sigma dataset document.
type Dataset struct {
	DatasetID   string  `json:"datasetId"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	URL         string  `json:"url"`
	Path        string  `json:"path"`
	Owner       string  `json:"owner"`
	CreatedBy   string  `json:"createdBy"`
	UpdatedBy   string  `json:"updatedBy"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
	IsArchived  bool    `json:"isArchived"`
}

func (c *Client) GetDataset(ctx context.Context, id string) (*Dataset, error) {
	var value Dataset
	err := c.getJSON(ctx, "/v2/datasets/"+url.PathEscape(id), &value)
	return &value, err
}

func (c *Client) ListDatasets(ctx context.Context) ([]Dataset, error) {
	return ListAll[Dataset](ctx, c, "/v2/datasets")
}

// Template is a Sigma template document.
type Template struct {
	TemplateID    string  `json:"templateId"`
	TemplateURLID string  `json:"templateUrlId"`
	Name          string  `json:"name"`
	URL           string  `json:"url"`
	Path          string  `json:"path"`
	LatestVersion float64 `json:"latestVersion"`
	CreatedBy     string  `json:"createdBy"`
	UpdatedBy     string  `json:"updatedBy"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
	IsArchived    bool    `json:"isArchived"`
}

func (c *Client) ListTemplates(ctx context.Context) ([]Template, error) {
	return ListAll[Template](ctx, c, "/v2/templates")
}
