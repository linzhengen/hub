package resource

import "context"

type Repository interface {
	FindOne(ctx context.Context, id string) (*Resource, error)
	FindOneByIdentifier(ctx context.Context, identifier string) (*Resource, error)
	Create(ctx context.Context, u *Resource) error
	Update(ctx context.Context, u *Resource) error
	Delete(ctx context.Context, id string) error
}
