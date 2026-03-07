package api

import "context"

type Repository interface {
	FindAll(ctx context.Context) (APIs, error)
}
