package rolepermission

import "context"

type Repository interface {
	FindByRoleId(ctx context.Context, roleId string) (RolePermissions, error)
	AssignPermission(ctx context.Context, roleId, permissionId string) error
	UnassignPermission(ctx context.Context, roleId, permissionId string) error
	Upsert(ctx context.Context, roleId string, permissionId []string) error
	IsPermissionInRole(ctx context.Context, roleId string, permissionId string) (bool, error)
}
