package auth

import (
	"context"
	"sync"
	"time"

	"github.com/linzhengen/hub/v1/server/internal/domain/auth"
)

// cachingRepository memoises a user's effective policies for a short window.
//
// Every guarded request runs SelectUserAuthorizedPolicies, a seven-way join
// over users, groups, roles, permissions and resources, so a busy client pays
// for that join once per call even though its answer changes only when an
// administrator edits the RBAC graph. Caching it trades a bounded amount of
// staleness for that work: a permission granted or revoked during the window
// takes effect once the entry expires.
//
// Because the trade is a security one it is opt-in — NewCachingRepository
// returns the underlying repository unchanged when ttl is not positive.
type cachingRepository struct {
	inner auth.Repository
	ttl   time.Duration
	now   func() time.Time

	mu      sync.RWMutex
	entries map[string]cacheEntry
}

type cacheEntry struct {
	policies  []auth.Policy
	expiresAt time.Time
}

// NewCachingRepository wraps repo so policy lookups are reused for ttl. A ttl
// of zero or less disables caching and returns repo itself.
func NewCachingRepository(repo auth.Repository, ttl time.Duration) auth.Repository {
	if ttl <= 0 {
		return repo
	}
	return &cachingRepository{
		inner:   repo,
		ttl:     ttl,
		now:     time.Now,
		entries: map[string]cacheEntry{},
	}
}

func (r *cachingRepository) FindUserAuthorizedPolicies(ctx context.Context, userId string) ([]auth.Policy, error) {
	if policies, ok := r.lookup(userId); ok {
		return policies, nil
	}

	policies, err := r.inner.FindUserAuthorizedPolicies(ctx, userId)
	if err != nil {
		return nil, err
	}
	r.store(userId, policies)
	return policies, nil
}

func (r *cachingRepository) lookup(userId string) ([]auth.Policy, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, ok := r.entries[userId]
	if !ok || !r.now().Before(entry.expiresAt) {
		return nil, false
	}
	return entry.policies, true
}

func (r *cachingRepository) store(userId string, policies []auth.Policy) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.now()
	// Expired entries are only dropped on write, which keeps reads lock-cheap
	// and bounds the map by the number of users seen within one ttl.
	for id, entry := range r.entries {
		if !now.Before(entry.expiresAt) {
			delete(r.entries, id)
		}
	}
	r.entries[userId] = cacheEntry{policies: policies, expiresAt: now.Add(r.ttl)}
}
