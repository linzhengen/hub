package usergroup

type UserGroup struct {
	UserId  string
	GroupId string
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
