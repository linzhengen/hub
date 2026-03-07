package permission

import "context"

type Repository interface {
	FindOne(ctx context.Context, id string) (*Permission, error)
	FindByResourceId(ctx context.Context, resourceId string) ([]*Permission, error)
	Create(ctx context.Context, p *Permission) error
	Update(ctx context.Context, p *Permission) error
	Delete(ctx context.Context, id string) error
}
