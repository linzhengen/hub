package hubcli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/linzhengen/hub/v1/server/pkg/apicatalog"
)

// run executes the CLI as a caller would and returns its stdout.
func run(t *testing.T, args ...string) (string, error) {
	t.Helper()

	// Point the config at an empty directory so a developer's own ~/.hub does
	// not leak into the test.
	t.Setenv("HUB_CONFIG", filepath.Join(t.TempDir(), "config.yaml"))

	out := &bytes.Buffer{}
	root := NewRootCommand(Options{})
	root.SetOut(out)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs(args)

	err := root.Execute()
	return out.String(), err
}

func TestGeneratedCommandsCoverEveryOperation(t *testing.T) {
	root := NewRootCommand(Options{})

	for _, op := range apicatalog.Default().Operations() {
		group, _, err := root.Find([]string{GroupName(op), CommandName(op)})
		require.NoError(t, err)
		assert.Equal(t, CommandName(op), group.Name(), "%s.%s has no command", op.Service, op.Method)
	}
}

func TestDryRunPrintsTheRequestWithoutSendingIt(t *testing.T) {
	out := &bytes.Buffer{}
	t.Setenv("HUB_CONFIG", filepath.Join(t.TempDir(), "config.yaml"))
	t.Setenv("HUB_ENDPOINT", "http://hub.test")

	root := NewRootCommand(Options{})
	root.SetOut(out)
	root.SetArgs([]string{"user", "list", "--dry-run", "--limit", "10", "--user-ids", "a", "--user-ids", "b"})
	require.NoError(t, root.Execute())

	var req Request
	require.NoError(t, json.Unmarshal(out.Bytes(), &req))
	assert.Equal(t, "GET", req.Method)
	assert.Equal(t, "http://hub.test/api/v1/users?limit=10&userIds=a&userIds=b", req.URL)
}

// Only the flags the caller set are sent, so `update-user --username x` cannot
// silently blank the email it never mentioned.
func TestUnsetFlagsAreNotSent(t *testing.T) {
	out := &bytes.Buffer{}
	t.Setenv("HUB_CONFIG", filepath.Join(t.TempDir(), "config.yaml"))

	root := NewRootCommand(Options{})
	root.SetOut(out)
	root.SetArgs([]string{"user", "update-user", "--dry-run", "--id", "u1", "--username", "taro"})
	require.NoError(t, root.Execute())

	var req Request
	require.NoError(t, json.Unmarshal(out.Bytes(), &req))
	assert.JSONEq(t, `{"username":"taro"}`, string(req.Body))
}

func TestAliasesResolveToTheCrudOperations(t *testing.T) {
	root := NewRootCommand(Options{})

	for _, args := range [][]string{
		{"user", "list"},
		{"group", "get"},
		{"role", "create"},
		{"permission", "update"},
		{"resource", "delete"},
	} {
		cmd, _, err := root.Find(args)
		require.NoError(t, err, args)
		assert.Contains(t, cmd.Aliases, args[1])
	}
}

func TestApiListIsMachineReadable(t *testing.T) {
	out, err := run(t, "api", "list", "--service", "UserService")
	require.NoError(t, err)

	var payload struct {
		Operations []operationSummary `json:"operations"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	assert.Len(t, payload.Operations, len(apicatalog.Default().OperationsOf("user.v1.UserService")))

	for _, op := range payload.Operations {
		assert.NotEmpty(t, op.Command)
		assert.NotEmpty(t, op.Path)
		assert.NotEmpty(t, op.Summary)
	}
}

func TestApiDescribeReportsTheRbacRule(t *testing.T) {
	out, err := run(t, "api", "describe", "UserService.DeleteUser")
	require.NoError(t, err)

	var summary operationSummary
	require.NoError(t, json.Unmarshal([]byte(out), &summary))
	assert.Equal(t, "api.user.v1.UserService", summary.Resource)
	assert.Equal(t, "DeleteUser", summary.Action)
	assert.False(t, summary.Public)
	assert.Equal(t, "hub user delete-user", summary.Command)
}

func TestApiCallBuildsAnArbitraryRequest(t *testing.T) {
	t.Setenv("HUB_ENDPOINT", "http://hub.test")

	out, err := run(t, "api", "call", "POST", "/api/v1/groups", "--data", `{"name":"platform"}`, "--dry-run")
	require.NoError(t, err)

	var req Request
	require.NoError(t, json.Unmarshal([]byte(out), &req))
	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, "http://hub.test/api/v1/groups", req.URL)
	assert.JSONEq(t, `{"name":"platform"}`, string(req.Body))
}

func TestApiCallRejectsBadInput(t *testing.T) {
	_, err := run(t, "api", "call", "TRACE", "/api/v1/users", "--dry-run")
	assert.ErrorContains(t, err, `unsupported HTTP method "TRACE"`)

	_, err = run(t, "api", "call", "POST", "/api/v1/users", "--data", "not json", "--dry-run")
	assert.ErrorContains(t, err, "not valid JSON")

	_, err = run(t, "api", "call", "GET", "/api/v1/users", "--query", "limit", "--dry-run")
	assert.ErrorContains(t, err, "key=value")
}

func TestUnknownOutputFormatIsRejected(t *testing.T) {
	_, err := run(t, "api", "list", "-o", "xml")
	assert.ErrorContains(t, err, "unknown output format")
}

func TestEndToEndAgainstAStubGateway(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/users", r.URL.Path)
		assert.Equal(t, "Bearer t0ken", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"users":[{"id":"u1","username":"taro"}],"total":"1"}`))
	}))
	defer server.Close()

	t.Setenv("HUB_ENDPOINT", server.URL)
	t.Setenv("HUB_TOKEN", "t0ken")

	out, err := run(t, "user", "list", "-o", "table")
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(out), "\n")
	require.Len(t, lines, 2)
	assert.Contains(t, lines[0], "id")
	assert.Contains(t, lines[0], "username")
	assert.Contains(t, lines[1], "u1")
	assert.Contains(t, lines[1], "taro")
}
