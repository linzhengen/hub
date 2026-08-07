package hubcli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func render(t *testing.T, format Format, value any) string {
	t.Helper()
	out := &bytes.Buffer{}
	require.NoError(t, Render(out, format, value))
	return out.String()
}

func TestParseFormat(t *testing.T) {
	for _, name := range Formats {
		format, err := ParseFormat(name)
		require.NoError(t, err)
		assert.Equal(t, Format(name), format)
	}

	_, err := ParseFormat("xml")
	assert.ErrorContains(t, err, "unknown output format")
}

func TestRenderJSON(t *testing.T) {
	t.Run("unwraps an API response rather than quoting it", func(t *testing.T) {
		out := render(t, FormatJSON, json.RawMessage(`{"user":{"id":"u1"}}`))
		assert.JSONEq(t, `{"user":{"id":"u1"}}`, out)
	})

	// Terminal output: escaping & would make a URL unreadable and unpasteable.
	t.Run("leaves URL characters alone", func(t *testing.T) {
		out := render(t, FormatJSON, Request{Method: "GET", URL: "http://hub.test/users?a=1&b=2"})
		assert.Contains(t, out, "?a=1&b=2")
		assert.NotContains(t, out, `\u0026`, "the ampersand must not be escaped")
	})
}

func TestRenderYAML(t *testing.T) {
	out := render(t, FormatYAML, json.RawMessage(`{"total":2}`))
	assert.Equal(t, "total: 2\n", out)
}

func TestRenderTable(t *testing.T) {
	t.Run("tabulates the response's list of objects", func(t *testing.T) {
		out := render(t, FormatTable, json.RawMessage(`{"users":[{"id":"u1","name":"taro"},{"id":"u2","name":"hanako"}],"total":"2"}`))

		lines := strings.Split(strings.TrimSpace(out), "\n")
		require.Len(t, lines, 3)
		assert.Regexp(t, `^id\s+name$`, lines[0])
		assert.Contains(t, lines[1], "u1")
		assert.Contains(t, lines[1], "taro")
	})

	t.Run("a row missing a column leaves the cell empty", func(t *testing.T) {
		out := render(t, FormatTable, json.RawMessage(`{"items":[{"a":"1","b":"2"},{"a":"3"}]}`))

		lines := strings.Split(strings.TrimSpace(out), "\n")
		require.Len(t, lines, 3)
		assert.Regexp(t, `^a\s+b$`, lines[0])
		assert.Equal(t, "3", strings.TrimSpace(lines[2]))
	})

	t.Run("nested values are rendered inline", func(t *testing.T) {
		out := render(t, FormatTable, json.RawMessage(`{"items":[{"ids":["a","b"],"meta":{"k":"v"}}]}`))

		assert.Contains(t, out, "a,b")
		assert.Contains(t, out, `{"k":"v"}`)
	})

	// Rather than invent a shape, anything that is not a single list of
	// objects falls back to JSON so no information is lost.
	t.Run("falls back to JSON for a single object", func(t *testing.T) {
		out := render(t, FormatTable, json.RawMessage(`{"user":{"id":"u1"}}`))
		assert.JSONEq(t, `{"user":{"id":"u1"}}`, out)
	})

	t.Run("falls back to JSON when two lists are present", func(t *testing.T) {
		out := render(t, FormatTable, json.RawMessage(`{"users":[{"id":"u1"}],"groups":[{"id":"g1"}]}`))
		assert.JSONEq(t, `{"users":[{"id":"u1"}],"groups":[{"id":"g1"}]}`, out)
	})

	t.Run("falls back to JSON for a list of scalars", func(t *testing.T) {
		out := render(t, FormatTable, json.RawMessage(`{"ids":["a","b"]}`))
		assert.JSONEq(t, `{"ids":["a","b"]}`, out)
	})
}
