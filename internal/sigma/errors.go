package sigma

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// APIError describes an error returned by the Sigma REST API.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	RequestID  string
}

func (err *APIError) Error() string {
	if err.Code != "" {
		return fmt.Sprintf("Sigma API error (%d, %s): %s", err.StatusCode, err.Code, err.Message)
	}
	return fmt.Sprintf("Sigma API error (%d): %s", err.StatusCode, err.Message)
}

// IsNotFound reports whether err indicates a missing Sigma resource.
//
// Only HTTP 404 responses that look like Sigma (or provider-synthesized) API
// errors qualify. HTML/proxy error pages are excluded so a misrouted reverse
// proxy does not remove Terraform state. When Sigma returns a bare 404 without
// a distinguishing code, the provider still treats it as not-found; that
// limitation is inherent to the API.
func IsNotFound(err error) bool {
	var apiError *APIError
	if !errors.As(err, &apiError) || apiError.StatusCode != http.StatusNotFound {
		return false
	}
	return isResourceNotFoundAPIError(apiError)
}

func isResourceNotFoundAPIError(apiError *APIError) bool {
	code := strings.ToLower(strings.TrimSpace(apiError.Code))
	switch code {
	case "", "not_found", "notfound", "resource_not_found", "entity_not_found":
	default:
		return false
	}

	message := strings.TrimSpace(apiError.Message)
	lower := strings.ToLower(message)
	if strings.Contains(lower, "<html") || strings.Contains(lower, "<!doctype") {
		return false
	}
	return true
}

func decodeAPIError(response *http.Response) error {
	body, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return fmt.Errorf("read Sigma API error response: %w", readErr)
	}

	var payload struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"requestId"`
		Error     struct {
			Code      string `json:"code"`
			Message   string `json:"message"`
			RequestID string `json:"requestId"`
		} `json:"error"`
	}
	_ = json.Unmarshal(body, &payload)
	if payload.Code == "" {
		payload.Code = payload.Error.Code
	}
	if payload.Message == "" {
		payload.Message = payload.Error.Message
	}
	if payload.RequestID == "" {
		payload.RequestID = payload.Error.RequestID
	}
	if payload.RequestID == "" {
		payload.RequestID = response.Header.Get("X-Request-Id")
	}
	if payload.Message == "" {
		payload.Message = string(body)
	}

	return &APIError{
		StatusCode: response.StatusCode,
		Code:       payload.Code,
		Message:    payload.Message,
		RequestID:  payload.RequestID,
	}
}
