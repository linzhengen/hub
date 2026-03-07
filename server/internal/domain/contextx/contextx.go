package contextx

import "context"

type (
	transCtx     struct{}
	transLockCtx struct{}
	noTransCtx   struct{}
	userIDCtx    struct{}
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
