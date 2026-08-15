package usergroup

import "time"

type UserGroup struct {
	UserId  string
	GroupId string
	// ExpiresAt is when the user leaves the group, nil when they do not.
	ExpiresAt *time.Time
}

type UserGroups []*UserGroup

func (u UserGroups) GroupIds(userId string) []string {
	var groupIds []string
	for _, group := range u {
		if group.UserId == userId {
			groupIds = append(groupIds, group.GroupId)
		}
	}
	return groupIds
}
