package user

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"github.com/linzhengen/hub/v1/server/internal/domain/trans"
	"github.com/linzhengen/hub/v1/server/internal/domain/user/usergroup"
)

type Service interface {
	// CreateIfNotExists provisions a user on first sight and returns the
	// stored user either way.
	//
	// It returns the user rather than just an error because it has to read the
	// row to decide whether to create it, and every caller needs that row
	// straight afterwards. Discarding it made the authentication interceptor's
	// lookup invisible to the authorization interceptor, which then read the
	// same row again on every single request.
	CreateIfNotExists(ctx context.Context, u *User) (*User, error)
}
type service struct {
	trans         trans.Repository
	repo          Repository
	userGroupRepo usergroup.Repository
}

func NewService(t trans.Repository, r Repository, userGroupRepo usergroup.Repository) Service {

	return &service{
		trans:         t,
		repo:          r,
		userGroupRepo: userGroupRepo,
	}
}

func (s service) CreateIfNotExists(ctx context.Context, u *User) (*User, error) {
	existing, err := s.repo.FindOne(ctx, u.Id)
	if err == nil {
		// ユーザーが存在するので、保存されている姿をそのまま返す。
		// 引数の u は呼び出し側が組み立てたものなので、ステータスなど
		// DB 側でしか分からない情報は持っていない。
		return existing, nil
	}
	// エラーが発生した場合
	if !errors.Is(err, sql.ErrNoRows) {
		// それが ErrNoRows 以外のエラーなら、そのエラーを返す
		return nil, err
	}

	// 以下は ErrNoRows だった場合の処理 (既存のロジック)
	if err := s.trans.ExecTrans(ctx, func(ctx context.Context) error {
		if err := s.repo.Create(ctx, u); err != nil {
			return err
		}
		if len(u.GroupIds) == 0 {
			u.GroupIds = append(u.GroupIds, uuid.Nil.String())
		}
		return s.userGroupRepo.Upsert(ctx, u.Id, u.GroupIds)
	}); err != nil {
		return nil, err
	}

	return u, nil
}
