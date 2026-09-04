package auth

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Service defines the interface for authorization service
type Service interface {
	// Enforce checks if a subject has permission to perform an action on an object
	Enforce(ctx context.Context, req Request) (bool, error)
	// Explain returns every route by which the subject is allowed the request,
	// empty when they are not allowed it at all. It answers the question
	// Enforce answers, and shows its working.
	Explain(ctx context.Context, req Request) ([]AccessPath, error)
	// PrincipalsFor returns the users allowed the request, each with the routes
	// that allow them. It reads the graph rather than asking Enforce once per
	// user, so it costs the same whether one person is allowed or everybody is.
	PrincipalsFor(ctx context.Context, req Request) ([]Principal, error)
	// VisibleOrgs returns the organizations a subject may be shown data from,
	// narrowed to orgId when they named one.
	//
	// It answers for a listing what Enforce answers for an operation, from the
	// same policies - so it is served by the same cache, and a grant that has
	// lapsed disappears from both at the same moment.
	VisibleOrgs(ctx context.Context, subject, orgId string) (Scope, error)
}

type service struct {
	authRepo Repository
	now      func() time.Time
}

// NewService creates a new authorization service
func NewService(authRepo Repository) Service {
	return NewServiceWithClock(authRepo, time.Now)
}

// NewServiceWithClock is NewService with the clock supplied, so a test can put
// a decision either side of an expiry without sleeping. The clock is read at
// the moment of the decision, never when the policies were fetched: that is
// what stops a warm cache from serving a grant that has since lapsed.
func NewServiceWithClock(authRepo Repository, now func() time.Time) Service {
	return &service{
		authRepo: authRepo,
		now:      now,
	}
}

// Enforce checks if a subject has permission to perform an action on an object
func (s *service) Enforce(ctx context.Context, req Request) (bool, error) {
	policies, err := s.authRepo.FindUserAuthorizedPolicies(ctx, req.Subject)
	if err != nil {
		return false, fmt.Errorf("failed to get user polices: %w", err)
	}
	now := s.now()
	for _, policy := range policies {
		if allows(policy.grant(), req, now) {
			return true, nil
		}
	}
	return false, nil
}

// VisibleOrgs returns the organizations the subject may be shown data from.
//
// A subject holding a grant through the platform organization sees everything,
// which is what keeps a single-tenant installation reading exactly as it did
// before organizations existed: every group the seed makes is the platform's.
//
// Naming an organization narrows the answer to that one, and to nothing at all
// when the subject holds no grant there. That is deliberately not an error: a
// caller asking about a tenant they cannot see is answered with an empty list,
// not told whether it exists.
func (s *service) VisibleOrgs(ctx context.Context, subject, orgId string) (Scope, error) {
	policies, err := s.authRepo.FindUserAuthorizedPolicies(ctx, subject)
	if err != nil {
		return Scope{}, fmt.Errorf("failed to get user polices: %w", err)
	}

	now := s.now()
	scope := Scope{}
	seen := map[string]bool{}
	for _, policy := range policies {
		if expired(policy.ExpiresAt, now) {
			continue
		}
		if policy.OrgWide {
			scope.All = true
			continue
		}
		if policy.OrgId != "" && !seen[policy.OrgId] {
			seen[policy.OrgId] = true
			scope.OrgIds = append(scope.OrgIds, policy.OrgId)
		}
	}

	if orgId == "" {
		return scope, nil
	}
	if scope.All || seen[orgId] {
		return Scope{OrgIds: []string{orgId}}, nil
	}
	return Scope{}, nil
}

// Explain returns every route by which the subject is allowed the request.
func (s *service) Explain(ctx context.Context, req Request) ([]AccessPath, error) {
	paths, err := s.authRepo.FindUserAccessPaths(ctx, req.Subject)
	if err != nil {
		return nil, fmt.Errorf("failed to get user access paths: %w", err)
	}
	return matching(paths, req, s.now()), nil
}

