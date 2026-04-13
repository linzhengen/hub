package rolepermission

import (
	"context"

	"github.com/linzhengen/hub/server/internal/domain/system/role/rolepermission"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence/mysql"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence/mysql/sqlc"
)

func New(q *sqlc.Queries) rolepermission.Repository {
	return &repositoryImpl{q: q}
}

type repositoryImpl struct {
	q *sqlc.Queries
}

func (r repositoryImpl) FindByRoleId(ctx context.Context, roleId string) (rolepermission.RolePermissions, error) {
	rows, err := mysql.GetQ(ctx, r.q).SelectRolePermissionByRoleId(ctx, roleId)
	if err != nil {
		return nil, err
	}
	var result rolepermission.RolePermissions
	for _, row := range rows {
		result = append(result, &rolepermission.RolePermission{
			RoleId:       row.RoleID,
			PermissionId: row.PermissionID,
		})
	}
	return result, nil
}

func (r repositoryImpl) AssignPermission(ctx context.Context, roleId, permissionId string) error {
	return mysql.GetQ(ctx, r.q).AddPermissionToRole(ctx, sqlc.AddPermissionToRoleParams{
		RoleID:       roleId,
		PermissionID: permissionId,
	})
}

func (r repositoryImpl) UnassignPermission(ctx context.Context, roleId, permissionId string) error {
	return mysql.GetQ(ctx, r.q).RemovePermissionFromRole(ctx, sqlc.RemovePermissionFromRoleParams{
		RoleID:       roleId,
		PermissionID: permissionId,
	})
}

func (r repositoryImpl) Upsert(ctx context.Context, roleId string, permissionId []string) error {
	err := mysql.GetQ(ctx, r.q).DeleteRoleAllPermission(ctx, roleId)
	if err != nil {
		return err
	}
	for _, id := range permissionId {
		if err = mysql.GetQ(ctx, r.q).AddPermissionToRole(ctx, sqlc.AddPermissionToRoleParams{
			RoleID:       roleId,
			PermissionID: id,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r repositoryImpl) IsPermissionInRole(ctx context.Context, roleId string, permissionId string) (bool, error) {
	return mysql.GetQ(ctx, r.q).IsPermissionInRole(ctx, sqlc.IsPermissionInRoleParams{
		RoleID:       roleId,
		PermissionID: permissionId,
	})
}
