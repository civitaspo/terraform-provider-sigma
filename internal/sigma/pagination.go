package sigma

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

const maxListPages = 10_000

type pageEnvelope[T any] struct {
	Entries  []T     `json:"entries"`
	NextPage *string `json:"nextPage"`
}

// ListAll follows Sigma's page cursor envelope and returns every entry.
// It also accepts legacy bare JSON arrays. Repeated nextPage cursors and
// unbounded page counts return an error instead of looping forever.
func ListAll[T any](ctx context.Context, client *Client, path string) ([]T, error) {
	var entries []T
	nextPath := path
	seen := map[string]struct{}{}

	for pageNum := 0; pageNum < maxListPages; pageNum++ {
		body, err := client.getRaw(ctx, nextPath)
		if err != nil {
			return nil, err
		}
		page, err := decodePageEnvelope[T](body)
		if err != nil {
			return nil, err
		}
		entries = append(entries, page.Entries...)
		if page.NextPage == nil || *page.NextPage == "" {
			return entries, nil
		}

		cursor := *page.NextPage
		if _, ok := seen[cursor]; ok {
			return nil, fmt.Errorf("sigma pagination cycle detected: nextPage %q repeated", cursor)
		}
		seen[cursor] = struct{}{}

		parsed, err := url.Parse(path)
		if err != nil {
			return nil, fmt.Errorf("parse paginated Sigma API path: %w", err)
		}
		query := parsed.Query()
		query.Set("page", cursor)
		parsed.RawQuery = query.Encode()
		nextPath = parsed.String()
	}
	return nil, fmt.Errorf("sigma pagination exceeded %d pages for %s", maxListPages, path)
}

func decodePageEnvelope[T any](body []byte) (pageEnvelope[T], error) {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return pageEnvelope[T]{}, fmt.Errorf("unexpected Sigma list response: empty body")
	}
	if trimmed[0] == '[' {
		var entries []T
		if err := json.Unmarshal(trimmed, &entries); err != nil {
			return pageEnvelope[T]{}, fmt.Errorf("decode Sigma API list array: %w", err)
		}
		return pageEnvelope[T]{Entries: entries}, nil
	}

	var probe map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &probe); err != nil {
		return pageEnvelope[T]{}, fmt.Errorf("decode Sigma API page envelope: %w", err)
	}
	_, hasEntries := probe["entries"]
	_, hasNext := probe["nextPage"]
	if len(probe) > 0 && !hasEntries && !hasNext {
		return pageEnvelope[T]{}, fmt.Errorf("unexpected Sigma list response envelope: expected entries/nextPage or a JSON array")
	}

	var page pageEnvelope[T]
	if err := json.Unmarshal(trimmed, &page); err != nil {
		return pageEnvelope[T]{}, fmt.Errorf("decode Sigma API page envelope: %w", err)
	}
	return page, nil
}
