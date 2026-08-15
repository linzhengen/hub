package grouprole

import (
	"context"
	"database/sql"
	"time"

	"github.com/linzhengen/hub/server/internal/domain/system/group/grouprole"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence"
)

func New(q persistence.Querier) grouprole.Repository {
	return &repositoryImpl{q: q}
}

type repositoryImpl struct {
	q persistence.Querier
}

func (r repositoryImpl) FindByGroupId(ctx context.Context, groupId string) (grouprole.GroupRoles, error) {
	rows, err := persistence.GetQ(ctx, r.q).SelectGroupRoleByGroupId(ctx, groupId)
	if err != nil {
		return nil, err
	}
	var result grouprole.GroupRoles
	for _, row := range rows {
		expiresAt := (*time.Time)(nil)
		if row.ExpiresAt.Valid {
			expiresAt = &row.ExpiresAt.Time
		}
		result = append(result, &grouprole.GroupRole{
			GroupId:   row.GroupID,
			RoleId:    row.RoleID,
			ExpiresAt: expiresAt,
		})
	}
	return result, nil
}

func (r repositoryImpl) AssignRole(ctx context.Context, groupId, roleId string, expiresAt *time.Time) error {
	expiry := sql.NullTime{}
	if expiresAt != nil {
		expiry = sql.NullTime{Time: *expiresAt, Valid: true}
	}
	return persistence.GetQ(ctx, r.q).CreateGroupRole(ctx, groupId, roleId, expiry)
}

func (r repositoryImpl) UnassignRole(ctx context.Context, groupId, roleId string) error {
	return persistence.GetQ(ctx, r.q).DeleteGroupRole(ctx, groupId, roleId)
}

func (r repositoryImpl) Upsert(ctx context.Context, groupId string, roleId []string) error {
	q := persistence.GetQ(ctx, r.q)
	if err := q.DeleteGroupAllRole(ctx, groupId); err != nil {
		return err
	}
	for _, id := range roleId {
		if err := q.CreateGroupRole(ctx, groupId, id, sql.NullTime{}); err != nil {
			return err
		}
	}
	return nil
}
