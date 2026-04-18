package rolepermission

import (
	"context"

	"github.com/linzhengen/hub/v1/server/internal/domain/system/role/rolepermission"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence"
)

func New(q persistence.Querier) rolepermission.Repository {
	return &repositoryImpl{q: q}
}

type repositoryImpl struct {
	q persistence.Querier
}

func (r repositoryImpl) FindByRoleId(ctx context.Context, roleId string) (rolepermission.RolePermissions, error) {
	rows, err := persistence.GetQ(ctx, r.q).SelectRolePermissionByRoleId(ctx, roleId)
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
	return persistence.GetQ(ctx, r.q).AddPermissionToRole(ctx, roleId, permissionId)
}

func (r repositoryImpl) UnassignPermission(ctx context.Context, roleId, permissionId string) error {
	return persistence.GetQ(ctx, r.q).RemovePermissionFromRole(ctx, roleId, permissionId)
}

func (r repositoryImpl) Upsert(ctx context.Context, roleId string, permissionId []string) error {
	q := persistence.GetQ(ctx, r.q)
	err := q.DeleteRoleAllPermission(ctx, roleId)
	if err != nil {
		return err
	}
	for _, id := range permissionId {
		if err = q.AddPermissionToRole(ctx, roleId, id); err != nil {
			return err
		}
	}
	return nil
}

func (r repositoryImpl) IsPermissionInRole(ctx context.Context, roleId string, permissionId string) (bool, error) {
	return persistence.GetQ(ctx, r.q).IsPermissionInRole(ctx, roleId, permissionId)
}
