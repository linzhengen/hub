package hubcli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/linzhengen/hub/server/pkg/apicatalog"
)

// Client calls the hub REST gateway.
type Client struct {
	endpoint string
	token    string
	http     *http.Client
}

// NewClient returns a client for the resolved settings.
func NewClient(settings Settings) *Client {
	return &Client{
		endpoint: strings.TrimRight(settings.Endpoint, "/"),
		token:    settings.Token,
		http:     &http.Client{Timeout: 60 * time.Second},
	}
}

// Request is a fully resolved HTTP call, kept as data so --dry-run can print
// exactly what would be sent without sending it.
type Request struct {
	Method string          `json:"method"`
	URL    string          `json:"url"`
	Body   json.RawMessage `json:"body,omitempty"`
}

// APIError is a non-2xx response from the gateway.
type APIError struct {
	StatusCode int
	Status     string
	Body       json.RawMessage
}

func (e *APIError) Error() string {
	if len(e.Body) > 0 {
		return fmt.Sprintf("%s: %s", e.Status, e.Body)
	}
	return e.Status
}

// BuildRequest maps an operation and its argument values onto an HTTP request.
//
// Where each value goes is decided by the operation's google.api.http rule, not
// by the caller: path fields are substituted into the URL, and the rest become
// either the JSON body or query parameters depending on whether the rpc
// declares a body. Values are keyed by protobuf or JSON field name.
func (c *Client) BuildRequest(op apicatalog.Operation, values map[string]any) (Request, error) {
	if op.HTTPMethod == "" {
		return Request{}, fmt.Errorf("%s.%s has no REST mapping", op.Service, op.Method)
	}

	path := op.Path
	query := url.Values{}
	body := map[string]any{}

	for name, value := range values {
		field, ok := op.Field(name)
		if !ok {
			return Request{}, fmt.Errorf("%s.%s has no field %q", op.Service, op.Method, name)
		}
		switch field.In {
		case apicatalog.InPath:
			placeholder := "{" + field.Name + "}"
			if !strings.Contains(path, placeholder) {
				return Request{}, fmt.Errorf("path %q has no placeholder for %q", op.Path, field.Name)
			}
			path = strings.ReplaceAll(path, placeholder, url.PathEscape(fmt.Sprint(value)))
		case apicatalog.InQuery:
			for _, item := range flatten(value) {
				query.Add(field.JSONName, item)
			}
		case apicatalog.InBody:
			body[field.JSONName] = value
		}
	}

	if i := strings.Index(path, "{"); i >= 0 {
		return Request{}, fmt.Errorf("missing required path parameter in %q", op.Path[i:])
	}

	req := Request{Method: op.HTTPMethod, URL: c.endpoint + path}
	if encoded := query.Encode(); encoded != "" {
		req.URL += "?" + encoded
	}
	if op.Body != "" {
		encoded, err := json.Marshal(body)
		if err != nil {
			return Request{}, err
		}
		req.Body = encoded
	}
	return req, nil
}

// Do sends a request and returns the decoded response body. A non-2xx status
// yields an *APIError carrying the gateway's own error payload, which is what
// tells a caller whether they were denied, throttled or simply wrong.
func (c *Client) Do(ctx context.Context, req Request) (json.RawMessage, error) {
	var reader io.Reader
	if len(req.Body) > 0 {
		reader = bytes.NewReader(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, reader)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Accept", "application/json")
	if len(req.Body) > 0 {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{StatusCode: resp.StatusCode, Status: resp.Status, Body: trimJSON(raw)}
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return json.RawMessage("{}"), nil
	}
	return trimJSON(raw), nil
}

// Call is BuildRequest followed by Do.
func (c *Client) Call(ctx context.Context, op apicatalog.Operation, values map[string]any) (json.RawMessage, error) {
	req, err := c.BuildRequest(op, values)
	if err != nil {
		return nil, err
	}
	return c.Do(ctx, req)
}

// flatten renders a query value as the one-or-more strings it contributes.
// Repeated fields become repeated parameters, which is how grpc-gateway reads
// them; a comma-joined value would arrive as a single element.
func flatten(value any) []string {
	switch v := value.(type) {
	case nil:
		return nil
	case []string:
		return v
	case []any:
		items := make([]string, 0, len(v))
		for _, item := range v {
			items = append(items, fmt.Sprint(item))
		}
		return items
	default:
		return []string{fmt.Sprint(v)}
	}
}

func trimJSON(raw []byte) json.RawMessage {
	trimmed := bytes.TrimSpace(raw)
	if !json.Valid(trimmed) {
		// Not JSON - a proxy error page, say. Quote it so the caller still
		// gets a well-formed document to print.
		quoted, err := json.Marshal(string(trimmed))
		if err != nil {
			return json.RawMessage(`""`)
		}
		return quoted
	}
	return trimmed
}
