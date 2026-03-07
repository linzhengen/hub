package grouprole

import "context"

type Repository interface {
	FindByGroupId(ctx context.Context, groupId string) (GroupRoles, error)
	AssignRole(ctx context.Context, groupId, roleId string) error
	UnassignRole(ctx context.Context, groupId, roleId string) error
	Upsert(ctx context.Context, groupId string, roleId []string) error
}
