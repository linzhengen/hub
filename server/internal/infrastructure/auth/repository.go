package auth

import (
	"context"

	"github.com/linzhengen/hub/server/internal/domain/auth"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence"
)

// Repository is the implementation of the auth.Repository interface
type Repository struct {
	q persistence.Querier
}

// NewRepository creates a new auth repository
func NewRepository(q persistence.Querier) auth.Repository {
	return &Repository{
		q: q,
	}
}

func (r *Repository) FindUserAuthorizedPolicies(ctx context.Context, userId string) ([]auth.Policy, error) {
	rows, err := persistence.GetQ(ctx, r.q).SelectUserAuthorizedPolicies(ctx, userId)
	if err != nil {
		return nil, err
	}

	var policies []auth.Policy
	for _, row := range rows {
		policies = append(policies, auth.Policy{
			Subject: row.ID,
			Object:  row.Identifier,
			Action:  row.Verb,
		})
	}

	return policies, nil
}

func (r *Repository) Revision(ctx context.Context) (int64, error) {
	return persistence.GetQ(ctx, r.q).SelectRbacRevision(ctx)
}
