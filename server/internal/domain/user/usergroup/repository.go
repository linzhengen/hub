package usergroup

import (
	"context"
	"time"
)

type Repository interface {
	FindByUserId(ctx context.Context, userId string) (UserGroups, error)
	// AssignGroup puts the user in the group until expiresAt, or for good when
	// expiresAt is nil.
	AssignGroup(ctx context.Context, userId, groupId string, expiresAt *time.Time) error
	UnassignGroup(ctx context.Context, userId, groupId string) error
	// Upsert replaces the user's groups. The memberships it writes never
	// expire: it is the "these are the groups" operation, not a grant with a
	// term.
	Upsert(ctx context.Context, userId string, groupId []string) error
	AddUsersToGroup(ctx context.Context, groupID string, userIDs []string, expiresAt *time.Time) error
	RemoveUsersFromGroup(ctx context.Context, groupID string, userIDs []string) error
	IsUserInGroup(ctx context.Context, groupID string, userID string) (bool, error)
}
