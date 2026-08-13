package sigma

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma/openapi"
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
	body, err := encodeBody(in)
	if err != nil {
		return nil, err
	}
	var value Tag
	if err := c.doDecode(func() (*http.Response, error) {
		return c.api.CreateVersionTagWithBody(ctx, nil, jsonContentType, body)
	}, &value); err != nil {
		return nil, err
	}
	return &value, nil
}

func (c *Client) ListTags(ctx context.Context) ([]Tag, error) {
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]Tag, *string, error) {
		return fetchPage[Tag](c, func() (*http.Response, error) {
			return c.api.ListVersionTag(ctx, &openapi.ListVersionTagParams{Page: page})
		})
	})
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
	body, err := encodeBody(in)
	if err != nil {
		return nil, err
	}
	var value Tag
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.UpdateVersionTagWithBody(ctx, id, nil, jsonContentType, body)
	}, &value)
	return &value, err
}

func (c *Client) DeleteTag(ctx context.Context, id string) error {
	return c.doVoid(func() (*http.Response, error) {
		return c.api.DeleteVersionTag(ctx, id, nil)
	})
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
	reader, err := encodeBody(body)
	if err != nil {
		return nil, err
	}
	var value WorkbookSchedule
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.PostWorkbookScheduleWithBody(ctx, workbookID, nil, jsonContentType, reader)
	}, &value)
	return &value, err
}

func (c *Client) ListWorkbookSchedules(ctx context.Context, workbookID string) ([]WorkbookSchedule, error) {
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]WorkbookSchedule, *string, error) {
		return fetchPage[WorkbookSchedule](c, func() (*http.Response, error) {
			return c.api.V21ListWorkbookSchedules(ctx, workbookID, &openapi.V21ListWorkbookSchedulesParams{Page: page})
		})
	})
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
	reader, err := encodeBody(body)
	if err != nil {
		return nil, err
	}
	var value WorkbookSchedule
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.UpdateWorkbookScheduleWithBody(ctx, workbookID, scheduleID, nil, jsonContentType, reader)
	}, &value)
	return &value, err
}

func (c *Client) DeleteWorkbookSchedule(ctx context.Context, workbookID, scheduleID string) error {
	return c.doVoid(func() (*http.Response, error) {
		return c.api.DeleteWorkbookSchedule(ctx, workbookID, scheduleID, nil)
	})
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
	reader, err := encodeBody(body)
	if err != nil {
		return nil, err
	}
	var value ReportSchedule
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.CreateReportScheduleWithBody(ctx, reportID, nil, jsonContentType, reader)
	}, &value)
	return &value, err
}

func (c *Client) ListReportSchedules(ctx context.Context, reportID string) ([]ReportSchedule, error) {
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]ReportSchedule, *string, error) {
		return fetchPage[ReportSchedule](c, func() (*http.Response, error) {
			return c.api.ListReportSchedules(ctx, reportID, &openapi.ListReportSchedulesParams{Page: page})
		})
	})
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
	reader, err := encodeBody(body)
	if err != nil {
		return nil, err
	}
	var value ReportSchedule
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.UpdateReportScheduleWithBody(ctx, reportID, scheduleID, nil, jsonContentType, reader)
	}, &value)
	return &value, err
}

func (c *Client) DeleteReportSchedule(ctx context.Context, reportID, scheduleID string) error {
	return c.doVoid(func() (*http.Response, error) {
		return c.api.DeleteReportSchedule(ctx, reportID, scheduleID, nil)
	})
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
	body, err := encodeBody(in)
	if err != nil {
		return nil, err
	}
	var value WorkbookEmbed
	err = c.doDecode(func() (*http.Response, error) {
		return c.api.CreateWorkbookEmbedWithBody(ctx, workbookID, nil, jsonContentType, body)
	}, &value)
	return &value, err
}

func (c *Client) ListWorkbookEmbeds(ctx context.Context, workbookID string) ([]WorkbookEmbed, error) {
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]WorkbookEmbed, *string, error) {
		return fetchPage[WorkbookEmbed](c, func() (*http.Response, error) {
			return c.api.ListWorkbookEmbeds(ctx, workbookID, &openapi.ListWorkbookEmbedsParams{Page: page})
		})
	})
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
	return c.doVoid(func() (*http.Response, error) {
		return c.api.DeleteWorkbookEmbeds(ctx, workbookID, embedID, nil)
	})
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

