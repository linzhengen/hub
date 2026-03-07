package usecase

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserUseCase_List_WithGroupIds(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	dialect := goqu.Dialect("mysql")

	uc := &userUseCase{
		db:             db,
		dialectWrapper: dialect,
	}

	params := &ListUserQueryParams{
		GroupIds: []string{"group1", "group2"},
	}

	rows := sqlmock.NewRows([]string{"id", "username", "email", "status", "created_at", "updated_at"}).
		AddRow("user1", "user1", "user1@example.com", "active", time.Now(), time.Now())

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)

	// Create a subquery to check if the user belongs to any of the specified groups
	subquery := dialect.From("user_groups").
		Select(goqu.L("1")).
		Where(goqu.Ex{
			"user_groups.user_id":  goqu.I("users.id"),
			"user_groups.group_id": params.GroupIds,
		})

	// Use EXISTS with the subquery
	countQuery, countArgs, _ := dialect.From("users").
		Select(goqu.COUNT("users.id")).
		Where(goqu.L("EXISTS ?", subquery)).
		Prepared(true).ToSQL()

	query, args, _ := dialect.From("users").
		Select("users.id", "users.username", "users.email", "users.status", "users.created_at", "users.updated_at").
		Where(goqu.L("EXISTS ?", subquery)).
		Prepared(true).ToSQL()

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

	// Expect the query for user-group relationships
	ugQuery, ugArgs, _ := dialect.From("user_groups").
		Select("user_id", "group_id").
		Where(goqu.Ex{"user_id": []string{"user1"}}).
		Prepared(true).ToSQL()

	ugDriverArgs := make([]driver.Value, len(ugArgs))
	for i, arg := range ugArgs {
		ugDriverArgs[i] = arg
	}

	ugRows := sqlmock.NewRows([]string{"user_id", "group_id"}).
		AddRow("user1", "group1")

	mock.ExpectQuery(ugQuery).WithArgs(ugDriverArgs...).WillReturnRows(ugRows)
	mock.ExpectClose()

	users, total, err := uc.List(ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, users, 1)

	assert.NoError(t, db.Close())
}

func TestUserUseCase_List_WithPagination(t *testing.T) {
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

	uc := &userUseCase{
		db:             db,
		dialectWrapper: dialect,
	}

	// Test with limit > 0
	params := &ListUserQueryParams{
		Limit:  10,
		Offset: 20,
	}

	rows := sqlmock.NewRows([]string{"id", "username", "email", "status", "created_at", "updated_at"}).
		AddRow("user1", "user1", "user1@example.com", "active", time.Now(), time.Now())

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)

	countQuery, countArgs, _ := dialect.From("users").Select(goqu.COUNT("users.id")).Prepared(true).ToSQL()
	query, args, _ := dialect.From("users").Select("users.id", "users.username", "users.email", "users.status", "users.created_at", "users.updated_at").Limit(uint(params.Limit)).Offset(uint(params.Offset)).Prepared(true).ToSQL()

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

	// Expect the query for user-group relationships
	ugQuery, ugArgs, _ := dialect.From("user_groups").
		Select("user_id", "group_id").
		Where(goqu.Ex{"user_id": []string{"user1"}}).
		Prepared(true).ToSQL()

	ugDriverArgs := make([]driver.Value, len(ugArgs))
	for i, arg := range ugArgs {
		ugDriverArgs[i] = arg
	}

	ugRows := sqlmock.NewRows([]string{"user_id", "group_id"}).
		AddRow("user1", "group1")

	mock.ExpectQuery(ugQuery).WithArgs(ugDriverArgs...).WillReturnRows(ugRows)
	mock.ExpectClose()

	users, total, err := uc.List(ctx, params)

	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, users, 1)
}

func TestUserUseCase_List_WithoutPagination(t *testing.T) {
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

	uc := &userUseCase{
		db:             db,
		dialectWrapper: dialect,
	}

	// Test with limit = 0 (no pagination)
	params := &ListUserQueryParams{
		Limit:  0,
		Offset: 0,
	}

	rows := sqlmock.NewRows([]string{"id", "username", "email", "status", "created_at", "updated_at"}).
		AddRow("user1", "user1", "user1@example.com", "active", time.Now(), time.Now()).
		AddRow("user2", "user2", "user2@example.com", "active", time.Now(), time.Now())

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)

	countQuery, countArgs, _ := dialect.From("users").Select(goqu.COUNT("users.id")).Prepared(true).ToSQL()
	// No limit or offset should be applied
	query, args, _ := dialect.From("users").Select("users.id", "users.username", "users.email", "users.status", "users.created_at", "users.updated_at").Prepared(true).ToSQL()

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

	// Expect the query for user-group relationships
	ugQuery, ugArgs, _ := dialect.From("user_groups").
		Select("user_id", "group_id").
		Where(goqu.Ex{"user_id": []string{"user1", "user2"}}).
		Prepared(true).ToSQL()

	ugDriverArgs := make([]driver.Value, len(ugArgs))
	for i, arg := range ugArgs {
		ugDriverArgs[i] = arg
	}

	ugRows := sqlmock.NewRows([]string{"user_id", "group_id"}).
		AddRow("user1", "group1").
		AddRow("user2", "group2")

	mock.ExpectQuery(ugQuery).WithArgs(ugDriverArgs...).WillReturnRows(ugRows)
	mock.ExpectClose()

	users, total, err := uc.List(ctx, params)

	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, users, 2)
}
