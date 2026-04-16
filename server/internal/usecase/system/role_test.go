package system

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	"github.com/linzhengen/hub/v1/server/internal/domain/system/role/rolepermission"
	"github.com/stretchr/testify/require"
)

// mockRolePermissionRepo is a mock implementation of rolepermission.Repository
// that returns empty results for all methods
type mockRolePermissionRepo struct{}

func (m *mockRolePermissionRepo) FindByRoleId(ctx context.Context, roleId string) (rolepermission.RolePermissions, error) {
	return rolepermission.RolePermissions{}, nil
}

func (m *mockRolePermissionRepo) AssignPermission(ctx context.Context, roleId, permissionId string) error {
	return nil
}

func (m *mockRolePermissionRepo) UnassignPermission(ctx context.Context, roleId, permissionId string) error {
	return nil
}

func (m *mockRolePermissionRepo) Upsert(ctx context.Context, roleId string, permissionId []string) error {
	return nil
}

func (m *mockRolePermissionRepo) IsPermissionInRole(ctx context.Context, roleId string, permissionId string) (bool, error) {
	return false, nil
}

func TestRoleUseCase_List_WithPagination(t *testing.T) {
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

	uc := &roleUseCase{
		db:                 db,
		dialectWrapper:     dialect,
		rolePermissionRepo: &mockRolePermissionRepo{},
	}

	// Test with limit > 0
	params := &ListRoleQueryParams{
		Limit:  10,
		Offset: 20,
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).
		AddRow("role1", "Role 1", "Description 1", now, now)

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)

	countQuery, countArgs, _ := dialect.From("roles").Select(goqu.COUNT("*")).Prepared(true).ToSQL()
	query, args, _ := dialect.From("roles").Select("*").Limit(uint(params.Limit)).Offset(uint(params.Offset)).Prepared(true).ToSQL()

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

	// Mock the role-permission query
	rpQuery, rpArgs, _ := dialect.From("role_permissions").
		Select("role_id", "permission_id").
		Where(goqu.Ex{"role_id": []string{"role1"}}).
		Prepared(true).ToSQL()

	rpDriverArgs := make([]driver.Value, len(rpArgs))
	for i, arg := range rpArgs {
		rpDriverArgs[i] = arg
	}

	rpRows := sqlmock.NewRows([]string{"role_id", "permission_id"})
	mock.ExpectQuery(rpQuery).WithArgs(rpDriverArgs...).WillReturnRows(rpRows)

	mock.ExpectClose()

	roles, total, err := uc.List(ctx, params)

	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, roles, 1)
	require.Equal(t, "role1", roles[0].Id)
	require.Equal(t, "Role 1", roles[0].Name)
}

func TestRoleUseCase_List_WithoutPagination(t *testing.T) {
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

	uc := &roleUseCase{
		db:                 db,
		dialectWrapper:     dialect,
		rolePermissionRepo: &mockRolePermissionRepo{},
	}

	// Test with limit = 0 (no pagination)
	params := &ListRoleQueryParams{
		Limit:  0,
		Offset: 0,
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).
		AddRow("role1", "Role 1", "Description 1", now, now).
		AddRow("role2", "Role 2", "Description 2", now, now)

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)

	countQuery, countArgs, _ := dialect.From("roles").Select(goqu.COUNT("*")).Prepared(true).ToSQL()
	// No limit or offset should be applied
	query, args, _ := dialect.From("roles").Select("*").Prepared(true).ToSQL()

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

	// Mock the role-permission query
	rpQuery, rpArgs, _ := dialect.From("role_permissions").
		Select("role_id", "permission_id").
		Where(goqu.Ex{"role_id": []string{"role1", "role2"}}).
		Prepared(true).ToSQL()

	rpDriverArgs := make([]driver.Value, len(rpArgs))
	for i, arg := range rpArgs {
		rpDriverArgs[i] = arg
	}

	rpRows := sqlmock.NewRows([]string{"role_id", "permission_id"})
	mock.ExpectQuery(rpQuery).WithArgs(rpDriverArgs...).WillReturnRows(rpRows)

	mock.ExpectClose()

	roles, total, err := uc.List(ctx, params)

	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, roles, 2)
	require.Equal(t, "role1", roles[0].Id)
	require.Equal(t, "role2", roles[1].Id)
}
