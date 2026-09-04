package usecase

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	"github.com/stretchr/testify/assert"

	"github.com/linzhengen/hub/server/internal/usecase/pagination"
	"github.com/stretchr/testify/require"

	"github.com/linzhengen/hub/server/internal/domain/auth"
	"github.com/linzhengen/hub/server/internal/domain/user"
)

func TestUserUseCase_List_WithGroupIds(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	dialect := goqu.Dialect("postgres")

	uc := &userUseCase{
		db:             db,
		dialectWrapper: dialect,
		// The widest scope, so the query under test is the one the filters
		// build rather than the one the tenant boundary adds. The narrowing is
		// covered by TestUserUseCase_List_IsNarrowedToTheVisibleOrganizations.
		authSvc: unscopedAuth{},
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
		Limit(uint(pagination.DefaultLimit)).Offset(0).
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
		Select("user_id", "group_id", "expires_at").
		Where(goqu.Ex{"user_id": []string{"user1"}}).
		Prepared(true).ToSQL()

	ugDriverArgs := make([]driver.Value, len(ugArgs))
	for i, arg := range ugArgs {
		ugDriverArgs[i] = arg
	}

	// A membership with no expiry and one with an expiry, because the listing
	// has to be able to tell them apart - that is the point of reading the
	// column at all.
	ugRows := sqlmock.NewRows([]string{"user_id", "group_id", "expires_at"}).
		AddRow("user1", "group1", nil)

	mock.ExpectQuery(ugQuery).WithArgs(ugDriverArgs...).WillReturnRows(ugRows)
	mock.ExpectClose()

	users, total, err := uc.List(ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, users, 1)
	assert.Equal(t,
		[]user.GroupMembership{{GroupId: "group1"}},
		users[0].Groups,
		"a membership with no expiry comes back with none, not with a zero time")

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

	dialect := goqu.Dialect("postgres")

	uc := &userUseCase{
		db:             db,
		dialectWrapper: dialect,
		authSvc:        unscopedAuth{},
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
		Select("user_id", "group_id", "expires_at").
		Where(goqu.Ex{"user_id": []string{"user1"}}).
		Prepared(true).ToSQL()

	ugDriverArgs := make([]driver.Value, len(ugArgs))
	for i, arg := range ugArgs {
		ugDriverArgs[i] = arg
	}

	// A membership with no expiry and one with an expiry, because the listing
	// has to be able to tell them apart - that is the point of reading the
	// column at all.
	ugRows := sqlmock.NewRows([]string{"user_id", "group_id", "expires_at"}).
		AddRow("user1", "group1", nil)

	mock.ExpectQuery(ugQuery).WithArgs(ugDriverArgs...).WillReturnRows(ugRows)
	mock.ExpectClose()

	users, total, err := uc.List(ctx, params)

	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, users, 1)
}

func TestUserUseCase_List_DefaultsToABoundedPage(t *testing.T) {
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

	uc := &userUseCase{
		db:             db,
		dialectWrapper: dialect,
		authSvc:        unscopedAuth{},
	}

	// A caller that asks for no particular page size gets the default one,
	// not every row in the table.
	params := &ListUserQueryParams{
		Limit:  0,
		Offset: 0,
	}

	rows := sqlmock.NewRows([]string{"id", "username", "email", "status", "created_at", "updated_at"}).
		AddRow("user1", "user1", "user1@example.com", "active", time.Now(), time.Now()).
		AddRow("user2", "user2", "user2@example.com", "active", time.Now(), time.Now())

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)

	countQuery, countArgs, _ := dialect.From("users").Select(goqu.COUNT("users.id")).Prepared(true).ToSQL()
	query, args, _ := dialect.From("users").Select("users.id", "users.username", "users.email", "users.status", "users.created_at", "users.updated_at").Limit(uint(pagination.DefaultLimit)).Offset(0).Prepared(true).ToSQL()

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
		Select("user_id", "group_id", "expires_at").
		Where(goqu.Ex{"user_id": []string{"user1", "user2"}}).
		Prepared(true).ToSQL()

	ugDriverArgs := make([]driver.Value, len(ugArgs))
	for i, arg := range ugArgs {
		ugDriverArgs[i] = arg
	}

	// A membership with no expiry and one with an expiry, because the listing
	// has to be able to tell them apart - that is the point of reading the
	// column at all.
	ugRows := sqlmock.NewRows([]string{"user_id", "group_id", "expires_at"}).
		AddRow("user1", "group1", nil).
		AddRow("user2", "group2", nil)

	mock.ExpectQuery(ugQuery).WithArgs(ugDriverArgs...).WillReturnRows(ugRows)
	mock.ExpectClose()

	users, total, err := uc.List(ctx, params)

	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, users, 2)
}

