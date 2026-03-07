package role

import (
	"context"

	"github.com/linzhengen/hub/server/internal/domain/contextx"
	"github.com/linzhengen/hub/server/internal/domain/system/role"
	"github.com/linzhengen/hub/server/internal/domain/system/role/rolepermission"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence/mysql"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence/mysql/sqlc"
)

func New(q *sqlc.Queries, rpRepo rolepermission.Repository) role.Repository {
	return &repositoryImpl{q: q, rpRepo: rpRepo}
}

type repositoryImpl struct {
	q      *sqlc.Queries
	rpRepo rolepermission.Repository
}

func (r repositoryImpl) FindOne(ctx context.Context, id string) (*role.Role, error) {
	rl, err := contextx.FindOne[sqlc.Role](ctx, id, mysql.GetQ(ctx, r.q).SelectRoleById, mysql.GetQ(ctx, r.q).SelectRoleForUpdate)
	if err != nil {
		return nil, err
	}
	return &role.Role{
		Id:          rl.ID,
		Name:        rl.Name,
		Description: rl.Description,
	}, nil
}

func (r repositoryImpl) Create(ctx context.Context, u *role.Role) error {
	return mysql.GetQ(ctx, r.q).CreateRole(ctx, sqlc.CreateRoleParams{
		ID:          u.Id,
		Name:        u.Name,
		Description: u.Description,
	})
}

func (r repositoryImpl) Update(ctx context.Context, u *role.Role) error {
	return mysql.GetQ(ctx, r.q).UpdateRole(ctx, sqlc.UpdateRoleParams{
		Name:        u.Name,
		Description: u.Description,
		ID:          u.Id,
	})
}

func (r repositoryImpl) Delete(ctx context.Context, id string) error {
	return mysql.GetQ(ctx, r.q).DeleteRole(ctx, id)
}

func (r repositoryImpl) AddPermissionsToRole(ctx context.Context, roleID string, permissionIDs []string) error {
	for _, permissionID := range permissionIDs {
		isIn, err := r.rpRepo.IsPermissionInRole(ctx, roleID, permissionID)
		if err != nil {
			return err
		}
		if isIn {
			continue
		}
		if err := r.rpRepo.AssignPermission(ctx, roleID, permissionID); err != nil {
			return err
		}
	}
	return nil
}

func (r repositoryImpl) RemovePermissionsFromRole(ctx context.Context, roleID string, permissionIDs []string) error {
	for _, permissionID := range permissionIDs {
		isIn, err := r.rpRepo.IsPermissionInRole(ctx, roleID, permissionID)
		if err != nil {
			return err
		}
		if !isIn {
			continue
		}
		if err := r.rpRepo.UnassignPermission(ctx, roleID, permissionID); err != nil {
			return err
		}
	}
	return nil
}
