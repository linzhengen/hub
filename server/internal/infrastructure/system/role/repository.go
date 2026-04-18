package role

import (
	"context"

	"github.com/linzhengen/hub/v1/server/internal/domain/contextx"
	"github.com/linzhengen/hub/v1/server/internal/domain/system/role"
	"github.com/linzhengen/hub/v1/server/internal/domain/system/role/rolepermission"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence"
)

func New(q persistence.Querier, rpRepo rolepermission.Repository) role.Repository {
	return &repositoryImpl{q: q, rpRepo: rpRepo}
}

type repositoryImpl struct {
	q      persistence.Querier
	rpRepo rolepermission.Repository
}

func (r repositoryImpl) FindOne(ctx context.Context, id string) (*role.Role, error) {
	q := persistence.GetQ(ctx, r.q)
	var rl *persistence.RoleModel
	var err error
	if contextx.FromTransLock(ctx) {
		rl, err = q.SelectRoleForUpdate(ctx, id)
	} else {
		rl, err = q.SelectRoleById(ctx, id)
	}
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
	return persistence.GetQ(ctx, r.q).CreateRole(ctx, u.Id, u.Name, u.Description)
}

func (r repositoryImpl) Update(ctx context.Context, u *role.Role) error {
	return persistence.GetQ(ctx, r.q).UpdateRole(ctx, u.Id, u.Name, u.Description)
}

func (r repositoryImpl) Delete(ctx context.Context, id string) error {
	return persistence.GetQ(ctx, r.q).DeleteRole(ctx, id)
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
