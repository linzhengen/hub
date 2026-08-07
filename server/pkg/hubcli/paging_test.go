package hubcli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/linzhengen/hub/v1/server/pkg/apicatalog"
)

func TestPaginated(t *testing.T) {
	catalog := apicatalog.Default()

	list, ok := catalog.ByFullMethod("/user.v1.UserService/ListUser")
	require.True(t, ok)
	assert.True(t, Paginated(list))

	get, ok := catalog.ByFullMethod("/user.v1.UserService/GetUser")
	require.True(t, ok)
	assert.False(t, Paginated(get))
}

// pagedServer serves `total` users, honouring limit and offset the way the
// gateway does. Note it reports the total as a string: protobuf's JSON mapping
// encodes int64 that way, and getting it wrong would break the loop.
func pagedServer(t *testing.T, total int) (*httptest.Server, *int) {
	t.Helper()

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		if limit == 0 {
			limit = 50
		}

		users := make([]map[string]string, 0, limit)
		for i := offset; i < offset+limit && i < total; i++ {
			users = append(users, map[string]string{"id": fmt.Sprintf("u%d", i)})
		}

		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"users": users,
			"total": strconv.Itoa(total),
		}))
	}))
	t.Cleanup(server.Close)
	return server, &requests
}

func TestCollectAllPages(t *testing.T) {
	op, err := apicatalog.Default().Find("ListUser")
	require.NoError(t, err)

	t.Run("follows every page", func(t *testing.T) {
		server, requests := pagedServer(t, 25)
		c := &CLI{client: NewClient(Settings{Endpoint: server.URL})}

		raw, err := c.collectAllPages(context.Background(), op, map[string]any{"limit": 10})
		require.NoError(t, err)

		var body struct {
			Users []map[string]string `json:"users"`
			Total int64               `json:"total"`
		}
		require.NoError(t, json.Unmarshal(raw, &body))
		assert.Len(t, body.Users, 25)
		assert.Equal(t, int64(25), body.Total)
		assert.Equal(t, "u0", body.Users[0]["id"])
		assert.Equal(t, "u24", body.Users[24]["id"])
		assert.Equal(t, 3, *requests, "25 items at 10 a page is three requests")
	})

	t.Run("a single page needs a single request", func(t *testing.T) {
		server, requests := pagedServer(t, 3)
		c := &CLI{client: NewClient(Settings{Endpoint: server.URL})}

		raw, err := c.collectAllPages(context.Background(), op, map[string]any{"limit": 10})
		require.NoError(t, err)

		assert.Contains(t, string(raw), "u2")
		assert.Equal(t, 1, *requests)
	})

	t.Run("an empty result is not an error", func(t *testing.T) {
		server, _ := pagedServer(t, 0)
		c := &CLI{client: NewClient(Settings{Endpoint: server.URL})}

		raw, err := c.collectAllPages(context.Background(), op, nil)
		require.NoError(t, err)
		assert.JSONEq(t, `{"users":[],"total":0}`, string(raw))
	})

	t.Run("refuses an operation that does not page", func(t *testing.T) {
		get, err := apicatalog.Default().Find("GetUser")
		require.NoError(t, err)
		c := &CLI{client: NewClient(Settings{Endpoint: "http://hub.test"})}

		_, err = c.collectAllPages(context.Background(), get, map[string]any{"id": "u1"})
		assert.ErrorContains(t, err, "--all only works on a list operation")
	})

	// A server that reports a total it never delivers must not spin forever.
	t.Run("stops when a page comes back empty", func(t *testing.T) {
		requests := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			requests++
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"users":[],"total":"999"}`))
		}))
		defer server.Close()

		c := &CLI{client: NewClient(Settings{Endpoint: server.URL})}
		raw, err := c.collectAllPages(context.Background(), op, map[string]any{"limit": 10})

		require.NoError(t, err)
		assert.JSONEq(t, `{"users":[],"total":999}`, string(raw))
		assert.Equal(t, 1, requests)
	})

	t.Run("the caller's values are not modified", func(t *testing.T) {
		server, _ := pagedServer(t, 25)
		c := &CLI{client: NewClient(Settings{Endpoint: server.URL})}
		values := map[string]any{"limit": 10}

		_, err := c.collectAllPages(context.Background(), op, values)
		require.NoError(t, err)

		assert.Equal(t, map[string]any{"limit": 10}, values)
	})
}

func TestToInt64(t *testing.T) {
	// protobuf's JSON mapping sends int64 as a string, so both forms arrive.
	assert.Equal(t, int64(42), toInt64("42"))
	assert.Equal(t, int64(42), toInt64(float64(42)))
	assert.Equal(t, int64(0), toInt64(nil))
	assert.Equal(t, int64(0), toInt64("not a number"))
}
