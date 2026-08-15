package auth

// Policy represents a Casbin policy rule
type Policy struct {
	Subject string // User or role
	Object  string // Resource
	Action  string // Action (e.g., read, write, delete)
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
}

// Principal is a user allowed some operation, with every route that allows it.
type Principal struct {
	UserId   string
	Username string
	Paths    []AccessPath
}

// Membership is one user's place in one group. It carries no permission of its
// own; it is the edge that joins a user to the paths their groups hold.
type Membership struct {
	UserId   string
	Username string
	GroupId  string
}
