package delegation

import (
	"context"
	"database/sql"
	"time"

	domain "github.com/linzhengen/hub/server/internal/domain/system/delegation"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence"
)

func New(q persistence.Querier) domain.Repository {
	return &repositoryImpl{q: q}
}

type repositoryImpl struct {
	q persistence.Querier
}

// Create writes the delegation and the permissions it carries.
//
// The two statements have to land together: a delegation row with no
// permissions grants nothing, so a half-written one is harmless, but it is
// still a row somebody has to explain. The use case runs this inside a
// transaction.
func (r repositoryImpl) Create(ctx context.Context, d *domain.Delegation) error {
	q := persistence.GetQ(ctx, r.q)

	if err := q.CreateDelegation(
		ctx,
		d.Id,
		d.AgentId,
		d.PrincipalUserId,
		d.GrantedByUserId,
		d.OrgId,
		d.Reason,
		int16(d.MaxDepth), //nolint:gosec // bounded by the request's max_depth rule
		nullTime(d.ExpiresAt),
	); err != nil {
		return err
	}

	for _, permissionId := range d.PermissionIds {
		if err := q.AddPermissionToDelegation(ctx, d.Id, permissionId); err != nil {
			return err
		}
	}
	return nil
}

func (r repositoryImpl) FindOne(ctx context.Context, id string) (*domain.Delegation, error) {
	q := persistence.GetQ(ctx, r.q)

	row, err := q.SelectDelegation(ctx, id)
	if err != nil {
		return nil, err
	}
	permissionIds, err := r.permissionIds(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert(row, permissionIds), nil
}

func (r repositoryImpl) List(
	ctx context.Context,
	params domain.ListParams,
) ([]*domain.Delegation, int64, error) {
	q := persistence.GetQ(ctx, r.q)

	total, err := q.CountDelegations(
		ctx,
		params.AgentId,
		params.PrincipalUserId,
		params.IncludeRevoked,
		params.OrgIds,
		params.SelfUserId,
	)
	if err != nil {
		return nil, 0, err
	}

	rows, err := q.ListDelegations(
		ctx,
		params.AgentId,
		params.PrincipalUserId,
		params.IncludeRevoked,
		params.OrgIds,
		params.SelfUserId,
		int32(params.Offset), //nolint:gosec // bounded by the pagination rule
		int32(params.Limit),  //nolint:gosec // bounded by the pagination rule
	)
	if err != nil {
		return nil, 0, err
	}

	delegations := make([]*domain.Delegation, 0, len(rows))
	for _, row := range rows {
		permissionIds, err := r.permissionIds(ctx, row.ID)
		if err != nil {
			return nil, 0, err
		}
		delegations = append(delegations, convert(row, permissionIds))
	}
	return delegations, total, nil
}

func (r repositoryImpl) Revoke(ctx context.Context, id string) (bool, error) {
	affected, err := persistence.GetQ(ctx, r.q).RevokeDelegation(ctx, id)
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (r repositoryImpl) permissionIds(ctx context.Context, delegationId string) ([]string, error) {
	rows, err := persistence.GetQ(ctx, r.q).SelectDelegationPermission(ctx, delegationId)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.PermissionID)
	}
	return ids, nil
}

func convert(m *persistence.DelegationModel, permissionIds []string) *domain.Delegation {
	return &domain.Delegation{
		Id:              m.ID,
		AgentId:         m.AgentID,
		PrincipalUserId: m.PrincipalUserID,
		GrantedByUserId: m.GrantedByUserID,
		OrgId:           m.OrgID,
		Reason:          m.Reason,
		PermissionIds:   permissionIds,
		MaxDepth:        uint32(m.MaxDepth), //nolint:gosec // the column is CHECKed positive
		ExpiresAt:       timePtr(m.ExpiresAt),
		RevokedAt:       timePtr(m.RevokedAt),
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func timePtr(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
