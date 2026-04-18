package system

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/linzhengen/hub/v1/server/internal/domain/system/group"
	"github.com/linzhengen/hub/v1/server/internal/domain/system/group/grouprole"
	"github.com/linzhengen/hub/v1/server/internal/domain/user/usergroup"
)

// MockTransRepository is a mock of trans.Repository.
type MockTransRepository struct {
	mock.Mock
}

func (m *MockTransRepository) ExecTrans(ctx context.Context, fn func(context.Context) error) error {
	// Execute the actual fn to test the logic within the transaction.
	return fn(ctx)
}

func (m *MockTransRepository) ExecTransWithLock(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// MockGroupRepository is a mock type for the group.Repository type
type MockGroupRepository struct {
	mock.Mock
}

func (m *MockGroupRepository) FindOne(ctx context.Context, id string) (*group.Group, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*group.Group), args.Error(1)
}

func (m *MockGroupRepository) Create(ctx context.Context, g *group.Group) error {
	args := m.Called(ctx, g)
	return args.Error(0)
}

func (m *MockGroupRepository) Update(ctx context.Context, g *group.Group) error {
	args := m.Called(ctx, g)
	return args.Error(0)
}

func (m *MockGroupRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockGroupRoleRepository is a mock type for the grouprole.Repository type
type MockGroupRoleRepository struct {
	mock.Mock
}

func (m *MockGroupRoleRepository) FindByGroupId(ctx context.Context, groupId string) (grouprole.GroupRoles, error) {
	args := m.Called(ctx, groupId)
	return args.Get(0).(grouprole.GroupRoles), args.Error(1)
}

func (m *MockGroupRoleRepository) AssignRole(ctx context.Context, groupId, roleId string) error {
	args := m.Called(ctx, groupId, roleId)
	return args.Error(0)
}

func (m *MockGroupRoleRepository) UnassignRole(ctx context.Context, groupId, roleId string) error {
	args := m.Called(ctx, groupId, roleId)
	return args.Error(0)
}

func (m *MockGroupRoleRepository) Upsert(ctx context.Context, groupId string, roleId []string) error {
	args := m.Called(ctx, groupId, roleId)
	return args.Error(0)
}

// MockUserGroupRepository is a mock type for the usergroup.Repository type
type MockUserGroupRepository struct {
	mock.Mock
}

func (m *MockUserGroupRepository) FindByUserId(ctx context.Context, userId string) (usergroup.UserGroups, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).(usergroup.UserGroups), args.Error(1)
}

func (m *MockUserGroupRepository) AssignGroup(ctx context.Context, userId, groupId string) error {
	args := m.Called(ctx, userId, groupId)
	return args.Error(0)
}

func (m *MockUserGroupRepository) UnassignGroup(ctx context.Context, userId, groupId string) error {
	args := m.Called(ctx, userId, groupId)
	return args.Error(0)
}

func (m *MockUserGroupRepository) Upsert(ctx context.Context, userId string, groupId []string) error {
	args := m.Called(ctx, userId, groupId)
	return args.Error(0)
}

func (m *MockUserGroupRepository) AddUsersToGroup(ctx context.Context, groupID string, userIDs []string) error {
	args := m.Called(ctx, groupID, userIDs)
	return args.Error(0)
}

func (m *MockUserGroupRepository) RemoveUsersFromGroup(ctx context.Context, groupID string, userIDs []string) error {
	args := m.Called(ctx, groupID, userIDs)
	return args.Error(0)
}

func (m *MockUserGroupRepository) IsUserInGroup(ctx context.Context, groupID string, userID string) (bool, error) {
	args := m.Called(ctx, groupID, userID)
	return args.Bool(0), args.Error(1)
}

func TestGroupUseCase_AddUsersToGroup(t *testing.T) {
	ctx := context.Background()
	groupID := "test-group-id"
	userIDs := []string{"user1", "user2"}

	mockGroupRepo := new(MockGroupRepository)
	mockUserGroupRepo := new(MockUserGroupRepository)
	mockTransRepo := new(MockTransRepository)

	uc := &groupUseCase{
		groupRepo:     mockGroupRepo,
		userGroupRepo: mockUserGroupRepo,
		transRepo:     mockTransRepo,
	}

	mockUserGroupRepo.On("AddUsersToGroup", ctx, groupID, userIDs).Return(nil)
	mockGroupRepo.On("FindOne", ctx, groupID).Return(&group.Group{Id: groupID}, nil)

	_, err := uc.AddUsersToGroup(ctx, groupID, userIDs)

	assert.NoError(t, err)
	mockUserGroupRepo.AssertExpectations(t)
	mockGroupRepo.AssertExpectations(t)
}

