package system

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/linzhengen/hub/server/internal/domain/system/permission"
	"github.com/linzhengen/hub/server/internal/domain/system/resource"
	"github.com/linzhengen/hub/server/internal/domain/system/resource/api"
	"github.com/linzhengen/hub/server/internal/domain/system/resource/menu"
)

// MockApiRepository is a mock of api.Repository.
type MockApiRepository struct {
	mock.Mock
}

func (m *MockApiRepository) FindAll(ctx context.Context) (api.APIs, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(api.APIs), args.Error(1)
}

// MockMenuRepository is a mock of menu.Repository.
type MockMenuRepository struct {
	mock.Mock
}

func (m *MockMenuRepository) FindAll(ctx context.Context) (menu.Menus, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(menu.Menus), args.Error(1)
}

// MockResourceRepository is a mock of resource.Repository.
type MockResourceRepository struct {
	mock.Mock
}

func (m *MockResourceRepository) FindOne(ctx context.Context, id string) (*resource.Resource, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*resource.Resource), args.Error(1)
}

func (m *MockResourceRepository) FindOneByIdentifier(ctx context.Context, identifier string) (*resource.Resource, error) {
	args := m.Called(ctx, identifier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*resource.Resource), args.Error(1)
}

func (m *MockResourceRepository) Create(ctx context.Context, u *resource.Resource) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockResourceRepository) Update(ctx context.Context, u *resource.Resource) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockResourceRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockPermissionRepository is a mock of permission.Repository.
type MockPermissionRepository struct {
	mock.Mock
}

func (m *MockPermissionRepository) FindOne(ctx context.Context, id string) (*permission.Permission, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*permission.Permission), args.Error(1)
}

func (m *MockPermissionRepository) FindByResourceId(ctx context.Context, resourceId string) ([]*permission.Permission, error) {
	args := m.Called(ctx, resourceId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*permission.Permission), args.Error(1)
}

func (m *MockPermissionRepository) Create(ctx context.Context, p *permission.Permission) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockPermissionRepository) Update(ctx context.Context, p *permission.Permission) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockPermissionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestResourceService_CreateResourcesAndPermissionsFromAPI(t *testing.T) {
	ctx := context.Background()
	mockApis := api.APIs{
		{Service: "user.v1.UserService", Method: "GetUser"},
		{Service: "user.v1.UserService", Method: "CreateUser"},
	}
	mockIdentifier := resource.Identifier{Api: "api", Category: "user.v1.UserService"}
	mockResource := &resource.Resource{Id: "resource-id-1", Identifier: mockIdentifier}

	tests := []struct {
		name          string
		setupMocks    func(apiRepo *MockApiRepository, resourceRepo *MockResourceRepository, permissionRepo *MockPermissionRepository)
		expectedError bool
	}{
		{
			name: "Success: create new resource and permissions",
			setupMocks: func(apiRepo *MockApiRepository, resourceRepo *MockResourceRepository, permissionRepo *MockPermissionRepository) {
				apiRepo.On("FindAll", ctx).Return(mockApis, nil).Once()
				resourceRepo.On("FindOneByIdentifier", ctx, mockIdentifier.String()).Return(nil, sql.ErrNoRows).Once()
				resourceRepo.On("Create", ctx, mock.AnythingOfType("*resource.Resource")).Return(nil).Once()
				permissionRepo.On("FindByResourceId", ctx, mock.AnythingOfType("string")).Return([]*permission.Permission{}, nil).Once()
				permissionRepo.On("Create", ctx, mock.AnythingOfType("*permission.Permission")).Return(nil).Twice()
			},
			expectedError: false,
		},
		{
			name: "Success: update existing resource, create new permissions",
			setupMocks: func(apiRepo *MockApiRepository, resourceRepo *MockResourceRepository, permissionRepo *MockPermissionRepository) {
				apiRepo.On("FindAll", ctx).Return(mockApis, nil).Once()
				resourceRepo.On("FindOneByIdentifier", ctx, mockIdentifier.String()).Return(mockResource, nil).Once()
				resourceRepo.On("Update", ctx, mock.AnythingOfType("*resource.Resource")).Return(nil).Once()
				permissionRepo.On("FindByResourceId", ctx, mockResource.Id).Return([]*permission.Permission{}, nil).Once()
				permissionRepo.On("Create", ctx, mock.AnythingOfType("*permission.Permission")).Return(nil).Twice()
			},
			expectedError: false,
		},
		{
			name: "Success: resource and permissions already exist",
			setupMocks: func(apiRepo *MockApiRepository, resourceRepo *MockResourceRepository, permissionRepo *MockPermissionRepository) {
				mockExistingPerms := []*permission.Permission{
					{ResourceId: mockResource.Id, Verb: "GetUser"},
					{ResourceId: mockResource.Id, Verb: "CreateUser"},
				}
				apiRepo.On("FindAll", ctx).Return(mockApis, nil).Once()
				resourceRepo.On("FindOneByIdentifier", ctx, mockIdentifier.String()).Return(mockResource, nil).Once()
				resourceRepo.On("Update", ctx, mock.AnythingOfType("*resource.Resource")).Return(nil).Once()
				permissionRepo.On("FindByResourceId", ctx, mockResource.Id).Return(mockExistingPerms, nil).Once()
			},
			expectedError: false,
		},
		{
			name: "Failure: apiRepo.FindAll returns error",
			setupMocks: func(apiRepo *MockApiRepository, resourceRepo *MockResourceRepository, permissionRepo *MockPermissionRepository) {
				apiRepo.On("FindAll", ctx).Return(nil, errors.New("api repo error")).Once()
			},
			expectedError: true,
		},
		{
			name: "Failure: resourceRepo.Create returns error",
			setupMocks: func(apiRepo *MockApiRepository, resourceRepo *MockResourceRepository, permissionRepo *MockPermissionRepository) {
				apiRepo.On("FindAll", ctx).Return(mockApis, nil).Once()
				resourceRepo.On("FindOneByIdentifier", ctx, mockIdentifier.String()).Return(nil, sql.ErrNoRows).Once()
				resourceRepo.On("Create", ctx, mock.AnythingOfType("*resource.Resource")).Return(errors.New("resource create error")).Once()
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiRepo := new(MockApiRepository)
			menuRepo := new(MockMenuRepository)
			resourceRepo := new(MockResourceRepository)
			permissionRepo := new(MockPermissionRepository)

			tt.setupMocks(apiRepo, resourceRepo, permissionRepo)

			service := NewResourceService(apiRepo, menuRepo, resourceRepo, permissionRepo)
			err := service.CreateResourcesAndPermissionsFromAPI(ctx)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			apiRepo.AssertExpectations(t)
			resourceRepo.AssertExpectations(t)
			permissionRepo.AssertExpectations(t)
		})
	}
}
