package auth

import (
	"context"
	"fmt"
	"strings"
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
}

type service struct {
	authRepo Repository
}

// NewService creates a new authorization service
func NewService(authRepo Repository) Service {
	return &service{
		authRepo: authRepo,
	}
}

// Enforce checks if a subject has permission to perform an action on an object
func (s *service) Enforce(ctx context.Context, req Request) (bool, error) {
	policies, err := s.authRepo.FindUserAuthorizedPolicies(ctx, req.Subject)
	if err != nil {
		return false, fmt.Errorf("failed to get user polices: %w", err)
	}
	for _, policy := range policies {
		if allows(policy.Object, policy.Action, req) {
			return true, nil
		}
	}
	return false, nil
}

// Explain returns every route by which the subject is allowed the request.
func (s *service) Explain(ctx context.Context, req Request) ([]AccessPath, error) {
	paths, err := s.authRepo.FindUserAccessPaths(ctx, req.Subject)
	if err != nil {
		return nil, fmt.Errorf("failed to get user access paths: %w", err)
	}
	return matching(paths, req), nil
}

// PrincipalsFor returns the users allowed the request.
func (s *service) PrincipalsFor(ctx context.Context, req Request) ([]Principal, error) {
	paths, err := s.authRepo.FindAccessPaths(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access paths: %w", err)
	}

	// Narrow the graph before the people: a group that grants nothing relevant
	// needs none of its members looked at.
	byGroup := map[string][]AccessPath{}
	for _, path := range matching(paths, req) {
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
		if !ok {
			continue
		}
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
func matching(paths []AccessPath, req Request) []AccessPath {
	var allowed []AccessPath
	for _, path := range paths {
		if allows(path.Object, path.Action, req) {
			allowed = append(allowed, path)
		}
	}
	return allowed
}

// allows reports whether a grant of object and action covers req.
//
// Enforce, Explain and PrincipalsFor all decide through this one function, so
// an answer and its explanation cannot disagree about what a pattern means.
// Splitting them would let `api.*` be enforced one way and explained another,
// and the explanation is only worth having if it is the same rule.
func allows(object, action string, req Request) bool {
	return matchString(object, req.Object) && matchString(action, req.Action)
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
