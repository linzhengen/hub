package group

import (
	"time"

	"github.com/linzhengen/hub/server/pkg/uuid"
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
	// OrgId is the organization this group belongs to.
	//
	// The group is where the tenant boundary lives, because the group is the
	// only edge that joins a user to a permission: every route through this
	// group is a route inside this organization. It is set when the group is
	// created and never changed - moving a group would carry its members'
	// access across a boundary in one write.
	OrgId     string
	CreatedAt time.Time
	UpdatedAt time.Time

	Roles []RoleGrant
}

// RoleGrant is one role the group holds, and when it stops holding it.
//
// It replaces a bare list of ids, for the same reason a membership did: a grant
// that ends on Friday and one that never ends are different facts.
type RoleGrant struct {
	RoleId string
	// ExpiresAt is when the grant lapses, nil when it does not.
	ExpiresAt *time.Time
}

func (g *Group) SetRoles(roles []RoleGrant) {
	g.Roles = roles
}

// RoleIds is the grants stripped back to ids, for the operations that take a
// set of roles rather than a set of grants.
func (g *Group) RoleIds() []string {
	ids := make([]string, 0, len(g.Roles))
	for _, r := range g.Roles {
		ids = append(ids, r.RoleId)
	}
	return ids
}

func Factory(
	Name string,
	Description string,
	OrgId string,
) *Group {
	return &Group{
		Id:          uuid.MustUUID().String(),
		Name:        Name,
		Description: Description,
		OrgId:       OrgId,
		Status:      Active,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}
