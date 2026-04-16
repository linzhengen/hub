package group

import (
	"time"

	"github.com/linzhengen/hub/v1/server/pkg/uuid"
)

const (
	AdminGroupId = "00000000-0000-0000-0000-000000000001"
)

type Status string

const (
	Active   Status = "Active"
	InActive Status = "Inactive"
)

type Group struct {
	Id          string
	Name        string
	Description string
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time

	RoleIds []string
}

func (g *Group) SetRoleIds(roleIds []string) {
	g.RoleIds = roleIds
}

func Factory(
	Name string,
	Description string,
) *Group {
	return &Group{
		Id:          uuid.MustUUID().String(),
		Name:        Name,
		Description: Description,
		Status:      Active,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}
