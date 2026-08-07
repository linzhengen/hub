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
		if matchString(policy.Object, req.Object) && matchString(policy.Action, req.Action) {
			return true, nil
		}
	}
	return false, nil
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
