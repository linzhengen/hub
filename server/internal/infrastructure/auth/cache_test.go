package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/linzhengen/hub/server/internal/domain/auth"
)

type countingRepository struct {
	calls         map[string]int
	revisionCalls int
	policies      []auth.Policy
	revision      int64
	err           error
	revisionErr   error
}

func newCountingRepository(policies ...auth.Policy) *countingRepository {
	return &countingRepository{calls: map[string]int{}, policies: policies, revision: 1}
}

func (r *countingRepository) FindUserAuthorizedPolicies(_ context.Context, userId string) ([]auth.Policy, error) {
	r.calls[userId]++
	if r.err != nil {
		return nil, r.err
	}
	return r.policies, nil
}

func (r *countingRepository) Revision(_ context.Context) (int64, error) {
	r.revisionCalls++
	if r.revisionErr != nil {
		return 0, r.revisionErr
	}
	return r.revision, nil
}

// newCache returns a cache with a controllable clock, so the tests never sleep.
func newCache(t *testing.T, inner auth.Repository, ttl time.Duration) (*cachingRepository, *time.Time) {
	t.Helper()

	clock := time.Now()
	repo, ok := NewCachingRepository(inner, ttl).(*cachingRepository)
	require.True(t, ok)
	repo.now = func() time.Time { return clock }
	return repo, &clock
}

func TestNewCachingRepositoryWithoutTTLReturnsTheRepositoryUnchanged(t *testing.T) {
	inner := newCountingRepository()

	assert.Same(t, inner, NewCachingRepository(inner, 0))
	assert.Same(t, inner, NewCachingRepository(inner, -time.Second))
}

func TestCachingRepositoryReusesPoliciesWhileNothingChanges(t *testing.T) {
	ctx := context.Background()
	policy := auth.Policy{Subject: "user1", Object: "api.*", Action: "*"}
	inner := newCountingRepository(policy)
	repo, clock := newCache(t, inner, time.Minute)

	for range 3 {
		policies, err := repo.FindUserAuthorizedPolicies(ctx, "user1")
		require.NoError(t, err)
		assert.Equal(t, []auth.Policy{policy}, policies)
	}
	assert.Equal(t, 1, inner.calls["user1"], "the permission join runs once")
	assert.Equal(t, 1, inner.revisionCalls, "and the revision is not re-read within the poll interval")

	// A second user is a separate entry, not a cache hit on the first.
	_, err := repo.FindUserAuthorizedPolicies(ctx, "user2")
	require.NoError(t, err)
	assert.Equal(t, 1, inner.calls["user2"])

	*clock = clock.Add(time.Minute)
	_, err = repo.FindUserAuthorizedPolicies(ctx, "user1")
	require.NoError(t, err)
	assert.Equal(t, 2, inner.calls["user1"], "an expired entry is re-read")
}

// The point of the revision counter: a grant or revocation must take effect
// without waiting out the TTL.
func TestCachingRepositoryDropsEverythingWhenTheGraphChanges(t *testing.T) {
	ctx := context.Background()
	inner := newCountingRepository(auth.Policy{Object: "api.*", Action: "*"})
	repo, clock := newCache(t, inner, time.Hour)

	_, err := repo.FindUserAuthorizedPolicies(ctx, "user1")
	require.NoError(t, err)
	_, err = repo.FindUserAuthorizedPolicies(ctx, "user2")
	require.NoError(t, err)
	assert.Equal(t, 1, inner.calls["user1"])
	assert.Equal(t, 1, inner.calls["user2"])

	// An administrator edits the RBAC graph; a trigger bumps the counter.
	inner.revision = 2
	*clock = clock.Add(revisionPollInterval)

	_, err = repo.FindUserAuthorizedPolicies(ctx, "user1")
	require.NoError(t, err)
	assert.Equal(t, 2, inner.calls["user1"], "the entry is re-read against the new graph")

	// Not just the user who happened to ask: the change may have been to a
	// role or a group, which we cannot attribute to one user.
	_, err = repo.FindUserAuthorizedPolicies(ctx, "user2")
	require.NoError(t, err)
	assert.Equal(t, 2, inner.calls["user2"])
}

// Between polls the revision is taken on trust, which is what keeps the check
// from costing a query per request. That is the staleness bound.
func TestCachingRepositoryDoesNotRereadTheRevisionWithinThePollInterval(t *testing.T) {
	ctx := context.Background()
	inner := newCountingRepository()
	repo, clock := newCache(t, inner, time.Hour)

	_, err := repo.FindUserAuthorizedPolicies(ctx, "user1")
	require.NoError(t, err)

	inner.revision = 2
	*clock = clock.Add(revisionPollInterval - time.Millisecond)

	_, err = repo.FindUserAuthorizedPolicies(ctx, "user1")
	require.NoError(t, err)
	assert.Equal(t, 1, inner.calls["user1"])
	assert.Equal(t, 1, inner.revisionCalls)

	*clock = clock.Add(time.Millisecond)

	_, err = repo.FindUserAuthorizedPolicies(ctx, "user1")
	require.NoError(t, err)
	assert.Equal(t, 2, inner.revisionCalls)
	assert.Equal(t, 2, inner.calls["user1"])
}

// If we cannot tell whether the cache is current, we must not use it.
func TestCachingRepositoryBypassesTheCacheWhenTheRevisionIsUnreadable(t *testing.T) {
	ctx := context.Background()
	policy := auth.Policy{Object: "api.*", Action: "*"}
	inner := newCountingRepository(policy)
	repo, clock := newCache(t, inner, time.Hour)

	_, err := repo.FindUserAuthorizedPolicies(ctx, "user1")
	require.NoError(t, err)
	assert.Equal(t, 1, inner.calls["user1"])

	inner.revisionErr = errors.New("db unavailable")
	*clock = clock.Add(revisionPollInterval)

	policies, err := repo.FindUserAuthorizedPolicies(ctx, "user1")
	require.NoError(t, err)
	assert.Equal(t, []auth.Policy{policy}, policies)
	assert.Equal(t, 2, inner.calls["user1"], "the join runs rather than trusting the cache")
}

func TestCachingRepositoryDoesNotCacheFailures(t *testing.T) {
	ctx := context.Background()
	inner := newCountingRepository()
	inner.err = errors.New("db unavailable")
	repo, _ := newCache(t, inner, time.Minute)

	for range 2 {
		_, err := repo.FindUserAuthorizedPolicies(ctx, "user1")
		assert.Error(t, err)
	}
	assert.Equal(t, 2, inner.calls["user1"])
}

// Entries are swept on write, so a cache serving many users does not grow
// without bound once they stop making requests.
func TestCachingRepositoryEvictsExpiredEntriesOnWrite(t *testing.T) {
	ctx := context.Background()
	inner := newCountingRepository()
	repo, clock := newCache(t, inner, time.Minute)

	_, err := repo.FindUserAuthorizedPolicies(ctx, "user1")
	require.NoError(t, err)

	*clock = clock.Add(2 * time.Minute)
	_, err = repo.FindUserAuthorizedPolicies(ctx, "user2")
	require.NoError(t, err)

	repo.mu.RLock()
	defer repo.mu.RUnlock()
	assert.NotContains(t, repo.entries, "user1")
	assert.Contains(t, repo.entries, "user2")
}

func TestCachingRepositoryPassesTheRevisionThrough(t *testing.T) {
	inner := newCountingRepository()
	inner.revision = 7
	repo, _ := newCache(t, inner, time.Minute)

	revision, err := repo.Revision(context.Background())

	require.NoError(t, err)
	assert.Equal(t, int64(7), revision)
}