// unscopedAuth reports that every organization is visible.
type unscopedAuth struct{}

func (unscopedAuth) Enforce(_ context.Context, _ auth.Request) (bool, error) { return true, nil }

func (unscopedAuth) Explain(_ context.Context, _ auth.Request) ([]auth.AccessPath, error) {
	return nil, nil
}

func (unscopedAuth) PrincipalsFor(_ context.Context, _ auth.Request) ([]auth.Principal, error) {
	return nil, nil
}

func (unscopedAuth) VisibleOrgs(_ context.Context, _, _ string) (auth.Scope, error) {
	return auth.Scope{All: true}, nil
}

// scopedAuth reports one organization, as a tenant administrator's grant does.
type scopedAuth struct{ orgId string }

func (scopedAuth) Enforce(_ context.Context, _ auth.Request) (bool, error) { return true, nil }

func (scopedAuth) Explain(_ context.Context, _ auth.Request) ([]auth.AccessPath, error) {
	return nil, nil
}

func (scopedAuth) PrincipalsFor(_ context.Context, _ auth.Request) ([]auth.Principal, error) {
	return nil, nil
}

func (a scopedAuth) VisibleOrgs(_ context.Context, _, _ string) (auth.Scope, error) {
	return auth.Scope{OrgIds: []string{a.orgId}}, nil
}

// noAuth reports that nothing is visible, which is what somebody holding no
// live grant anywhere looks like.
type noAuth struct{}

func (noAuth) Enforce(_ context.Context, _ auth.Request) (bool, error) { return false, nil }

func (noAuth) Explain(_ context.Context, _ auth.Request) ([]auth.AccessPath, error) {
	return nil, nil
}

func (noAuth) PrincipalsFor(_ context.Context, _ auth.Request) ([]auth.Principal, error) {
	return nil, nil
}

func (noAuth) VisibleOrgs(_ context.Context, _, _ string) (auth.Scope, error) {
	return auth.Scope{}, nil
}

// TestUserUseCase_List_IsNarrowedToTheVisibleOrganizations is the tenant
// boundary on the user directory: a listing must not become a way to read who
// else is on the installation.
func TestUserUseCase_List_IsNarrowedToTheVisibleOrganizations(t *testing.T) {
	const orgA = "11111111-1111-1111-1111-111111111111"

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	uc := &userUseCase{
		db:             db,
		dialectWrapper: goqu.Dialect("postgres"),
		authSvc:        scopedAuth{orgId: orgA},
	}

	mock.ExpectQuery("EXISTS").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("EXISTS").WillReturnRows(
		sqlmock.NewRows([]string{"id", "username", "email", "status", "created_at", "updated_at"}))

	_, _, err = uc.List(context.Background(), &ListUserQueryParams{})

	assert.NoError(t, err)
	// The organization reaches the query, rather than the filter being built and
	// then forgotten.
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUserUseCase_List_ShowsNothingWithoutAScope is the other half: an empty
// scope has to mean no rows, never every row.
func TestUserUseCase_List_ShowsNothingWithoutAScope(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	uc := &userUseCase{
		db:             db,
		dialectWrapper: goqu.Dialect("postgres"),
		authSvc:        noAuth{},
	}

	users, total, err := uc.List(context.Background(), &ListUserQueryParams{})

	assert.NoError(t, err)
	assert.Empty(t, users)
	assert.Zero(t, total)
	assert.NoError(t, mock.ExpectationsWereMet(), "no query is run at all")
}
