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

func matchString(pattern, str string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(str, prefix)
	}
	return pattern == str
}
