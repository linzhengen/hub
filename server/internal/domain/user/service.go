package user

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"github.com/linzhengen/hub/server/internal/domain/trans"
	"github.com/linzhengen/hub/server/internal/domain/user/usergroup"
)

type Service interface {
	CreateIfNotExists(ctx context.Context, u *User) error
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

func (s service) CreateIfNotExists(ctx context.Context, u *User) error {
	_, err := s.repo.FindOne(ctx, u.Id)
	if err == nil {
		// ユーザーが存在するので、何もしないで正常終了
		return nil
	}
	// エラーが発生した場合
	if !errors.Is(err, sql.ErrNoRows) {
		// それが ErrNoRows 以外のエラーなら、そのエラーを返す
		return err
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
		return err
	}

	return nil
}
