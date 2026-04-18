package system

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	"github.com/stretchr/testify/require"
)

func TestPermissionUseCase_List_WithPagination(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer func() {
		err := db.Close()
		if err != nil {
			t.Errorf("error closing db: %v", err)
		}
	}()

	dialect := goqu.Dialect("postgres")

	uc := &permissionUseCase{
		db:             db,
		dialectWrapper: dialect,
	}

	// Test with limit > 0
	params := &ListPermissionQueryParams{
		Limit:  10,
		Offset: 20,
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "verb", "resource_id", "description", "created_at", "updated_at"}).
		AddRow("perm1", "read", "resource1", "Description 1", now, now)

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)

	countQuery, countArgs, _ := dialect.From("permissions").Select(goqu.COUNT("*")).Prepared(true).ToSQL()
	query, args, _ := dialect.From("permissions").Select("*").Limit(uint(params.Limit)).Offset(uint(params.Offset)).Prepared(true).ToSQL()

	// convert []interface{} to []driver.Value
	countDriverArgs := make([]driver.Value, len(countArgs))
	for i, arg := range countArgs {
		countDriverArgs[i] = arg
	}
	driverArgs := make([]driver.Value, len(args))
	for i, arg := range args {
		driverArgs[i] = arg
	}

	mock.ExpectQuery(countQuery).WithArgs(countDriverArgs...).WillReturnRows(countRows)
	mock.ExpectQuery(query).WithArgs(driverArgs...).WillReturnRows(rows)
	mock.ExpectClose()

	permissions, total, err := uc.List(ctx, params)

	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, permissions, 1)
	require.Equal(t, "perm1", permissions[0].Id)
}

func TestPermissionUseCase_List_WithoutPagination(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer func() {
		err := db.Close()
		if err != nil {
			t.Errorf("error closing db: %v", err)
		}
	}()

	dialect := goqu.Dialect("postgres")

	uc := &permissionUseCase{
		db:             db,
		dialectWrapper: dialect,
	}

	// Test with limit = 0 (no pagination)
	params := &ListPermissionQueryParams{
		Limit:  0,
		Offset: 0,
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "verb", "resource_id", "description", "created_at", "updated_at"}).
		AddRow("perm1", "read", "resource1", "Description 1", now, now).
		AddRow("perm2", "write", "resource1", "Description 2", now, now)

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)

	countQuery, countArgs, _ := dialect.From("permissions").Select(goqu.COUNT("*")).Prepared(true).ToSQL()
	// No limit or offset should be applied
	query, args, _ := dialect.From("permissions").Select("*").Prepared(true).ToSQL()

	// convert []interface{} to []driver.Value
	countDriverArgs := make([]driver.Value, len(countArgs))
	for i, arg := range countArgs {
		countDriverArgs[i] = arg
	}
	driverArgs := make([]driver.Value, len(args))
	for i, arg := range args {
		driverArgs[i] = arg
	}

	mock.ExpectQuery(countQuery).WithArgs(countDriverArgs...).WillReturnRows(countRows)
	mock.ExpectQuery(query).WithArgs(driverArgs...).WillReturnRows(rows)
	mock.ExpectClose()

	permissions, total, err := uc.List(ctx, params)

	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, permissions, 2)
	require.Equal(t, "perm1", permissions[0].Id)
	require.Equal(t, "perm2", permissions[1].Id)
}
