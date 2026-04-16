package usecase

import (
	"context"

	"github.com/linzhengen/hub/v1/server/internal/domain/system/resource/menu"
)

type UserFinder interface {
	GetMeMenus(ctx context.Context) ([]*menu.Menu, error)
}
