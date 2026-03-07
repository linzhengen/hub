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

func TestResourceUseCase_List_WithPagination(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer func() {
		err := db.Close()
		if err != nil {
			t.Errorf("error closing db: %v", err)
		}
	}()

	dialect := goqu.Dialect("mysql")

	uc := &resourceUseCase{
		db:             db,
		dialectWrapper: dialect,
	}

	// Test with limit > 0
	params := &ListResourceQueryParams{
		Limit:  10,
		Offset: 20,
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "parent_id", "name", "identifier", "type", "path", "component", "display_order", "description", "metadata", "status", "created_at", "updated_at"}).
		AddRow("resource1", "", "Resource 1", "api.test", "api", nil, nil, nil, "Description 1", []byte("{}"), "active", now, now)

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)

	countQuery, countArgs, _ := dialect.From("resources").Select(goqu.COUNT("*")).Prepared(true).ToSQL()
	query, args, _ := dialect.From("resources").Select("*").Order(goqu.C("display_order").Asc()).Limit(uint(params.Limit)).Offset(uint(params.Offset)).Prepared(true).ToSQL()

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

	resources, total, err := uc.List(ctx, params)

	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, resources, 1)
	require.Equal(t, "resource1", resources[0].Id)
}

func TestResourceUseCase_List_WithoutPagination(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer func() {
		err := db.Close()
		if err != nil {
			t.Errorf("error closing db: %v", err)
		}
	}()

	dialect := goqu.Dialect("mysql")

	uc := &resourceUseCase{
		db:             db,
		dialectWrapper: dialect,
	}

	// Test with limit = 0 (no pagination)
	params := &ListResourceQueryParams{
		Limit:  0,
		Offset: 0,
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "parent_id", "name", "identifier", "type", "path", "component", "display_order", "description", "metadata", "status", "created_at", "updated_at"}).
		AddRow("resource1", "", "Resource 1", "api.test", "api", nil, nil, nil, "Description 1", []byte("{}"), "active", now, now).
		AddRow("resource2", "", "Resource 2", "api.test2", "api", nil, nil, nil, "Description 2", []byte("{}"), "active", now, now)

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)

	countQuery, countArgs, _ := dialect.From("resources").Select(goqu.COUNT("*")).Prepared(true).ToSQL()
	// No limit or offset should be applied
	query, args, _ := dialect.From("resources").Select("*").Order(goqu.C("display_order").Asc()).Prepared(true).ToSQL()

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

	resources, total, err := uc.List(ctx, params)

	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, resources, 2)
	require.Equal(t, "resource1", resources[0].Id)
	require.Equal(t, "resource2", resources[1].Id)
}
