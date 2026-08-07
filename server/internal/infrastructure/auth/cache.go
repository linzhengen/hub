package auth

import (
	"context"
	"sync"
	"time"

	"github.com/linzhengen/hub/v1/server/internal/domain/auth"
)

// revisionPollInterval bounds how stale an authorization decision can be.
//
// The cache does not re-read the RBAC revision on every request - that would
// trade one query for another - but it never trusts a revision older than
// this. So an administrator's change reaches every replica within this window,
// whatever the cache TTL is.
const revisionPollInterval = time.Second

// cachingRepository memoises a user's effective policies until the RBAC graph
// changes underneath them.
//
// Every guarded request runs SelectUserAuthorizedPolicies, a seven-way join
// over users, groups, roles, permissions and resources, so a busy client pays
// for that join once per call even though its answer changes only when an
// administrator edits the graph.
//
// Time alone is a poor way to notice such an edit: a short TTL throws the
// cache away constantly, and a long one leaves a revoked user working. So the
// cache keys everything on the revision counter the database maintains -
// bumped by triggers on every table the decision reads - and drops the lot the
// moment that number moves. The TTL then only bounds how long an *unchanged*
// entry is kept, which is a memory question rather than a security one.
//
// The counter is shared, so this works across replicas: an edit on one is seen
// by the others within revisionPollInterval.
type cachingRepository struct {
	inner auth.Repository
	ttl   time.Duration
	now   func() time.Time

	mu       sync.RWMutex
	entries  map[string]cacheEntry
	revision int64
	// revisionReadAt is when revision was last read from the database. The
	// zero time means "never", which forces a read on the first lookup.
	revisionReadAt time.Time
}

type cacheEntry struct {
	policies  []auth.Policy
	expiresAt time.Time
}

// NewCachingRepository wraps repo so policy lookups are reused until either the
// RBAC graph changes or ttl elapses. A ttl of zero or less disables caching and
// returns repo itself.
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

func (r *cachingRepository) Revision(ctx context.Context) (int64, error) {
	return r.inner.Revision(ctx)
}

func (r *cachingRepository) FindUserAuthorizedPolicies(ctx context.Context, userId string) ([]auth.Policy, error) {
	// Fail closed: if we cannot tell whether the cache is current, do not use
	// it. Serving a cached decision against an unknown graph is exactly the
	// mistake this type exists to avoid.
	if err := r.syncRevision(ctx); err != nil {
		return r.inner.FindUserAuthorizedPolicies(ctx, userId)
	}

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

// syncRevision re-reads the RBAC revision if the last read has gone stale, and
// empties the cache when the number has moved.
func (r *cachingRepository) syncRevision(ctx context.Context) error {
	if r.revisionIsFresh() {
		return nil
	}

	revision, err := r.inner.Revision(ctx)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if revision != r.revision {
		r.entries = map[string]cacheEntry{}
		r.revision = revision
	}
	r.revisionReadAt = r.now()
	return nil
}

func (r *cachingRepository) revisionIsFresh() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return !r.revisionReadAt.IsZero() && r.now().Before(r.revisionReadAt.Add(revisionPollInterval))
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
