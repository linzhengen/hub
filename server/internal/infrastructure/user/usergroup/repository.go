package usergroup

import (
	"context"

	"github.com/linzhengen/hub/v1/server/internal/domain/user/usergroup"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence"
)

func New(q persistence.Querier) usergroup.Repository {
	return &repositoryImpl{q: q}
}

type repositoryImpl struct {
	q persistence.Querier
}

func (r repositoryImpl) FindByUserId(ctx context.Context, userId string) (usergroup.UserGroups, error) {
	rows, err := persistence.GetQ(ctx, r.q).SelectUserGroupByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}
	var result usergroup.UserGroups
	for _, row := range rows {
		result = append(result, &usergroup.UserGroup{
			UserId:  row.UserID,
			GroupId: row.GroupID,
		})
	}
	return result, nil
}

func (r repositoryImpl) AssignGroup(ctx context.Context, userId, groupId string) error {
	return persistence.GetQ(ctx, r.q).CreateUserGroup(ctx, userId, groupId)
}

func (r repositoryImpl) UnassignGroup(ctx context.Context, userId, groupId string) error {
	return persistence.GetQ(ctx, r.q).DeleteUserGroup(ctx, userId, groupId)
}

func (r repositoryImpl) Upsert(ctx context.Context, userId string, groupId []string) error {
	q := persistence.GetQ(ctx, r.q)
	err := q.DeleteUserAllGroup(ctx, userId)
	if err != nil {
		return err
	}
	for _, id := range groupId {
		err = q.CreateUserGroup(ctx, userId, id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r repositoryImpl) AddUsersToGroup(ctx context.Context, groupID string, userIDs []string) error {
	q := persistence.GetQ(ctx, r.q)
	for _, userID := range userIDs {
		// 既にGroupに属している場合、AddをSkip
		isUserInGroup, err := r.IsUserInGroup(ctx, groupID, userID)
		if err != nil {
			return err
		}
		if isUserInGroup {
			continue
		}
		err = q.CreateUserGroup(ctx, userID, groupID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r repositoryImpl) RemoveUsersFromGroup(ctx context.Context, groupID string, userIDs []string) error {
	q := persistence.GetQ(ctx, r.q)
	for _, userID := range userIDs {
		// 既にGroupに属していない場合、RemoveをSkip
		isUserInGroup, err := r.IsUserInGroup(ctx, groupID, userID)
		if err != nil {
			return err
		}
		if !isUserInGroup {
			continue
		}
		err = q.DeleteUserGroup(ctx, userID, groupID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r repositoryImpl) IsUserInGroup(ctx context.Context, groupID string, userID string) (bool, error) {
	return persistence.GetQ(ctx, r.q).IsUserInGroup(ctx, userID, groupID)
}
