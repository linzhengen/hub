package role

import (
	"time"

	"github.com/linzhengen/hub/server/pkg/uuid"
)

type Role struct {
	Id          string
	Name        string
	Description string
	// OrgId is the organization that defines this role, empty for one this
	// installation provides to every organization.
	//
	// Unlike a group's, this may be empty: a group is where the tenant boundary
	// lives, so it always has an organization, whereas a role is a named bundle
	// of permissions that can sensibly be shared.
	OrgId     string
	CreatedAt time.Time
	UpdatedAt time.Time

	PermissionIds []string
}

// Shared reports whether every organization may hold this role.
func (r *Role) Shared() bool {
	return r.OrgId == ""
}

func (r *Role) SetPermissionIds(permissionIds []string) {
	r.PermissionIds = permissionIds
}

func Factory(
	name string,
	description string,
	orgId string,
) *Role {
	return &Role{
		Id:          uuid.MustUUID().String(),
		Name:        name,
		Description: description,
		OrgId:       orgId,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}
