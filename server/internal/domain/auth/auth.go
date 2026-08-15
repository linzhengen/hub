package auth

import "time"

// Policy represents a Casbin policy rule
type Policy struct {
	Subject string // User or role
	Object  string // Resource
	Action  string // Action (e.g., read, write, delete)
	// ExpiresAt is when the grant behind this policy lapses, nil for one that
	// does not. It is the earliest expiry on the route, so a route is only live
	// while every edge of it is.
	//
	// The repository returns lapsed policies rather than filtering them out, and
	// Enforce drops them as it decides. Expiry is the one change to the graph
	// that nobody writes, so no trigger fires for it and rbac_revisions does not
	// move: a policy cache filtered in SQL would keep serving a grant that had
	// already lapsed. Deciding against the clock rather than against whenever
	// the cache was filled is what makes the expiry mean anything.
	ExpiresAt *time.Time
}

// Request represents an authorization request
type Request struct {
	Subject string // User ID
	Object  string // Resource identifier
	Action  string // Action to perform
}

// AccessPath is one route by which a user holds a permission: the group they
// belong to, the role that group holds, and the permission that role grants.
//
// A Policy says what somebody may do. An AccessPath says how they came to be
// able to do it, which is what has to be edited to stop them.
//
// Object and Action are the pattern the permission itself carries, not the
// request it was matched against: a path granted through `api.*` reports
// `api.*`. Resolving that away would hide the grant most worth finding.
type AccessPath struct {
	GroupId      string
	GroupName    string
	RoleId       string
	RoleName     string
	PermissionId string
	Object       string
	Action       string
	// ExpiresAt is the earliest expiry on the route, nil when no edge of it
	// expires.
	ExpiresAt *time.Time
}

// Principal is a user allowed some operation, with every route that allows it.
type Principal struct {
	UserId   string
	Username string
	Paths    []AccessPath
}

// Earliest returns whichever of two expiries comes first, and nil only when
// neither expires.
//
// A route is live while every edge of it is, so an edge that lapses tomorrow
// ends a route whose other edge lasts a year. nil means "never", which is why
// it loses to any actual time rather than winning as a zero value would.
func Earliest(a, b *time.Time) *time.Time {
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	case b.Before(*a):
		return b
	default:
		return a
	}
}

// Membership is one user's place in one group. It carries no permission of its
// own; it is the edge that joins a user to the paths their groups hold.
type Membership struct {
	UserId   string
	Username string
	GroupId  string
	// ExpiresAt is when this user leaves this group, nil when they do not. It
	// is only the membership edge: the route it completes may lapse earlier
	// because the group's hold on the role does.
	ExpiresAt *time.Time
}
