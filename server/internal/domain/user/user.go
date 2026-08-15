package user

import "time"

type Status string

const (
	Active   Status = "Active"
	InActive Status = "Inactive"
)

func (s Status) IsAllowedValue() bool {
	switch s {
	case Active, InActive:
		return true
	}
	return false
}

// GroupMembership is one group the user is in, and when they leave it.
//
// It replaces a bare list of ids. A membership that ends on Friday and one that
// never ends are different facts, and a list of ids could not tell them apart -
// which meant the console showed them the same way.
type GroupMembership struct {
	GroupId string
	// ExpiresAt is when the membership lapses, nil when it does not.
	ExpiresAt *time.Time
}

type User struct {
	Id        string
	Username  string
	Email     string
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time

	Groups []GroupMembership
}

func (u *User) SetGroups(groups []GroupMembership) {
	u.Groups = groups
}

// GroupIds is the memberships stripped back to ids, for the operations that
// take a set of groups rather than a set of grants - creating a user, and
// replacing the whole set.
func (u *User) GroupIds() []string {
	ids := make([]string, 0, len(u.Groups))
	for _, g := range u.Groups {
		ids = append(ids, g.GroupId)
	}
	return ids
}

// PermanentMemberships is what a set of ids means when nobody said otherwise:
// membership that does not end.
func PermanentMemberships(groupIds []string) []GroupMembership {
	memberships := make([]GroupMembership, 0, len(groupIds))
	for _, id := range groupIds {
		memberships = append(memberships, GroupMembership{GroupId: id})
	}
	return memberships
}
