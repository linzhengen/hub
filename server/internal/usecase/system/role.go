package system

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/doug-martin/goqu/v9"

	"github.com/linzhengen/hub/v1/server/internal/domain/system/role"
	"github.com/linzhengen/hub/v1/server/internal/domain/system/role/rolepermission"
	"github.com/linzhengen/hub/v1/server/internal/domain/trans"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence/postgres"

	"github.com/linzhengen/hub/v1/server/pkg/logger"
)

type RoleUseCase interface {
	Get(ctx context.Context, roleId string) (*role.Role, error)
	Create(ctx context.Context, r *role.Role) (*role.Role, error)
	Update(ctx context.Context, r *role.Role) (*role.Role, error)
	Delete(ctx context.Context, roleId string) error
	List(ctx context.Context, params *ListRoleQueryParams) ([]*role.Role, int64, error)
	AssignPermission(ctx context.Context, roleId, permissionId string) (*role.Role, error)
	AddPermissionsToRole(ctx context.Context, roleId string, permissionIds []string) (*role.Role, error)
	RemovePermissionsFromRole(ctx context.Context, roleId string, permissionIds []string) (*role.Role, error)
}

func NewRoleUseCase(
	db *sql.DB,
	transRepo trans.Repository,
	roleRepo role.Repository,
	rolePermissionRepo rolepermission.Repository,
	dialectWrapper persistence.DialectWrapper,
) RoleUseCase {
	return &roleUseCase{
		db:                 db,
		transRepo:          transRepo,
		roleRepo:           roleRepo,
		rolePermissionRepo: rolePermissionRepo,
		dialectWrapper:     dialectWrapper,
	}
}

type roleUseCase struct {
	db                 *sql.DB
	transRepo          trans.Repository
	roleRepo           role.Repository
	rolePermissionRepo rolepermission.Repository
	dialectWrapper     persistence.DialectWrapper
}

type ListRoleQueryParams struct {
	Limit         uint32
	Offset        uint32
	RoleIds       []string
	RoleName      string
	PermissionIds []string
}

func (uc roleUseCase) Get(ctx context.Context, roleId string) (*role.Role, error) {
	return uc.roleRepo.FindOne(ctx, roleId)
}

func (uc roleUseCase) Create(ctx context.Context, r *role.Role) (*role.Role, error) {
	if err := uc.roleRepo.Create(ctx, r); err != nil {
		return nil, err
	}
	return uc.roleRepo.FindOne(ctx, r.Id)
}

func (uc roleUseCase) Update(ctx context.Context, r *role.Role) (*role.Role, error) {
	if err := uc.transRepo.ExecTransWithLock(ctx, func(ctx context.Context) error {
		_, err := uc.roleRepo.FindOne(ctx, r.Id)
		if err != nil {
			return err
		}
		return uc.roleRepo.Update(ctx, r)
	}); err != nil {
		return nil, err
	}
	return uc.roleRepo.FindOne(ctx, r.Id)
}

func (uc roleUseCase) Delete(ctx context.Context, roleId string) error {
	return uc.transRepo.ExecTransWithLock(ctx, func(ctx context.Context) error {
		_, err := uc.roleRepo.FindOne(ctx, roleId)
		if err != nil {
			return err
		}
		return uc.roleRepo.Delete(ctx, roleId)
	})
}

func (uc roleUseCase) List(ctx context.Context, params *ListRoleQueryParams) ([]*role.Role, int64, error) {
	b := uc.dialectWrapper.From("roles")
	if params.RoleIds != nil {
		b = b.Where(goqu.Ex{"id": params.RoleIds})
	}
	if params.RoleName != "" {
		b = b.Where(goqu.C("name").Like(fmt.Sprintf("%%%s%%", params.RoleName)))
	}
	if params.PermissionIds != nil {
		// Create a subquery to check if the role belongs to any of the specified permissions
		subquery := uc.dialectWrapper.From("role_permissions").
			Select(goqu.L("1")).
			Where(goqu.Ex{
				"role_permissions.role_id":       goqu.I("roles.id"),
				"role_permissions.permission_id": params.PermissionIds,
			})

		// Use EXISTS with the subquery
		b = b.Where(goqu.L("EXISTS ?", subquery))
	}
	cnt, err := postgres.SelectCount(ctx, uc.db, b)
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination only when limit > 0
	if params.Limit > 0 {
		b = b.Limit(uint(params.Limit)).Offset(uint(params.Offset))
	}

	items, err := uc.list(ctx, b)
	if err != nil {
		return nil, 0, err
	}
	if len(items) == 0 {
		return items, cnt, nil
	}

	// Collect all role IDs
	var roleIds []string
	for _, item := range items {
		roleIds = append(roleIds, item.Id)
	}

	// Fetch all role-permission relationships in a single query
	rolePermissionMap := make(map[string][]string)

	// Build a query to get all role-permission relationships for the role IDs
	rpQuery := uc.dialectWrapper.From("role_permissions").
		Select("role_id", "permission_id").
		Where(goqu.Ex{"role_id": roleIds})

	rpSQL, rpParams, err := rpQuery.Prepared(true).ToSQL()
	if err != nil {
		return nil, 0, err
	}

	// Execute the query
	rpRows, err := uc.db.QueryContext(ctx, rpSQL, rpParams...)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		err := rpRows.Close()
		if err != nil {
			logger.Errorf("List: error closing role-permission rows: %v", err)
		}
	}()

	// Process the results
	for rpRows.Next() {
		var roleId, permissionId string
		if err := rpRows.Scan(&roleId, &permissionId); err != nil {
			return nil, 0, err
		}
		// Fix: Add the permission ID to the map
		rolePermissionMap[roleId] = append(rolePermissionMap[roleId], permissionId)
	}

	if err := rpRows.Err(); err != nil {
		return nil, 0, err
	}

	// Set permission IDs for each role
	for _, item := range items {
		if permissionIds, ok := rolePermissionMap[item.Id]; ok {
			item.SetPermissionIds(permissionIds)
		}
	}
	return items, cnt, nil
}

func (uc roleUseCase) list(ctx context.Context, b *goqu.SelectDataset) ([]*role.Role, error) {
	b = b.Select("*")
	query, queryParams, err := b.Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}
	rows, err := uc.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			logger.Infof("error closing rows: %v", err)
		}
	}()
	var items []*role.Role
	for rows.Next() {
		var i role.Role
		if err := rows.Scan(
			&i.Id,
			&i.Name,
			&i.Description,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (uc roleUseCase) AssignPermission(ctx context.Context, roleId, permissionId string) (*role.Role, error) {
	if err := uc.rolePermissionRepo.AssignPermission(ctx, roleId, permissionId); err != nil {
		return nil, err
	}
	return uc.Get(ctx, roleId)
}

func (uc roleUseCase) AddPermissionsToRole(ctx context.Context, roleId string, permissionIds []string) (*role.Role, error) {
	if err := uc.transRepo.ExecTrans(ctx, func(ctx context.Context) error {
		return uc.roleRepo.AddPermissionsToRole(ctx, roleId, permissionIds)
	}); err != nil {
		return nil, err
	}
	return uc.Get(ctx, roleId)
}

func (uc roleUseCase) RemovePermissionsFromRole(ctx context.Context, roleId string, permissionIds []string) (*role.Role, error) {
	if err := uc.transRepo.ExecTrans(ctx, func(ctx context.Context) error {
		return uc.roleRepo.RemovePermissionsFromRole(ctx, roleId, permissionIds)
	}); err != nil {
		return nil, err
	}
	return uc.Get(ctx, roleId)
}
