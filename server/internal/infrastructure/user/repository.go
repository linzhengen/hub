package user

import (
	"context"

	"github.com/linzhengen/hub/v1/server/internal/domain/contextx"
	"github.com/linzhengen/hub/v1/server/internal/domain/user"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence/mysql"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence/mysql/sqlc"
)

type repositoryImpl struct {
	q *sqlc.Queries
}

func New(q *sqlc.Queries) user.Repository {
	return &repositoryImpl{q: q}
}

func (r repositoryImpl) FindOne(ctx context.Context, id string) (*user.User, error) {
	u, err := contextx.FindOne[sqlc.User](ctx, id, mysql.GetQ(ctx, r.q).SelectUserById, mysql.GetQ(ctx, r.q).SelectUserForUpdate)
	if err != nil {
		return nil, err
	}
	return &user.User{
		Id:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Status:    user.Status(u.Status),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}, nil
}

func (r repositoryImpl) Create(ctx context.Context, u *user.User) error {
	return mysql.GetQ(ctx, r.q).CreateUser(ctx, sqlc.CreateUserParams{
		ID:       u.Id,
		Username: u.Username,
		Email:    u.Email,
		Status:   string(u.Status),
	})
}

func (r repositoryImpl) Update(ctx context.Context, u *user.User) error {
	return mysql.GetQ(ctx, r.q).UpdateUser(ctx, sqlc.UpdateUserParams{
		Username: u.Username,
		Email:    u.Email,
		Status:   string(u.Status),
		ID:       u.Id,
	})
}

func (r repositoryImpl) Delete(ctx context.Context, id string) error {
	return mysql.GetQ(ctx, r.q).DeleteUser(ctx, id)
}
