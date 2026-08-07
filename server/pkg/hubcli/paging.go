package hubcli

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/linzhengen/hub/v1/server/pkg/apicatalog"
)

// maxPagesFollowed stops --all from looping forever if a server keeps
// reporting a total it never reaches.
const maxPagesFollowed = 1000

// Paginated reports whether an operation is a list endpoint --all can follow:
// one that takes a limit and an offset.
func Paginated(op apicatalog.Operation) bool {
	limit, hasLimit := op.Field("limit")
	offset, hasOffset := op.Field("offset")
	return hasLimit && hasOffset && limit.In == apicatalog.InQuery && offset.In == apicatalog.InQuery
}

// collectAllPages walks a list endpoint to the end and returns one response
// holding every item.
//
// Since list endpoints stopped returning the whole table, a caller that wants
// everything has to page for it. Doing that in the CLI keeps the loop - and the
// off-by-one risks in it - out of every script and agent that needs the full
// set.
func (c *CLI) collectAllPages(ctx context.Context, op apicatalog.Operation, values map[string]any) (json.RawMessage, error) {
	if !Paginated(op) {
		return nil, fmt.Errorf("--all only works on a list operation; %s.%s takes no limit and offset", op.Service, op.Method)
	}

	// Copy: the caller's map is theirs, and we are about to drive the offset.
	page := make(map[string]any, len(values)+2)
	for k, v := range values {
		page[k] = v
	}

	var (
		listKey string
		total   int64
		offset  int64
	)
	// Not nil: an empty result must serialise as [] rather than null, or a
	// caller walking `.users[]` breaks on the empty case.
	items := []any{}
	if requested, ok := page["offset"]; ok {
		offset = toInt64(requested)
	}

	for range maxPagesFollowed {
		page["offset"] = offset

		raw, err := c.client.Call(ctx, op, page)
		if err != nil {
			return nil, err
		}

		var body map[string]any
		if err := json.Unmarshal(raw, &body); err != nil {
			return nil, fmt.Errorf("list response was not an object: %w", err)
		}

		key, batch, err := singleList(body)
		if err != nil {
			return nil, err
		}
		listKey = key
		items = append(items, batch...)
		total = toInt64(body["total"])

		// A page the server could not fill is the last one; without this an
		// endpoint that reports no total would loop until the cap.
		if len(batch) == 0 || int64(len(items)) >= total {
			break
		}
		offset += int64(len(batch))
	}

	return json.Marshal(map[string]any{listKey: items, "total": total})
}

// singleList finds the one array field a list response wraps its items in.
func singleList(body map[string]any) (string, []any, error) {
	var (
		key   string
		items []any
		found int
	)
	for k, v := range body {
		if list, ok := v.([]any); ok {
			key, items, found = k, list, found+1
		}
	}
	if found != 1 {
		return "", nil, fmt.Errorf("expected the response to hold exactly one list of items, found %d", found)
	}
	return key, items, nil
}

// toInt64 reads a JSON number that may have arrived as a string, which is how
// protobuf's JSON mapping encodes int64.
func toInt64(value any) int64 {
	switch v := value.(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}
