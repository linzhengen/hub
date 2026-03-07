package auth

import "context"

type Repository interface {
	FindUserAuthorizedPolicies(ctx context.Context, userId string) ([]Policy, error)
}
