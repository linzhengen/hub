package persistence

import (
	"context"
	"database/sql"

	"github.com/doug-martin/goqu/v9"

	"github.com/linzhengen/hub/v1/server/config"
	"github.com/linzhengen/hub/v1/server/internal/domain/contextx"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence/postgres"
	"github.com/linzhengen/hub/v1/server/pkg/logger"
)

// DialectWrapper is a type alias for goqu.DialectWrapper
type DialectWrapper = goqu.DialectWrapper

// NewConnection creates a new database connection based on the configuration
func NewConnection(cfg config.EnvConfig) (*sql.DB, error) {
	logger.Info("Using PostgreSQL database")
	return postgres.NewConn(cfg.PostgreSQL)
}

// GetConnection returns the database connection
func GetConnection(db *sql.DB) *sql.DB {
	return db
}

// GetQ returns the appropriate Querier from context or the default one
func GetQ(ctx context.Context, q Querier) Querier {
	if trTX, ok := contextx.FromTrans(ctx); ok {
		if tq, ok := trTX.(Querier); ok {
			return tq
		}
	}
	return q
}
