package auth

import "context"

type Repository interface {
	FindUserAuthorizedPolicies(ctx context.Context, userId string) ([]Policy, error)
	// Revision changes whenever anything an authorization decision reads
	// changes. It exists so a cache can ask "is what I hold still true?" in one
	// cheap query instead of repeating the permission join.
	Revision(ctx context.Context) (int64, error)
}
