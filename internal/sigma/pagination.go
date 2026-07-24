package sigma

import (
	"context"
	"fmt"
	"net/url"
)

type pageEnvelope[T any] struct {
	Entries  []T     `json:"entries"`
	NextPage *string `json:"nextPage"`
}

// ListAll follows Sigma's page cursor envelope and returns every entry.
func ListAll[T any](ctx context.Context, client *Client, path string) ([]T, error) {
	var entries []T
	nextPath := path

	for {
		var page pageEnvelope[T]
		if err := client.getJSON(ctx, nextPath, &page); err != nil {
			return nil, err
		}
		entries = append(entries, page.Entries...)
		if page.NextPage == nil || *page.NextPage == "" {
			return entries, nil
		}

		parsed, err := url.Parse(path)
		if err != nil {
			return nil, fmt.Errorf("parse paginated Sigma API path: %w", err)
		}
		query := parsed.Query()
		query.Set("page", *page.NextPage)
		parsed.RawQuery = query.Encode()
		nextPath = parsed.String()
	}
}
