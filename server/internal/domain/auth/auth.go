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
