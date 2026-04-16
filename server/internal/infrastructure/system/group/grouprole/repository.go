package grouprole

import (
	"context"

	"github.com/linzhengen/hub/v1/server/internal/domain/system/group/grouprole"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence/mysql"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence/mysql/sqlc"
)

func New(q *sqlc.Queries) grouprole.Repository {
	return &repositoryImpl{q: q}
}

type repositoryImpl struct {
	q *sqlc.Queries
}

func (r repositoryImpl) FindByGroupId(ctx context.Context, groupId string) (grouprole.GroupRoles, error) {
	rows, err := mysql.GetQ(ctx, r.q).SelectGroupRoleByGroupId(ctx, groupId)
	if err != nil {
		return nil, err
	}
	var result grouprole.GroupRoles
	for _, row := range rows {
		result = append(result, &grouprole.GroupRole{
			GroupId: row.GroupID,
			RoleId:  row.RoleID,
		})
	}
	return result, nil
}

func (r repositoryImpl) AssignRole(ctx context.Context, groupId, roleId string) error {
	return mysql.GetQ(ctx, r.q).CreateGroupRole(ctx, sqlc.CreateGroupRoleParams{
		GroupID: groupId,
		RoleID:  roleId,
	})
}

func (r repositoryImpl) UnassignRole(ctx context.Context, groupId, roleId string) error {
	return mysql.GetQ(ctx, r.q).DeleteGroupRole(ctx, sqlc.DeleteGroupRoleParams{
		GroupID: groupId,
		RoleID:  roleId,
	})
}

func (r repositoryImpl) Upsert(ctx context.Context, groupId string, roleId []string) error {
	if err := mysql.GetQ(ctx, r.q).DeleteGroupAllRole(ctx, groupId); err != nil {
		return err
	}
	for _, id := range roleId {
		if err := mysql.GetQ(ctx, r.q).CreateGroupRole(ctx, sqlc.CreateGroupRoleParams{
			GroupID: groupId,
			RoleID:  id,
		}); err != nil {
			return err
		}
	}
	return nil
}
