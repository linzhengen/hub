package hubcli

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/linzhengen/hub/v1/server/pkg/apicatalog"
)

// The reference is what an agent reads to decide which endpoint to call, so a
// gap in it produces confidently wrong calls rather than an obvious failure.
func TestRenderReferenceCoversEveryOperation(t *testing.T) {
	catalog := apicatalog.Default()
	reference := RenderReference(catalog)

	for _, op := range catalog.Operations() {
		assert.Contains(t, reference, "### "+op.Method, "%s is missing", op.Method)
		assert.Contains(t, reference, fmt.Sprintf("`%s %s`", op.HTTPMethod, op.Path))
		assert.Contains(t, reference, fmt.Sprintf("`hub %s %s`", GroupName(op), CommandName(op)))
		if op.Summary != "" {
			assert.Contains(t, reference, op.Summary)
		}
	}

	for _, service := range catalog.Services() {
		assert.Contains(t, reference, "## "+service)
	}
}

func TestRenderReferenceDescribesRbac(t *testing.T) {
	reference := RenderReference(apicatalog.Default())

	// A guarded rpc names the permission it needs.
	assert.Contains(t, reference, "- rbac: `DeleteUser` on `api.user.v1.UserService`")
	// A public one says so instead of naming a permission nobody holds.
	assert.Contains(t, reference, "- rbac: none, any authenticated caller may use this rpc")

	getMe := section(t, reference, "GetMe")
	assert.NotContains(t, getMe, "on `api.user.v1.UserService`")
}

func TestRenderReferenceTabulatesRequestFields(t *testing.T) {
	reference := RenderReference(apicatalog.Default())

	listUser := section(t, reference, "ListUser")
	assert.Contains(t, listUser, "| flag | field | in | type |")
	assert.Contains(t, listUser, "| `--user-ids` | `userIds` | query | repeated string |")
	assert.Contains(t, listUser, "| `--limit` | `limit` | query | uint32 |")
	// Enum values are spelled out, so an agent does not have to guess them.
	assert.Contains(t, listUser, "`STATUS_ACTIVE`")

	getMe := section(t, reference, "GetMe")
	assert.Contains(t, getMe, "Takes no parameters.")
}

// section returns the part of the reference describing one rpc.
func section(t *testing.T, reference, method string) string {
	t.Helper()

	start := strings.Index(reference, "### "+method+"\n")
	require.GreaterOrEqual(t, start, 0, "%s is not in the reference", method)

	rest := reference[start+1:]
	if end := strings.Index(rest, "\n### "); end >= 0 {
		return rest[:end]
	}
	return rest
}
