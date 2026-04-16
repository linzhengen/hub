package auth

import (
	"context"

	"github.com/linzhengen/hub/v1/server/internal/domain/auth"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence/mysql"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence/mysql/sqlc"
)

// Repository is the implementation of the auth.Repository interface
type Repository struct {
	q *sqlc.Queries
}

// NewRepository creates a new auth repository
func NewRepository(q *sqlc.Queries) auth.Repository {
	return &Repository{
		q: q,
	}
}

func (r *Repository) FindUserAuthorizedPolicies(ctx context.Context, userId string) ([]auth.Policy, error) {
	rows, err := mysql.GetQ(ctx, r.q).SelectUserAuthorizedPolices(ctx, userId)
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
