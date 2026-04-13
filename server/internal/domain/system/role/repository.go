package role

import "context"

type Repository interface {
	FindOne(ctx context.Context, id string) (*Role, error)
	Create(ctx context.Context, u *Role) error
	Update(ctx context.Context, u *Role) error
	Delete(ctx context.Context, id string) error
	AddPermissionsToRole(ctx context.Context, roleID string, permissionIDs []string) error
	RemovePermissionsFromRole(ctx context.Context, roleID string, permissionIDs []string) error
}
