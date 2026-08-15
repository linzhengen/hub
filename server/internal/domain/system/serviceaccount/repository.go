package serviceaccount

import "context"

// ListParams is the page to read. There is no filter: a deployment has a
// handful of machines, and who created one is a column to read rather than a
// question to ask the database.
type ListParams struct {
	Limit  uint32
	Offset uint32
}

type Repository interface {
	Create(ctx context.Context, s *ServiceAccount) error
	FindOne(ctx context.Context, id string) (*ServiceAccount, error)
	List(ctx context.Context, params ListParams) ([]*ServiceAccount, int64, error)
	Delete(ctx context.Context, id string) error
}
