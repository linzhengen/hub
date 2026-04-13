package group

import "context"

type Repository interface {
	FindOne(ctx context.Context, id string) (*Group, error)
	Create(ctx context.Context, g *Group) error
	Update(ctx context.Context, g *Group) error
	Delete(ctx context.Context, id string) error
}
