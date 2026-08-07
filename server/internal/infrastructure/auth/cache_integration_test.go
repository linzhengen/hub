package auth

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	// Registers the "postgres" driver. The same import gives cmd/cli its
	// migration driver, so this adds no dependency.
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence/postgres/sqlc"
)

// The cache's correctness rests on database triggers bumping a counter, which
// no unit test can check: a mock repository would simply do what it is told.
// This exercises the real triggers.
//
// It needs a migrated PostgreSQL, so it is skipped unless HUB_TEST_DSN is set:
//
//	make dev
//	HUB_TEST_DSN='postgres://postgres:password%23123@localhost:56836/hub?sslmode=disable' \
//	  go test ./internal/infrastructure/auth/ -run Live
func TestLiveRevisionInvalidation(t *testing.T) {
	dsn := os.Getenv("HUB_TEST_DSN")
	if dsn == "" {
		t.Skip("set HUB_TEST_DSN to run this against a migrated database")
	}

	ctx := context.Background()
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	require.NoError(t, db.PingContext(ctx))

	userID := seedRBACGraph(t, db)

	clock := time.Now()
	repo, ok := NewCachingRepository(NewRepository(persistence.NewPostgreSQLQuerier(sqlc.New(db))), time.Hour).(*cachingRepository)
	require.True(t, ok)
	repo.now = func() time.Time { return clock }

	before, err := repo.FindUserAuthorizedPolicies(ctx, userID)
	require.NoError(t, err)
	require.Len(t, before, 1)

	// Grant a second permission. Nothing tells the cache; the trigger does.
	_, err = db.ExecContext(ctx,
		`INSERT INTO permissions (id, verb, resource_id, description) VALUES ($1, 'Extra', $2, '')`,
		"perm-live-2", "res-live")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx,
		`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`, "role-live", "perm-live-2")
	require.NoError(t, err)

	// The TTL is an hour, so without the revision check this would never
	// notice. Within the poll interval it deliberately does not.
	same, err := repo.FindUserAuthorizedPolicies(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, same, 1, "inside the poll interval the cached answer stands")

	clock = clock.Add(revisionPollInterval)

	after, err := repo.FindUserAuthorizedPolicies(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, after, 2, "past the poll interval the new grant is visible despite the hour-long TTL")

	// A revocation must land the same way.
	_, err = db.ExecContext(ctx, `DELETE FROM permissions WHERE id = $1`, "perm-live-2")
	require.NoError(t, err)
	clock = clock.Add(revisionPollInterval)

	revoked, err := repo.FindUserAuthorizedPolicies(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, revoked, 1, "a revocation is not left in place until the TTL expires")
}

// seedRBACGraph builds a complete user -> group -> role -> permission ->
// resource chain and removes it afterwards. Deleting the resource and the
// group cascades to everything hanging off them.
func seedRBACGraph(t *testing.T, db *sql.DB) string {
	t.Helper()
	ctx := context.Background()

	statements := []struct {
		query string
		args  []any
	}{
		{`INSERT INTO resources (id, parent_id, name, identifier, type, path, display_order, status)
		  VALUES ($1, '', 'live test', 'api.live.test', 'api', '', 0, 'Active')`, []any{"res-live"}},
		{`INSERT INTO permissions (id, verb, resource_id, description) VALUES ($1, 'Read', $2, '')`,
			[]any{"perm-live-1", "res-live"}},
		{`INSERT INTO roles (id, name, description) VALUES ($1, 'live test', '')`, []any{"role-live"}},
		{`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`, []any{"role-live", "perm-live-1"}},
		{`INSERT INTO groups (id, name, description, status) VALUES ($1, 'live test', '', 'Active')`, []any{"group-live"}},
		{`INSERT INTO group_roles (group_id, role_id) VALUES ($1, $2)`, []any{"group-live", "role-live"}},
		{`INSERT INTO users (id, username, email, status) VALUES ($1, 'live test', 'live@test.invalid', 'Active')`,
			[]any{"user-live"}},
		{`INSERT INTO user_groups (user_id, group_id) VALUES ($1, $2)`, []any{"user-live", "group-live"}},
	}

	cleanup := func() {
		for _, query := range []string{
			`DELETE FROM user_groups WHERE user_id = 'user-live'`,
			`DELETE FROM users WHERE id = 'user-live'`,
			`DELETE FROM groups WHERE id = 'group-live'`,
			`DELETE FROM roles WHERE id = 'role-live'`,
			`DELETE FROM resources WHERE id = 'res-live'`,
		} {
			_, _ = db.ExecContext(ctx, query)
		}
	}
	cleanup()
	t.Cleanup(cleanup)

	for _, s := range statements {
		_, err := db.ExecContext(ctx, s.query, s.args...)
		require.NoError(t, err, s.query)
	}
	return "user-live"
}
