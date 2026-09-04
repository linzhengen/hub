package postgres

import (
	"context"
	"database/sql"

	"github.com/linzhengen/hub/server/internal/infrastructure/persistence/postgres/sqlc"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"

	"github.com/linzhengen/hub/server/config"
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

// In is `column IN (values)`, and matches nothing when values is empty.
//
// goqu renders an empty slice as `IN ()`, which PostgreSQL rejects as a syntax
// error rather than returning no rows - so a listing narrowed to an empty set
// used to fail with a 500 instead of an empty page. `GetMe` was the case that
// showed it: a principal in no group asks for its groups, the set is empty, and
// the one rpc that is public precisely so a caller with no grants can read its
// own profile was the one that broke.
//
// The rule is written here rather than at each call site because it was
// previously written at two of them, by hand, and not at the third.
func In(column string, values []string) goqu.Expression {
	if len(values) == 0 {
		return goqu.L("FALSE")
	}
	return goqu.Ex{column: values}
}
