package usergroup

import (
	"context"

	"github.com/linzhengen/hub/server/internal/domain/user/usergroup"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence/mysql"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence/mysql/sqlc"
)

func New(q *sqlc.Queries) usergroup.Repository {
	return &repositoryImpl{q: q}
}

type repositoryImpl struct {
	q *sqlc.Queries
}

func (r repositoryImpl) FindByUserId(ctx context.Context, userId string) (usergroup.UserGroups, error) {
	rows, err := mysql.GetQ(ctx, r.q).SelectUserGroupByUserId(ctx, userId)
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
	return mysql.GetQ(ctx, r.q).CreateUserGroup(ctx, sqlc.CreateUserGroupParams{
		UserID:  userId,
		GroupID: groupId,
	})
}

func (r repositoryImpl) UnassignGroup(ctx context.Context, userId, groupId string) error {
	return mysql.GetQ(ctx, r.q).DeleteUserGroup(ctx, sqlc.DeleteUserGroupParams{
		UserID:  userId,
		GroupID: groupId,
	})
}

func (r repositoryImpl) Upsert(ctx context.Context, userId string, groupId []string) error {
	err := mysql.GetQ(ctx, r.q).DeleteUserAllGroup(ctx, userId)
	if err != nil {
		return err
	}
	for _, id := range groupId {
		err = mysql.GetQ(ctx, r.q).CreateUserGroup(ctx, sqlc.CreateUserGroupParams{
			UserID:  userId,
			GroupID: id,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r repositoryImpl) AddUsersToGroup(ctx context.Context, groupID string, userIDs []string) error {
	for _, userID := range userIDs {
		// 既にGroupに属している場合、AddをSkip
		isUserInGroup, err := r.IsUserInGroup(ctx, groupID, userID)
		if err != nil {
			return err
		}
		if isUserInGroup {
			continue
		}
		err = mysql.GetQ(ctx, r.q).CreateUserGroup(ctx, sqlc.CreateUserGroupParams{
			UserID:  userID,
			GroupID: groupID,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r repositoryImpl) RemoveUsersFromGroup(ctx context.Context, groupID string, userIDs []string) error {
	for _, userID := range userIDs {
		// 既にGroupに属していない場合、RemoveをSkip
		isUserInGroup, err := r.IsUserInGroup(ctx, groupID, userID)
		if err != nil {
			return err
		}
		if !isUserInGroup {
			continue
		}
		err = mysql.GetQ(ctx, r.q).DeleteUserGroup(ctx, sqlc.DeleteUserGroupParams{
			UserID:  userID,
			GroupID: groupID,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r repositoryImpl) IsUserInGroup(ctx context.Context, groupID string, userID string) (bool, error) {
	return mysql.GetQ(ctx, r.q).IsUserInGroup(ctx, sqlc.IsUserInGroupParams{
		UserID:  userID,
		GroupID: groupID,
	})
}
