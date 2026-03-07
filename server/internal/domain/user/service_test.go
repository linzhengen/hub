package user

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "github.com/linzhengen/hub/server/internal/domain/trans"
	"github.com/linzhengen/hub/server/internal/domain/user/usergroup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

// MockUserRepository is a mock of user.Repository.
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindOne(ctx context.Context, id string) (*User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, u *User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, u *User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockUserGroupRepository is a mock of usergroup.Repository.
type MockUserGroupRepository struct {
	mock.Mock
}

func (m *MockUserGroupRepository) AssignGroup(ctx context.Context, userId, groupId string) error {
	args := m.Called(ctx, userId, groupId)
	return args.Error(0)
}

func (m *MockUserGroupRepository) FindByUserId(ctx context.Context, userId string) (usergroup.UserGroups, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(usergroup.UserGroups), args.Error(1)
}

func (m *MockUserGroupRepository) UnassignGroup(ctx context.Context, userId, groupId string) error {
	args := m.Called(ctx, userId, groupId)
	return args.Error(0)
}

func (m *MockUserGroupRepository) Upsert(ctx context.Context, userId string, groupIds []string) error {
	args := m.Called(ctx, userId, groupIds)
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

func TestUserService_CreateIfNotExists(t *testing.T) {
	ctx := context.Background()
	testUser := &User{Id: "test-id", Username: "test-user"}

	tests := []struct {
		name          string
		setupMocks    func(transRepo *MockTransRepository, userRepo *MockUserRepository, userGroupRepo *MockUserGroupRepository)
		expectedError error
	}{
		{
			name: "Success: user already exists",
			setupMocks: func(transRepo *MockTransRepository, userRepo *MockUserRepository, userGroupRepo *MockUserGroupRepository) {
				userRepo.On("FindOne", ctx, testUser.Id).Return(testUser, nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "Success: create new user if not exists",
			setupMocks: func(transRepo *MockTransRepository, userRepo *MockUserRepository, userGroupRepo *MockUserGroupRepository) {
				userRepo.On("FindOne", ctx, testUser.Id).Return(nil, sql.ErrNoRows).Once()
				userRepo.On("Create", ctx, testUser).Return(nil).Once()
				userGroupRepo.On("Upsert", ctx, testUser.Id, mock.AnythingOfType("[]string")).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "Failure: FindOne returns unexpected error",
			setupMocks: func(transRepo *MockTransRepository, userRepo *MockUserRepository, userGroupRepo *MockUserGroupRepository) {
				userRepo.On("FindOne", ctx, testUser.Id).Return(nil, errors.New("db error")).Once()
			},
			expectedError: errors.New("db error"),
		},
		{
			name: "Failure: Create returns error",
			setupMocks: func(transRepo *MockTransRepository, userRepo *MockUserRepository, userGroupRepo *MockUserGroupRepository) {
				userRepo.On("FindOne", ctx, testUser.Id).Return(nil, sql.ErrNoRows).Once()
				userRepo.On("Create", ctx, testUser).Return(errors.New("create error")).Once()
			},
			expectedError: errors.New("create error"),
		},
		{
			name: "Failure: AssignGroup returns error",
			setupMocks: func(transRepo *MockTransRepository, userRepo *MockUserRepository, userGroupRepo *MockUserGroupRepository) {
				userRepo.On("FindOne", ctx, testUser.Id).Return(nil, sql.ErrNoRows).Once()
				userRepo.On("Create", ctx, testUser).Return(nil).Once()
				userGroupRepo.On("Upsert", ctx, testUser.Id, mock.AnythingOfType("[]string")).Return(errors.New("assign error")).Once()
			},
			expectedError: errors.New("assign error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transRepo := new(MockTransRepository)
			userRepo := new(MockUserRepository)
			userGroupRepo := new(MockUserGroupRepository)

			tt.setupMocks(transRepo, userRepo, userGroupRepo)

			service := NewService(transRepo, userRepo, userGroupRepo)
			err := service.CreateIfNotExists(ctx, testUser)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			transRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			userGroupRepo.AssertExpectations(t)
		})
	}
}
