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
	// OrgId is the organization of the group this policy was reached through.
	// A policy is answerable about that organization and no other, unless
	// OrgWide says otherwise.
	OrgId string
	// OrgWide is true for a policy held through the platform organization,
	// which answers about every organization.
	//
	// It is carried as a decided boolean rather than as the organization's kind
	// so that the rule - stated once, in organization.Kind.AppliesEverywhere -
	// is applied where the graph is read rather than re-derived at each
	// decision. A decision only ever asks "does this reach here?".
	OrgWide bool
}

// grant is the part of a Policy or an AccessPath that a decision reads.
//
// Enforce works from policies and Explain from access paths, and they must
// agree; giving both one shape to be reduced to is what makes it possible for
// them to share the single `allows`.
type grant struct {
	Object    string
	Action    string
	ExpiresAt *time.Time
	OrgId     string
	OrgWide   bool
}

func (p Policy) grant() grant {
	return grant{Object: p.Object, Action: p.Action, ExpiresAt: p.ExpiresAt, OrgId: p.OrgId, OrgWide: p.OrgWide}
}

// reaches reports whether this grant is answerable about orgId.
//
// An empty orgId is a request that names no organization, and every grant
// reaches it. That is what a request looked like before organizations existed
// and it still means the same thing: "may they, anywhere they hold access?".
// Narrowing only happens when a caller actually names a place.
func (g grant) reaches(orgId string) bool {
	return g.OrgWide || orgId == "" || g.OrgId == orgId
}

// Request represents an authorization request
type Request struct {
	Subject string // User ID
	Object  string // Resource identifier
	Action  string // Action to perform
	// OrgId narrows the question to one organization. Empty asks it of every
	// organization the subject holds access in, which is what every caller
	// asked before organizations existed.
	OrgId string
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
	// OrgId and OrgName are the organization the group belongs to, and OrgWide
	// says the route answers about every organization rather than only that
	// one.
	//
	// An explanation that left these out would be a wrong answer once the same
	// role name exists in two tenants: "they are in group X, which holds role
	// Y" does not say which X.
	OrgId   string
	OrgName string
	OrgWide bool
}

func (p AccessPath) grant() grant {
	return grant{Object: p.Object, Action: p.Action, ExpiresAt: p.ExpiresAt, OrgId: p.OrgId, OrgWide: p.OrgWide}
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

// Scope is the set of organizations a subject may be shown data from.
//
// Enforce answers "may they do this?"; a listing has to answer "and about
// whom?". The two are the same question asked of different things, so the scope
// is derived from the same policies Enforce decides with rather than from a
// second query - which is also what stops the two from disagreeing about a
// grant that has just lapsed.
type Scope struct {
	// All is true when the subject holds a live grant through the platform
	// organization, whose grants reach every organization. Their view is not
	// narrowed at all, and OrgIds is then meaningless.
	All bool
	// OrgIds are the organizations the subject holds a live grant in.
	OrgIds []string
}

// Empty reports whether the scope admits nothing at all, which is a listing
// that must return no rows rather than every row.
func (s Scope) Empty() bool {
	return !s.All && len(s.OrgIds) == 0
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
