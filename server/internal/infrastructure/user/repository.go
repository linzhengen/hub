package user

import (
	"context"

	"github.com/linzhengen/hub/v1/server/internal/domain/contextx"
	"github.com/linzhengen/hub/v1/server/internal/domain/user"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence"
)

type repositoryImpl struct {
	q persistence.Querier
}

func New(q persistence.Querier) user.Repository {
	return &repositoryImpl{q: q}
}

func (r repositoryImpl) FindOne(ctx context.Context, id string) (*user.User, error) {
	q := persistence.GetQ(ctx, r.q)
	var userModel *persistence.UserModel
	var err error

	if contextx.FromTransLock(ctx) {
		userModel, err = q.SelectUserForUpdate(ctx, id)
	} else {
		userModel, err = q.SelectUserById(ctx, id)
	}
	if err != nil {
		return nil, err
	}
	return convertUserModel(userModel), nil
}

func (r repositoryImpl) Create(ctx context.Context, u *user.User) error {
	return persistence.GetQ(ctx, r.q).CreateUser(ctx, u.Id, u.Username, u.Email, string(u.Status))
}

func (r repositoryImpl) Update(ctx context.Context, u *user.User) error {
	return persistence.GetQ(ctx, r.q).UpdateUser(ctx, u.Id, u.Username, u.Email, string(u.Status))
}

func (r repositoryImpl) Delete(ctx context.Context, id string) error {
	return persistence.GetQ(ctx, r.q).DeleteUser(ctx, id)
}

// convertUserModel converts persistence.UserModel to domain user.User
func convertUserModel(userModel *persistence.UserModel) *user.User {
	return &user.User{
		Id:        userModel.ID,
		Username:  userModel.Username,
		Email:     userModel.Email,
		Status:    user.Status(userModel.Status),
		CreatedAt: userModel.CreatedAt,
		UpdatedAt: userModel.UpdatedAt,
	}
}
