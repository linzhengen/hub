package trans

import "context"

type Repository interface {
	ExecTrans(ctx context.Context, fn func(context.Context) error) error
	ExecTransWithLock(ctx context.Context, fn func(context.Context) error) error
}
