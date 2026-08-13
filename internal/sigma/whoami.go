package sigma

import (
	"context"
	"net/http"
)

// WhoamiResponse is the identity returned for the authenticated principal.
type WhoamiResponse struct {
	UserID         string `json:"userId"`
	OrganizationID string `json:"organizationId"`
}

// Whoami returns the identity of the authenticated principal.
func (client *Client) Whoami(ctx context.Context) (*WhoamiResponse, error) {
	var result WhoamiResponse
	if err := client.doDecode(func() (*http.Response, error) {
		return client.api.Whoami(ctx, nil)
	}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
