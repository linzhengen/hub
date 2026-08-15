package grouprole

import (
	"context"
	"time"
)

type Repository interface {
	FindByGroupId(ctx context.Context, groupId string) (GroupRoles, error)
	// AssignRole grants the role to the group until expiresAt, or for good when
	// expiresAt is nil.
	AssignRole(ctx context.Context, groupId, roleId string, expiresAt *time.Time) error
	UnassignRole(ctx context.Context, groupId, roleId string) error
	// Upsert replaces the group's roles. The grants it writes never expire: it
	// is the "these are the roles" operation, not a grant with a term.
	Upsert(ctx context.Context, groupId string, roleId []string) error
}
