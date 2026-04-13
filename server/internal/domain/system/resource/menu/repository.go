package menu

import "context"

type Repository interface {
	FindAll(ctx context.Context) (Menus, error)
}