// PrincipalsFor returns the users allowed the request.
func (s *service) PrincipalsFor(ctx context.Context, req Request) ([]Principal, error) {
	paths, err := s.authRepo.FindAccessPaths(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access paths: %w", err)
	}

	// Narrow the graph before the people: a group that grants nothing relevant
	// needs none of its members looked at.
	now := s.now()
	byGroup := map[string][]AccessPath{}
	for _, path := range matching(paths, req, now) {
		byGroup[path.GroupId] = append(byGroup[path.GroupId], path)
	}
	if len(byGroup) == 0 {
		return nil, nil
	}

	memberships, err := s.authRepo.FindMemberships(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get memberships: %w", err)
	}

	// A user in two granting groups is one principal holding both routes, not
	// two principals, so they are gathered by user id before being returned.
	principals := map[string]*Principal{}
	var order []string
	for _, membership := range memberships {
		granted, ok := byGroup[membership.GroupId]
		if !ok || expired(membership.ExpiresAt, now) {
			continue
		}
		// A route is live only while every edge of it is, and the membership is
		// an edge the path query never saw. Reporting the earlier of the two is
		// what makes the answer say when the access really ends.
		granted = withExpiry(granted, membership.ExpiresAt)
		principal, seen := principals[membership.UserId]
		if !seen {
			principal = &Principal{UserId: membership.UserId, Username: membership.Username}
			principals[membership.UserId] = principal
			order = append(order, membership.UserId)
		}
		principal.Paths = append(principal.Paths, granted...)
	}

	result := make([]Principal, 0, len(order))
	for _, userId := range order {
		result = append(result, *principals[userId])
	}
	return result, nil
}

// matching keeps the paths that allow req, by the rule Enforce decides with.
func matching(paths []AccessPath, req Request, now time.Time) []AccessPath {
	var allowed []AccessPath
	for _, path := range paths {
		if allows(path.grant(), req, now) {
			allowed = append(allowed, path)
		}
	}
	return allowed
}

// withExpiry brings each path forward to expiry when that is the earlier of the
// two. A nil expiry is "never", so it never brings anything forward.
func withExpiry(paths []AccessPath, expiry *time.Time) []AccessPath {
	if expiry == nil {
		return paths
	}
	narrowed := make([]AccessPath, 0, len(paths))
	for _, path := range paths {
		path.ExpiresAt = Earliest(path.ExpiresAt, expiry)
		narrowed = append(narrowed, path)
	}
	return narrowed
}

// expired reports whether a grant has lapsed by now. A nil expiry never has.
func expired(expiresAt *time.Time, now time.Time) bool {
	return expiresAt != nil && !now.Before(*expiresAt)
}

// allows reports whether a grant covers req: it has not lapsed, it reaches the
// organization asked about, and its patterns cover the resource and the action.
//
// Enforce, Explain and PrincipalsFor all decide through this one function, so
// an answer and its explanation cannot disagree about what a pattern means.
// Splitting them would let `api.*` be enforced one way and explained another,
// and the explanation is only worth having if it is the same rule.
func allows(g grant, req Request, now time.Time) bool {
	return !expired(g.ExpiresAt, now) &&
		g.reaches(req.OrgId) &&
		matchString(g.Object, req.Object) &&
		matchString(g.Action, req.Action)
}

// matchString reports whether str satisfies a policy pattern. `*` stands for
// any run of characters and may appear anywhere, any number of times, so
// `api.*`, `api.system.*.v1.*Service` and a literal name are all valid
// patterns. Everything else must match exactly.
func matchString(pattern, str string) bool {
	segments := strings.Split(pattern, "*")
	if len(segments) == 1 {
		return pattern == str
	}

	// The text before the first `*` and after the last one is anchored; the
	// segments in between may appear anywhere, in order.
	if !strings.HasPrefix(str, segments[0]) {
		return false
	}
	rest := str[len(segments[0]):]

	last := segments[len(segments)-1]
	for _, segment := range segments[1 : len(segments)-1] {
		i := strings.Index(rest, segment)
		if i < 0 {
			return false
		}
		rest = rest[i+len(segment):]
	}
	return len(rest) >= len(last) && strings.HasSuffix(rest, last)
}
