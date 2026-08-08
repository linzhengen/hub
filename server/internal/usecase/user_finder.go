package usecase

import (
	"context"

	"github.com/linzhengen/hub/server/internal/domain/system/resource/menu"
)

type UserFinder interface {
	GetMeMenus(ctx context.Context) ([]*menu.Menu, error)
}
