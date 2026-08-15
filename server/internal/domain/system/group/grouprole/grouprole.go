package grouprole

import "time"

type GroupRole struct {
	GroupId string
	RoleId  string
	// ExpiresAt is when the group loses the role, nil when it does not.
	ExpiresAt *time.Time
}

type GroupRoles []*GroupRole
