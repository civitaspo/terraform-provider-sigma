package sigma

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

// IsNotFound reports whether err is a Sigma API 404 response.
func IsNotFound(err error) bool {
	var apiError *APIError
	return errors.As(err, &apiError) && apiError.StatusCode == http.StatusNotFound
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
