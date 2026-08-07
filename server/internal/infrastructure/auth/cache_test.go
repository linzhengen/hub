package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/linzhengen/hub/v1/server/internal/domain/auth"
)

type countingRepository struct {
	calls    map[string]int
	policies []auth.Policy
	err      error
}

func newCountingRepository(policies ...auth.Policy) *countingRepository {
	return &countingRepository{calls: map[string]int{}, policies: policies}
}

func (r *countingRepository) FindUserAuthorizedPolicies(_ context.Context, userId string) ([]auth.Policy, error) {
	r.calls[userId]++
	if r.err != nil {
		return nil, r.err
	}
	return r.policies, nil
}

func TestNewCachingRepositoryWithoutTTLReturnsTheRepositoryUnchanged(t *testing.T) {
	inner := newCountingRepository()

	assert.Same(t, inner, NewCachingRepository(inner, 0))
	assert.Same(t, inner, NewCachingRepository(inner, -time.Second))
}

func TestCachingRepositoryReusesPoliciesWithinTheTTL(t *testing.T) {
	ctx := context.Background()
	policy := auth.Policy{Subject: "user1", Object: "api.*", Action: "*"}
	inner := newCountingRepository(policy)

	clock := time.Now()
	repo := NewCachingRepository(inner, time.Minute).(*cachingRepository)
	repo.now = func() time.Time { return clock }

	for range 3 {
		policies, err := repo.FindUserAuthorizedPolicies(ctx, "user1")
		require.NoError(t, err)
		assert.Equal(t, []auth.Policy{policy}, policies)
	}
	assert.Equal(t, 1, inner.calls["user1"])

	// A second user is a separate entry, not a cache hit on the first.
	_, err := repo.FindUserAuthorizedPolicies(ctx, "user2")
	require.NoError(t, err)
	assert.Equal(t, 1, inner.calls["user2"])

	clock = clock.Add(time.Minute)
	_, err = repo.FindUserAuthorizedPolicies(ctx, "user1")
	require.NoError(t, err)
	assert.Equal(t, 2, inner.calls["user1"], "an expired entry is re-read")
}

func TestCachingRepositoryDoesNotCacheFailures(t *testing.T) {
	ctx := context.Background()
	inner := newCountingRepository()
	inner.err = errors.New("db unavailable")

	repo := NewCachingRepository(inner, time.Minute)

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

	clock := time.Now()
	repo := NewCachingRepository(inner, time.Minute).(*cachingRepository)
	repo.now = func() time.Time { return clock }

	_, err := repo.FindUserAuthorizedPolicies(ctx, "user1")
	require.NoError(t, err)

	clock = clock.Add(2 * time.Minute)
	_, err = repo.FindUserAuthorizedPolicies(ctx, "user2")
	require.NoError(t, err)

	repo.mu.RLock()
	defer repo.mu.RUnlock()
	assert.NotContains(t, repo.entries, "user1")
	assert.Contains(t, repo.entries, "user2")
}