func (c *Client) CreateOrgTranslation(ctx context.Context, in CreateOrgTranslationInput) error {
	body, err := encodeBody(in)
	if err != nil {
		return err
	}
	return c.doVoid(func() (*http.Response, error) {
		return c.api.CreateOrgTranslationWithBody(ctx, nil, jsonContentType, body)
	})
}

func (c *Client) GetOrgTranslation(ctx context.Context, lng, variant string) (*OrgTranslation, error) {
	var value OrgTranslation
	var err error
	if variant == "" {
		err = c.doDecode(func() (*http.Response, error) {
			return c.api.GetOrgTranslations(ctx, lng, nil)
		}, &value)
	} else {
		err = c.doDecode(func() (*http.Response, error) {
			return c.api.GetOrgTranslationsWithVariant(ctx, lng, variant, nil)
		}, &value)
	}
	return &value, err
}

func (c *Client) UpdateOrgTranslation(ctx context.Context, lng, variant string, in UpdateOrgTranslationInput) error {
	body, err := encodeBody(in)
	if err != nil {
		return err
	}
	if variant == "" {
		return c.doVoid(func() (*http.Response, error) {
			return c.api.UpdateOrgTranslationWithBody(ctx, lng, nil, jsonContentType, body)
		})
	}
	return c.doVoid(func() (*http.Response, error) {
		return c.api.UpdateOrgTranslationWithVariantWithBody(ctx, lng, variant, nil, jsonContentType, body)
	})
}

func (c *Client) DeleteOrgTranslation(ctx context.Context, lng, variant string) error {
	if variant == "" {
		return c.doVoid(func() (*http.Response, error) {
			return c.api.DeleteOrgTranslation(ctx, lng, nil)
		})
	}
	return c.doVoid(func() (*http.Response, error) {
		return c.api.DeleteOrgTranslationWithVariant(ctx, lng, variant, nil)
	})
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
	err := c.doDecode(func() (*http.Response, error) {
		return c.api.GetWorkbook(ctx, id, nil)
	}, &value)
	return &value, err
}

func (c *Client) ListWorkbooks(ctx context.Context) ([]Workbook, error) {
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]Workbook, *string, error) {
		return fetchPage[Workbook](c, func() (*http.Response, error) {
			return c.api.ListWorkbooks(ctx, &openapi.ListWorkbooksParams{Page: page})
		})
	})
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
	err := c.doDecode(func() (*http.Response, error) {
		return c.api.GetReport(ctx, id, nil)
	}, &value)
	return &value, err
}

func (c *Client) ListReports(ctx context.Context) ([]Report, error) {
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]Report, *string, error) {
		return fetchPage[Report](c, func() (*http.Response, error) {
			return c.api.ListReports(ctx, &openapi.ListReportsParams{Page: page})
		})
	})
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
	err := c.doDecode(func() (*http.Response, error) {
		return c.api.GetDataModel(ctx, id, nil)
	}, &value)
	return &value, err
}

func (c *Client) ListDataModels(ctx context.Context) ([]DataModel, error) {
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]DataModel, *string, error) {
		return fetchPage[DataModel](c, func() (*http.Response, error) {
			return c.api.ListDataModels(ctx, &openapi.ListDataModelsParams{Page: page})
		})
	})
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
	err := c.doDecode(func() (*http.Response, error) {
		return c.api.GetDataset(ctx, id, nil)
	}, &value)
	return &value, err
}

func (c *Client) ListDatasets(ctx context.Context) ([]Dataset, error) {
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]Dataset, *string, error) {
		return fetchPage[Dataset](c, func() (*http.Response, error) {
			return c.api.ListDatasets(ctx, &openapi.ListDatasetsParams{Page: page})
		})
	})
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
	return listAllByPage(ctx, func(ctx context.Context, page *string) ([]Template, *string, error) {
		return fetchPage[Template](c, func() (*http.Response, error) {
			return c.api.ListTemplates(ctx, &openapi.ListTemplatesParams{Page: page})
		})
	})
}
