package role

import (
	"time"

	"github.com/linzhengen/hub/server/pkg/uuid"
)

type Role struct {
	Id          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time

	PermissionIds []string
}

func (r *Role) SetPermissionIds(permissionIds []string) {
	r.PermissionIds = permissionIds
}

func Factory(
	name string,
	description string,
) *Role {
	return &Role{
		Id:          uuid.MustUUID().String(),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}
