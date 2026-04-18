package postgres

import (
	"context"
	"database/sql"

	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence/postgres/sqlc"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"

	"github.com/linzhengen/hub/v1/server/config"
)

func NewConn(cfg config.PostgreSQL) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(cfg.MaxLifetime)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	return db, nil
}

type DialectWrapper = goqu.DialectWrapper

func NewDialect() DialectWrapper {
	return goqu.Dialect("postgres")
}

func SelectCount(ctx context.Context, db *sql.DB, b *goqu.SelectDataset) (int64, error) {
	b = b.Select(goqu.COUNT("*"))
	cntQuery, cntQueryParams, err := b.Prepared(true).ToSQL()
	if err != nil {
		return 0, err
	}
	row := db.QueryRowContext(ctx, cntQuery, cntQueryParams...)
	var cnt int64
	if err := row.Scan(&cnt); err != nil {
		return 0, err
	}
	return cnt, nil
}

// NewQuerier creates a new sqlc.Queries instance for PostgreSQL
func NewQuerier(db *sql.DB) *sqlc.Queries {
	return sqlc.New(db)
}
