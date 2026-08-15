package serviceaccount

import (
	"context"

	domain "github.com/linzhengen/hub/server/internal/domain/system/serviceaccount"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence"
)

func New(q persistence.Querier) domain.Repository {
	return &repositoryImpl{q: q}
}

type repositoryImpl struct {
	q persistence.Querier
}

func (r repositoryImpl) Create(ctx context.Context, s *domain.ServiceAccount) error {
	return persistence.GetQ(ctx, r.q).CreateServiceAccount(
		ctx,
		s.Id,
		s.UserId,
		s.Name,
		s.Description,
		s.ClientId,
		s.KeycloakId,
		s.CreatedByUserId,
	)
}

func (r repositoryImpl) FindOne(ctx context.Context, id string) (*domain.ServiceAccount, error) {
	row, err := persistence.GetQ(ctx, r.q).SelectServiceAccount(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert(row), nil
}

func (r repositoryImpl) List(
	ctx context.Context,
	params domain.ListParams,
) ([]*domain.ServiceAccount, int64, error) {
	q := persistence.GetQ(ctx, r.q)

	total, err := q.CountServiceAccounts(ctx)
	if err != nil {
		return nil, 0, err
	}

	rows, err := q.ListServiceAccounts(
		ctx,
		int32(params.Offset), //nolint:gosec // bounded by the pagination rule
		int32(params.Limit),  //nolint:gosec // bounded by the pagination rule
	)
	if err != nil {
		return nil, 0, err
	}

	accounts := make([]*domain.ServiceAccount, 0, len(rows))
	for _, row := range rows {
		accounts = append(accounts, convert(row))
	}
	return accounts, total, nil
}

func (r repositoryImpl) Delete(ctx context.Context, id string) error {
	return persistence.GetQ(ctx, r.q).DeleteServiceAccount(ctx, id)
}

func convert(m *persistence.ServiceAccountModel) *domain.ServiceAccount {
	return &domain.ServiceAccount{
		Id:              m.ID,
		UserId:          m.UserID,
		Name:            m.Name,
		Description:     m.Description,
		ClientId:        m.ClientID,
		KeycloakId:      m.KeycloakID,
		CreatedByUserId: m.CreatedByUserID,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}
