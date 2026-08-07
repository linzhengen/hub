package hubcli

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// Format is how a command renders its result.
type Format string

const (
	// FormatJSON is indented JSON. It is the default because the CLI's main
	// consumers - scripts and agents - parse what they read.
	FormatJSON Format = "json"
	// FormatYAML is YAML, for eyeballing large responses.
	FormatYAML Format = "yaml"
	// FormatTable renders a response's list field as columns, falling back to
	// JSON for anything that is not a list of objects.
	FormatTable Format = "table"
)

// Formats lists the accepted --output values.
var Formats = []string{string(FormatJSON), string(FormatYAML), string(FormatTable)}

// ParseFormat validates an --output value.
func ParseFormat(value string) (Format, error) {
	switch Format(value) {
	case FormatJSON, FormatYAML, FormatTable:
		return Format(value), nil
	default:
		return "", fmt.Errorf("unknown output format %q, want one of %s", value, strings.Join(Formats, ", "))
	}
}

// Render writes a value in the requested format.
func Render(w io.Writer, format Format, value any) error {
	switch format {
	case FormatYAML:
		return renderYAML(w, value)
	case FormatTable:
		return renderTable(w, value)
	default:
		return renderJSON(w, value)
	}
}

func renderJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	// Terminal output, not a web page: leaving &, < and > alone keeps URLs and
	// query strings readable and copy-pasteable.
	encoder.SetEscapeHTML(false)
	return encoder.Encode(decoded(value))
}

func renderYAML(w io.Writer, value any) error {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	if err := encoder.Encode(decoded(value)); err != nil {
		return err
	}
	return encoder.Close()
}

// renderTable prints the rows of the response's single list field. Responses
// wrap their payload ({"users": [...], "total": 3}), so the table is of that
// list; anything else falls back to JSON rather than inventing a shape.
func renderTable(w io.Writer, value any) error {
	rows, columns, ok := tabulate(normalise(value))
	if !ok {
		return renderJSON(w, value)
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	// tabwriter buffers, so a write error surfaces from Flush below.
	_, _ = fmt.Fprintln(tw, strings.Join(columns, "\t"))
	for _, row := range rows {
		cells := make([]string, 0, len(columns))
		for _, column := range columns {
			cells = append(cells, cell(row[column]))
		}
		_, _ = fmt.Fprintln(tw, strings.Join(cells, "\t"))
	}
	return tw.Flush()
}

func tabulate(value any) ([]map[string]any, []string, bool) {
	list, ok := listOf(value)
	if !ok {
		return nil, nil, false
	}

	rows := make([]map[string]any, 0, len(list))
	seen := map[string]bool{}
	var columns []string
	for _, item := range list {
		row, ok := item.(map[string]any)
		if !ok {
			return nil, nil, false
		}
		rows = append(rows, row)
		for key := range row {
			if !seen[key] {
				seen[key] = true
				columns = append(columns, key)
			}
		}
	}
	if len(columns) == 0 {
		return nil, nil, false
	}
	sort.Strings(columns)
	return rows, columns, true
}

// listOf finds the list to tabulate: the value itself when it is already a
// list, otherwise the object's only list field.
func listOf(value any) ([]any, bool) {
	switch v := value.(type) {
	case []any:
		return v, true
	case map[string]any:
		var found []any
		count := 0
		for _, key := range sortedKeys(v) {
			if list, ok := v[key].([]any); ok {
				found = list
				count++
			}
		}
		return found, count == 1
	default:
		return nil, false
	}
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func cell(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case []any:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			parts = append(parts, cell(item))
		}
		return strings.Join(parts, ",")
	case map[string]any:
		encoded, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprint(v)
		}
		return string(encoded)
	default:
		return fmt.Sprint(v)
	}
}

// decoded unwraps an API response so the encoder sees its structure rather
// than a byte string. The CLI's own result types are passed through untouched,
// which keeps their fields in declaration order instead of alphabetised.
func decoded(value any) any {
	raw, ok := value.(json.RawMessage)
	if !ok {
		return value
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return string(raw)
	}
	return out
}

// normalise flattens anything - raw JSON, a struct, a map - into plain maps and
// slices, which is what the table renderer needs in order to find its columns.
func normalise(value any) any {
	switch v := value.(type) {
	case json.RawMessage:
		return decoded(v)
	case map[string]any, []any, string, nil:
		return v
	default:
		encoded, err := json.Marshal(v)
		if err != nil {
			return v
		}
		var out any
		if err := json.Unmarshal(encoded, &out); err != nil {
			return v
		}
		return out
	}
}
