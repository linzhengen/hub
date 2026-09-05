package delegation

import "context"

// ListParams is the page to read, and what to narrow it to.
type ListParams struct {
	// AgentId narrows to one agent's delegations, empty meaning no filter.
	AgentId string
	// PrincipalUserId narrows to one person's, empty meaning no filter.
	PrincipalUserId string
	// IncludeRevoked keeps withdrawn delegations in the answer. They are left
	// out by default: the usual question is "what may act for me now", and a
	// revoked row answers a different one.
	IncludeRevoked bool
	Limit          uint32
	Offset         uint32
	// OrgIds keeps the listing to the organizations the caller holds a live
	// grant in. nil is every organization, which is what a platform grant
	// means; it is not the same as an empty slice, which admits nothing.
	OrgIds []string
	// SelfUserId is always shown regardless of OrgIds: a person can see the
	// delegations they themselves granted, wherever the agent lives. Without
	// it, revoking your own delegation could require access you do not have.
	SelfUserId string
}

type Repository interface {
	// Create writes the delegation and the permissions it carries together.
	Create(ctx context.Context, d *Delegation) error
	FindOne(ctx context.Context, id string) (*Delegation, error)
	List(ctx context.Context, params ListParams) ([]*Delegation, int64, error)
	// Revoke stamps the row as withdrawn and reports whether it was still live.
	// A second revocation is not an error to the caller, but it is not a second
	// event either.
	Revoke(ctx context.Context, id string) (bool, error)
}
