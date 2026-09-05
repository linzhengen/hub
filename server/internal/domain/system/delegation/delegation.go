// Package delegation is the record of which agent may act on whose behalf.
//
// A delegation is a row, not a token. That is the whole design: a row can be
// revoked now - the revision counter moves on the write and every policy cache
// drops what it holds within a second - whereas a token has a residual lifetime
// that can only be shortened by making it short-lived or by consulting a
// revocation list, which is a table again.
//
// Authentication is unchanged by this package. An agent authenticates as itself
// through the client credentials grant; what a delegation adds is the right to
// *say* it is acting for someone. An agent cannot claim an arbitrary person
// because the row does not exist.
package delegation

import (
	"time"

	"github.com/linzhengen/hub/server/pkg/uuid"
)

// Delegation is one agent's permission to act for one person.
//
// The effective authority of an agent acting under it is
//
//	the agent's own grants ∩ the principal's grants ∩ this delegation
//
// and the intersection is taken when the decision is made, never when the row
// is written: folding it in at write time would keep the agent working after
// the principal lost the access it was derived from.
type Delegation struct {
	Id string
	// AgentId is the agent that may act.
	AgentId string
	// PrincipalUserId is whose authority is lent. Only this person may create
	// the row through the API, which is what stops an administrator from
	// handing an agent a stranger's authority.
	PrincipalUserId string
	// GrantedByUserId is who wrote the row. Today always the principal; it is
	// separate because the other route in - an approved access request - has
	// somebody else settle it.
	GrantedByUserId string
	// OrgId is the organization the delegation is answerable in, copied from
	// the agent.
	OrgId string
	// Reason is why it exists, in the words of the person who agreed to it. A
	// delegation nobody can explain is one nobody dares revoke.
	Reason string
	// PermissionIds is what the delegation carries: a subset of what the
	// principal holds, named in the same vocabulary a decision is made in.
	//
	// It is never empty. There is no "everything" delegation - see Valid.
	PermissionIds []string
	// MaxDepth is how many agents may stand in the chain. 1 is the agent named
	// here and no further. Nothing reads it yet; the chain arrives with the
	// authorization change.
	MaxDepth uint32
	// ExpiresAt is when the delegation stops, nil for one that does not. The
	// meaning is user_groups.expires_at's exactly, and like it, it is dropped
	// by the decision against the clock rather than filtered in SQL.
	ExpiresAt *time.Time
	// RevokedAt is when it was revoked, nil while live. Revocation is a write,
	// so unlike an expiry it moves the revision counter.
	RevokedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Live reports whether the delegation may be relied on at a moment.
//
// Both halves are here rather than in the caller so that "is this delegation
// usable" has one answer. Revocation and expiry end it for different reasons -
// one is a write, the other is the clock - but a caller never has to care
// which.
func (d *Delegation) Live(now time.Time) bool {
	if d.RevokedAt != nil && !now.Before(*d.RevokedAt) {
		return false
	}
	return d.ExpiresAt == nil || now.Before(*d.ExpiresAt)
}

// Revoked reports whether the delegation has been withdrawn, whatever the
// clock says.
func (d *Delegation) Revoked() bool {
	return d.RevokedAt != nil
}

// MaxChainDepth is the deepest chain a delegation may authorise.
//
// A bound exists because each step adds an agent that has to be trusted, and a
// chain nobody can hold in their head is one nobody can review. The number is
// arbitrary; that there is one is not.
const MaxChainDepth uint32 = 8

// Factory builds a delegation. The identifiers are settled by the caller
// because every one of them is checked against something before it gets here:
// the agent must exist, the principal must be the caller, and the permissions
// must be ones the principal actually holds.
func Factory(
	agentId, principalUserId, grantedByUserId, orgId, reason string,
	permissionIds []string,
	maxDepth uint32,
	expiresAt *time.Time,
) *Delegation {
	now := time.Now()
	return &Delegation{
		Id:              uuid.MustUUID().String(),
		AgentId:         agentId,
		PrincipalUserId: principalUserId,
		GrantedByUserId: grantedByUserId,
		OrgId:           orgId,
		Reason:          reason,
		PermissionIds:   permissionIds,
		MaxDepth:        maxDepth,
		ExpiresAt:       expiresAt,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}
