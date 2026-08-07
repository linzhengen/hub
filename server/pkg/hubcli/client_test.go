package hubcli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/linzhengen/hub/v1/server/pkg/apicatalog"
)

func operation(t *testing.T, ref string) apicatalog.Operation {
	t.Helper()
	op, err := apicatalog.Default().Find(ref)
	require.NoError(t, err)
	return op
}

func TestBuildRequestRoutesFieldsByTheirHttpRule(t *testing.T) {
	client := NewClient(Settings{Endpoint: "http://hub.test/"})

	t.Run("path and query", func(t *testing.T) {
		req, err := client.BuildRequest(operation(t, "ListUser"), map[string]any{
			"limit":    10,
			"user_ids": []string{"a", "b"},
			"status":   "STATUS_ACTIVE",
		})
		require.NoError(t, err)

		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "http://hub.test/api/v1/users?limit=10&status=STATUS_ACTIVE&userIds=a&userIds=b", req.URL)
		assert.Empty(t, req.Body, "a GET carries no body")
	})

	t.Run("path parameter and body", func(t *testing.T) {
		req, err := client.BuildRequest(operation(t, "AddUsersToGroup"), map[string]any{
			"id":       "g1",
			"user_ids": []string{"u1"},
		})
		require.NoError(t, err)

		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "http://hub.test/api/v1/groups/g1/users/add", req.URL)
		assert.JSONEq(t, `{"userIds":["u1"]}`, string(req.Body))
	})

	t.Run("fields are addressable by their JSON name too", func(t *testing.T) {
		req, err := client.BuildRequest(operation(t, "AddPermissionsToRole"), map[string]any{"permissionIds": []string{"p1"}, "id": "r1"})
		require.NoError(t, err)
		assert.Equal(t, "http://hub.test/api/v1/roles/r1/permissions/add", req.URL)
		assert.JSONEq(t, `{"permissionIds":["p1"]}`, string(req.Body))
	})
}

func TestBuildRequestRejectsBadInput(t *testing.T) {
	client := NewClient(Settings{Endpoint: "http://hub.test"})

	_, err := client.BuildRequest(operation(t, "GetUser"), nil)
	assert.ErrorContains(t, err, "missing required path parameter")

	_, err = client.BuildRequest(operation(t, "GetUser"), map[string]any{"nope": "x"})
	assert.ErrorContains(t, err, `has no field "nope"`)
}

// Path values are escaped, so an id containing a slash cannot walk the caller
// onto a different endpoint.
func TestBuildRequestEscapesPathValues(t *testing.T) {
	client := NewClient(Settings{Endpoint: "http://hub.test"})

	req, err := client.BuildRequest(operation(t, "GetUser"), map[string]any{"id": "a/b"})
	require.NoError(t, err)
	assert.Equal(t, "http://hub.test/api/v1/users/a%2Fb", req.URL)
}

func TestDo(t *testing.T) {
	var (
		gotAuth   string
		gotMethod string
		gotPath   string
		gotBody   string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth, gotMethod, gotPath = r.Header.Get("Authorization"), r.Method, r.URL.RequestURI()
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		gotBody = string(body)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"user":{"id":"u1"}}`))
	}))
	defer server.Close()

	client := NewClient(Settings{Endpoint: server.URL, Token: "secret"})
	response, err := client.Call(context.Background(), operation(t, "UpdateUser"), map[string]any{
		"id":       "u1",
		"username": "taro",
	})
	require.NoError(t, err)

	assert.Equal(t, "Bearer secret", gotAuth)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/api/v1/users/u1", gotPath)
	assert.JSONEq(t, `{"username":"taro"}`, gotBody)
	assert.JSONEq(t, `{"user":{"id":"u1"}}`, string(response))
}

func TestDoReportsTheGatewayError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":7,"message":"permission denied for ListUser on api.user.v1.UserService"}`))
	}))
	defer server.Close()

	client := NewClient(Settings{Endpoint: server.URL})
	_, err := client.Call(context.Background(), operation(t, "ListUser"), nil)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusForbidden, apiErr.StatusCode)
	assert.Contains(t, err.Error(), "permission denied for ListUser")
}

// An empty 200 - what a Delete returns - must still decode as a document
// rather than blowing up whatever is parsing the CLI's output.
func TestDoTreatsAnEmptyBodyAsAnEmptyObject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(Settings{Endpoint: server.URL})
	response, err := client.Call(context.Background(), operation(t, "DeleteUser"), map[string]any{"id": "u1"})

	require.NoError(t, err)
	assert.Equal(t, json.RawMessage("{}"), response)
}

func TestDoWrapsANonJsonErrorBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("<html>proxy error</html>"))
	}))
	defer server.Close()

	client := NewClient(Settings{Endpoint: server.URL})
	_, err := client.Call(context.Background(), operation(t, "ListUser"), nil)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.True(t, json.Valid(apiErr.Body), "the body is always a valid JSON document")
	assert.Contains(t, string(apiErr.Body), "proxy error")
}
