package sigma

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

var maxListPages = 10_000

type pageEnvelope[T any] struct {
	Entries  []T     `json:"entries"`
	NextPage *string `json:"nextPage"`
}

type pageTokenEnvelope[T any] struct {
	Entries       []T     `json:"entries"`
	NextPageToken *string `json:"nextPageToken"`
}

type pageFetcher[T any] func(ctx context.Context, cursor *string) ([]T, *string, error)

// listAllByPage follows Sigma's nextPage cursor, sending it as the generated
// `page` request parameter on each subsequent call.
func listAllByPage[T any](ctx context.Context, fetch pageFetcher[T]) ([]T, error) {
	return listAllCursors(ctx, fetch, "nextPage")
}

// listAllByPageToken follows Sigma's nextPageToken cursor, sending it as the
// generated `pageToken` request parameter on each subsequent call.
func listAllByPageToken[T any](ctx context.Context, fetch pageFetcher[T]) ([]T, error) {
	return listAllCursors(ctx, fetch, "nextPageToken")
}

func listAllCursors[T any](ctx context.Context, fetch pageFetcher[T], cursorName string) ([]T, error) {
	entries := make([]T, 0)
	var cursor *string
	seen := map[string]struct{}{}

	for pageNum := 0; pageNum < maxListPages; pageNum++ {
		page, next, err := fetch(ctx, cursor)
		if err != nil {
			return nil, err
		}
		entries = append(entries, page...)
		if next == nil || *next == "" {
			return entries, nil
		}
		if _, ok := seen[*next]; ok {
			return nil, fmt.Errorf("sigma pagination cycle detected: %s %q repeated", cursorName, *next)
		}
		seen[*next] = struct{}{}
		cursor = next
	}
	return nil, fmt.Errorf("sigma pagination exceeded %d pages", maxListPages)
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

func decodePageTokenEnvelope[T any](body []byte) (pageTokenEnvelope[T], error) {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return pageTokenEnvelope[T]{}, fmt.Errorf("unexpected Sigma list response: empty body")
	}

	var probe map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &probe); err != nil {
		return pageTokenEnvelope[T]{}, fmt.Errorf("decode Sigma API page token envelope: %w", err)
	}
	_, hasEntries := probe["entries"]
	_, hasNext := probe["nextPageToken"]
	if len(probe) > 0 && !hasEntries && !hasNext {
		return pageTokenEnvelope[T]{}, fmt.Errorf("unexpected Sigma list response envelope: expected entries/nextPageToken")
	}

	var page pageTokenEnvelope[T]
	if err := json.Unmarshal(trimmed, &page); err != nil {
		return pageTokenEnvelope[T]{}, fmt.Errorf("decode Sigma API page token envelope: %w", err)
	}
	return page, nil
}

func fetchPage[T any](client *Client, call func() (*http.Response, error)) ([]T, *string, error) {
	body, err := client.doAPI(call, true)
	if err != nil {
		return nil, nil, err
	}
	page, err := decodePageEnvelope[T](body)
	if err != nil {
		return nil, nil, err
	}
	return page.Entries, page.NextPage, nil
}

func fetchPageToken[T any](client *Client, call func() (*http.Response, error)) ([]T, *string, error) {
	body, err := client.doAPI(call, true)
	if err != nil {
		return nil, nil, err
	}
	page, err := decodePageTokenEnvelope[T](body)
	if err != nil {
		return nil, nil, err
	}
	return page.Entries, page.NextPageToken, nil
}
