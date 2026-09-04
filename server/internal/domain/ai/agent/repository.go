package agent

import "context"

// ListParams is the page to read, and what to narrow it to.
type ListParams struct {
	// OrgId narrows to one tenant, empty meaning all the caller can see.
	OrgId string
	// ParentAgentId narrows to one agent's children, empty meaning no filter.
	ParentAgentId string
	Limit         uint32
	Offset        uint32
	// OrgIds keeps the listing to the organizations the caller holds a live
	// grant in. nil is every organization, which is what a platform grant
	// means; it is not the same as an empty slice, which admits nothing.
	OrgIds []string
}

type Repository interface {
	Create(ctx context.Context, a *Agent) error
	FindOne(ctx context.Context, id string) (*Agent, error)
	List(ctx context.Context, params ListParams) ([]*Agent, int64, error)
	// CountChildren is how many agents name this one as their parent. Deleting
	// an agent that still has children would leave their credentials working
	// with nothing recording them, so the count is asked before anything is
	// removed rather than left to a foreign key to refuse mid-transaction.
	CountChildren(ctx context.Context, id string) (int64, error)
	// RecordSecretRotation stamps the moment a new credential was issued. The
	// secret itself is never stored, so this is all hub can report about it.
	RecordSecretRotation(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}