func TestGroupUseCase_RemoveUsersFromGroup(t *testing.T) {
	ctx := context.Background()
	groupID := "test-group-id"
	userIDs := []string{"user1", "user2"}

	mockGroupRepo := new(MockGroupRepository)
	mockUserGroupRepo := new(MockUserGroupRepository)
	mockTransRepo := new(MockTransRepository)

	uc := &groupUseCase{
		groupRepo:     mockGroupRepo,
		userGroupRepo: mockUserGroupRepo,
		transRepo:     mockTransRepo,
	}

	mockUserGroupRepo.On("RemoveUsersFromGroup", ctx, groupID, userIDs).Return(nil)
	mockGroupRepo.On("FindOne", ctx, groupID).Return(&group.Group{Id: groupID}, nil)

	_, err := uc.RemoveUsersFromGroup(ctx, groupID, userIDs)

	assert.NoError(t, err)
	mockUserGroupRepo.AssertExpectations(t)
	mockGroupRepo.AssertExpectations(t)
}

func TestGroupUseCase_List_WithPagination(t *testing.T) {
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

	// Create a mock for groupRoleRepo
	mockGroupRoleRepo := new(MockGroupRoleRepository)

	uc := &groupUseCase{
		db:             db,
		dialectWrapper: dialect,
		groupRoleRepo:  mockGroupRoleRepo,
	}

	// Test with limit > 0
	params := &ListGroupQueryParams{
		Limit:  10,
		Offset: 20,
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "description", "status", "created_at", "updated_at"}).
		AddRow("group1", "Group 1", "Description 1", "active", now, now)

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)

	countQuery, countArgs, _ := dialect.From("groups").Select(goqu.COUNT("*")).Prepared(true).ToSQL()
	query, args, _ := dialect.From("groups").Select("*").Limit(uint(params.Limit)).Offset(uint(params.Offset)).Prepared(true).ToSQL()

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

	// Expect the query for group roles
	grQuery, grArgs, _ := dialect.From("group_roles").
		Select("group_id", "role_id").
		Where(goqu.Ex{"group_id": []string{"group1"}}).
		Prepared(true).ToSQL()

	// convert []interface{} to []driver.Value
	grDriverArgs := make([]driver.Value, len(grArgs))
	for i, arg := range grArgs {
		grDriverArgs[i] = arg
	}

	// Create empty rows for group roles
	grRows := sqlmock.NewRows([]string{"group_id", "role_id"})

	mock.ExpectQuery(grQuery).WithArgs(grDriverArgs...).WillReturnRows(grRows)
	mock.ExpectClose()

	groups, total, err := uc.List(ctx, params)

	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, groups, 1)
	require.Equal(t, "group1", groups[0].Id)
	require.Equal(t, "Group 1", groups[0].Name)
}

func TestGroupUseCase_List_WithoutPagination(t *testing.T) {
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

	// Create a mock for groupRoleRepo
	mockGroupRoleRepo := new(MockGroupRoleRepository)

	uc := &groupUseCase{
		db:             db,
		dialectWrapper: dialect,
		groupRoleRepo:  mockGroupRoleRepo,
	}

	// Test with limit = 0 (no pagination)
	params := &ListGroupQueryParams{
		Limit:  0,
		Offset: 0,
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "description", "status", "created_at", "updated_at"}).
		AddRow("group1", "Group 1", "Description 1", "active", now, now).
		AddRow("group2", "Group 2", "Description 2", "active", now, now)

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)

	countQuery, countArgs, _ := dialect.From("groups").Select(goqu.COUNT("*")).Prepared(true).ToSQL()
	// No limit or offset should be applied
	query, args, _ := dialect.From("groups").Select("*").Prepared(true).ToSQL()

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

	// Expect the query for group roles
	grQuery, grArgs, _ := dialect.From("group_roles").
		Select("group_id", "role_id").
		Where(goqu.Ex{"group_id": []string{"group1", "group2"}}).
		Prepared(true).ToSQL()

	// convert []interface{} to []driver.Value
	grDriverArgs := make([]driver.Value, len(grArgs))
	for i, arg := range grArgs {
		grDriverArgs[i] = arg
	}

	// Create empty rows for group roles
	grRows := sqlmock.NewRows([]string{"group_id", "role_id"})

	mock.ExpectQuery(grQuery).WithArgs(grDriverArgs...).WillReturnRows(grRows)
	mock.ExpectClose()

	groups, total, err := uc.List(ctx, params)

	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, groups, 2)
	require.Equal(t, "group1", groups[0].Id)
	require.Equal(t, "group2", groups[1].Id)
}
