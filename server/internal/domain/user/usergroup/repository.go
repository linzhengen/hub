package usergroup

import "context"

type Repository interface {
	FindByUserId(ctx context.Context, userId string) (UserGroups, error)
	AssignGroup(ctx context.Context, userId, groupId string) error
	UnassignGroup(ctx context.Context, userId, groupId string) error
	Upsert(ctx context.Context, userId string, groupId []string) error
	AddUsersToGroup(ctx context.Context, groupID string, userIDs []string) error
	RemoveUsersFromGroup(ctx context.Context, groupID string, userIDs []string) error
	IsUserInGroup(ctx context.Context, groupID string, userID string) (bool, error)
}
