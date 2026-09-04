package postgres

import (
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sqlFor(t *testing.T, e goqu.Expression) string {
	t.Helper()
	query, _, err := goqu.Dialect("postgres").From("groups").Where(e).ToSQL()
	require.NoError(t, err)
	return query
}

// TestIn_EmptyMatchesNothing is the regression.
//
// goqu renders an empty slice as `IN ()`, which PostgreSQL rejects with a
// syntax error - so a listing narrowed to an empty set failed with a 500 rather
// than returning an empty page. `GetMe` was the case that showed it: a
// principal in no group asks for its groups, and the one rpc that is public so
// that a caller with no grants can still read its own profile was the one that
// broke.
func TestIn_EmptyMatchesNothing(t *testing.T) {
	query := sqlFor(t, In("id", []string{}))

	assert.NotContains(t, query, "IN ()", "an empty set must not render as invalid SQL")
	assert.Contains(t, query, "FALSE", "an empty set matches nothing")
}

// A nil slice reaches here only when a caller decided to filter by nothing,
// which is the same answer as an empty one: the "no filter at all" case is the
// caller not calling In.
func TestIn_NilMatchesNothing(t *testing.T) {
	assert.Contains(t, sqlFor(t, In("id", nil)), "FALSE")
}

func TestIn_NonEmptyFiltersToTheSet(t *testing.T) {
	query := sqlFor(t, In("id", []string{"a", "b"}))

	assert.Contains(t, query, `"id" IN ('a', 'b')`)
}

// The column may be qualified, because two of the call sites filter inside an
// EXISTS subquery against another table.
func TestIn_KeepsAQualifiedColumn(t *testing.T) {
	query := sqlFor(t, In("user_groups.group_id", []string{"a"}))

	assert.Contains(t, query, `"user_groups"."group_id" IN ('a')`)
}
