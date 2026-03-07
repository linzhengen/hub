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

type User struct {
	Id        string
	Username  string
	Email     string
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time

	GroupIds []string
}

func (u *User) SetGroupIds(GroupIds []string) {
	u.GroupIds = GroupIds
}
