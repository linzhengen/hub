package organization

import "context"

// ListParams is the page to read, and what to narrow it to.
type ListParams struct {
	// Name and Slug are partial matches, empty meaning no filter.
	Name string
	Slug string
	// Kind narrows to one sort of tenant, empty meaning all of them.
	Kind   Kind
	Limit  uint32
	Offset uint32
	// Ids narrows to a known set, which is how a caller is kept to the
	// organizations they can reach. nil is every organization.
	Ids []string
}

type Repository interface {
	Create(ctx context.Context, o *Organization) error
	FindOne(ctx context.Context, id string) (*Organization, error)
	FindBySlug(ctx context.Context, slug string) (*Organization, error)
	Update(ctx context.Context, o *Organization) error
	Delete(ctx context.Context, id string) error
	// FindByUser is where a user is, answered from the graph rather than from a
	// membership table of its own: belonging to an organization is having a
	// group in it, which is the edge an authorization decision already reads.
	// A second table would be a second answer to the same question.
	FindByUser(ctx context.Context, userId string) ([]*Organization, error)
}
