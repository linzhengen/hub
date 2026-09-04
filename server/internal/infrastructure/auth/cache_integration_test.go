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

	"github.com/linzhengen/hub/server/internal/infrastructure/persistence"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence/postgres/sqlc"
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
//
// The fixture ids are real uuids because every primary key is one. They used to
// be readable strings like "user-live", which stopped inserting when the keys
// moved to the uuid type - and nothing noticed, because the whole test is
// skipped without HUB_TEST_DSN. A test that only runs when somebody remembers
// to run it rots quietly, so keep it runnable.
const (
	liveResourceId     = "0f000000-0000-0000-0000-000000000001"
	livePermissionId   = "0f000000-0000-0000-0000-000000000002"
	liveExtraPermId    = "0f000000-0000-0000-0000-000000000003"
	liveRoleId         = "0f000000-0000-0000-0000-000000000004"
	liveGroupId        = "0f000000-0000-0000-0000-000000000005"
	liveUserId         = "0f000000-0000-0000-0000-000000000006"
	liveOrganizationId = "0f000000-0000-0000-0000-000000000007"
)

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
		liveExtraPermId, liveResourceId)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx,
		`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`, liveRoleId, liveExtraPermId)
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
	_, err = db.ExecContext(ctx, `DELETE FROM permissions WHERE id = $1`, liveExtraPermId)
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
		// parent_id is NULL rather than '': it is a nullable uuid, and a menu
		// resource at the top of the tree has no parent.
		{`INSERT INTO resources (id, parent_id, name, identifier, type, path, display_order, status)
		  VALUES ($1, NULL, 'live test', 'api.live.test', 'api', '', 0, 'Active')`, []any{liveResourceId}},
		{`INSERT INTO permissions (id, verb, resource_id, description) VALUES ($1, 'Read', $2, '')`,
			[]any{livePermissionId, liveResourceId}},
		{`INSERT INTO roles (id, name, description) VALUES ($1, 'live test', '')`, []any{liveRoleId}},
		{`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`, []any{liveRoleId, livePermissionId}},
		// The group is put in its own organization rather than the platform
		// one, so that the organization's own trigger is exercised below and
		// the policy carries a real organization rather than the one that
		// matches everything.
		{`INSERT INTO organizations (id, name, slug, kind, description, status)
		  VALUES ($1, 'live test', 'live-test', 'BUSINESS', '', 'Active')`, []any{liveOrganizationId}},
		{`INSERT INTO groups (id, name, description, status, org_id) VALUES ($1, 'live test', '', 'Active', $2)`,
			[]any{liveGroupId, liveOrganizationId}},
		{`INSERT INTO group_roles (group_id, role_id) VALUES ($1, $2)`, []any{liveGroupId, liveRoleId}},
		{`INSERT INTO users (id, username, email, status) VALUES ($1, 'live test', 'live@test.invalid', 'Active')`,
			[]any{liveUserId}},
		{`INSERT INTO user_groups (user_id, group_id) VALUES ($1, $2)`, []any{liveUserId, liveGroupId}},
	}

	cleanup := func() {
		for _, query := range []string{
			`DELETE FROM user_groups WHERE user_id = '` + liveUserId + `'`,
			`DELETE FROM users WHERE id = '` + liveUserId + `'`,
			`DELETE FROM groups WHERE id = '` + liveGroupId + `'`,
			`DELETE FROM organizations WHERE id = '` + liveOrganizationId + `'`,
			`DELETE FROM roles WHERE id = '` + liveRoleId + `'`,
			`DELETE FROM resources WHERE id = '` + liveResourceId + `'`,
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
	return liveUserId
}

// TestLiveOrganizationInvalidation is the same guarantee for the table this
// change added.
//
// An authorization decision reads `organizations` - a policy carries the
// organization its group belongs to, and the query drops an inactive one - so a
// write there can make a cached policy stale. Without a trigger of its own,
// deactivating a tenant would leave everyone in it working for up to the full
// TTL, which is exactly the failure rbac_revisions exists to prevent.
func TestLiveOrganizationInvalidation(t *testing.T) {
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
	assert.Equal(t, liveOrganizationId, before[0].OrgId,
		"a policy carries the organization of the group it was reached through")
	assert.False(t, before[0].OrgWide,
		"only the platform organization's grants answer about every organization")

	// Deactivating the tenant is not a write to any table the old triggers
	// watched. Nothing tells the cache; the new trigger does.
	_, err = db.ExecContext(ctx,
		`UPDATE organizations SET status = 'Inactive' WHERE id = $1`, liveOrganizationId)
	require.NoError(t, err)

	clock = clock.Add(revisionPollInterval)

	after, err := repo.FindUserAuthorizedPolicies(ctx, userID)
	require.NoError(t, err)
	assert.Empty(t, after, "deactivating an organization revokes the access held inside it")
}
