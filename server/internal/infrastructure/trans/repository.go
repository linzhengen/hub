package trans

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/linzhengen/hub/v1/server/internal/domain/contextx"
	"github.com/linzhengen/hub/v1/server/internal/domain/trans"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence"

	"github.com/linzhengen/hub/v1/server/pkg/logger"
)

func New(db *sql.DB, q persistence.Querier) trans.Repository {
	return &repository{
		db: db,
		q:  q,
	}
}

type repository struct {
	db *sql.DB
	q  persistence.Querier
}

func (a *repository) ExecTrans(ctx context.Context, fn func(context.Context) error) (txErr error) {
	if _, ok := contextx.FromTrans(ctx); ok {
		return fn(ctx)
	}

	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if txErr != nil {
			if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
				logger.Infof("rollback err: %v", err)
				txErr = fmt.Errorf("rollback err: %w", err)
			}
		}
	}()
	qTx := a.q.WithTx(tx)
	if txErr = fn(contextx.NewTrans(ctx, qTx)); txErr != nil {
		return txErr
	}
	return tx.Commit()
}

func (a *repository) ExecTransWithLock(ctx context.Context, fn func(context.Context) error) error {
	if !contextx.FromTransLock(ctx) {
		ctx = contextx.NewTransLock(ctx)
	}
	return a.ExecTrans(ctx, fn)
}
