package user

import "context"

// contextKey is unexported so nothing outside this package can collide with it
// or read the value without going through FromContext.
type contextKey struct{}

// NewContext carries the authenticated user's stored row through a request.
//
// Authentication already has to read the row to provision the user on first
// sight, and authorization needs the same row a moment later to check that the
// account is still active. Passing it along turns two identical queries per
// request into one.
func NewContext(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, contextKey{}, u)
}

// FromContext returns the user NewContext stored. The second result is false
// when no user was stored, which is the case for any caller that did not come
// through the authentication interceptor.
func FromContext(ctx context.Context) (*User, bool) {
	u, ok := ctx.Value(contextKey{}).(*User)
	return u, ok && u != nil
}
