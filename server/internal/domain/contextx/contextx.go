package contextx

import "context"

type (
	transCtx     struct{}
	transLockCtx struct{}
	noTransCtx   struct{}
	userIDCtx    struct{}
	orgIDCtx     struct{}
)

func NewTrans(ctx context.Context, trans interface{}) context.Context {
	return context.WithValue(ctx, transCtx{}, trans)
}

func FromTrans(ctx context.Context) (interface{}, bool) {
	v := ctx.Value(transCtx{})
	return v, v != nil
}

func NewTransLock(ctx context.Context) context.Context {
	return context.WithValue(ctx, transLockCtx{}, true)
}

func FromTransLock(ctx context.Context) bool {
	v := ctx.Value(transLockCtx{})
	return v != nil && v.(bool)
}

func NewNoTrans(ctx context.Context) context.Context {
	return context.WithValue(ctx, noTransCtx{}, true)
}

func FromNoTrans(ctx context.Context) bool {
	v := ctx.Value(noTransCtx{})
	return v != nil && v.(bool)
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDCtx{}, userID)
}

func GetUserID(ctx context.Context) (string, bool) {
	v := ctx.Value(userIDCtx{})
	if v != nil {
		if s, ok := v.(string); ok {
			return s, s != ""
		}
	}
	return "", false
}

// WithOrgID records which organization the request is about.
//
// It is a property of the request rather than of the caller: a person can hold
// access in several organizations, and which one they are acting in is
// something they say, not something hub infers for them. An absent value is
// therefore not an error - it is the question every caller asked before
// organizations existed, "may I, anywhere I hold access?" - and it is what
// keeps every client that predates this header working unchanged.
func WithOrgID(ctx context.Context, orgID string) context.Context {
	return context.WithValue(ctx, orgIDCtx{}, orgID)
}

// GetOrgID returns the organization the request is about, empty when it names
// none.
func GetOrgID(ctx context.Context) string {
	if v, ok := ctx.Value(orgIDCtx{}).(string); ok {
		return v
	}
	return ""
}

func FindOne[T any](
	ctx context.Context,
	id string,
	noLockFn func(ctx context.Context, id string) (*T, error),
	lockFn func(ctx context.Context, id string) (*T, error),
) (*T, error) {
	if FromTransLock(ctx) {
		return lockFn(ctx, id)
	}
	return noLockFn(ctx, id)
}
