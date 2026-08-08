package system

import (
	"context"
	"testing"

	"github.com/linzhengen/hub/server/internal/domain/system/resource"
	"github.com/linzhengen/hub/server/internal/usecase/system"
	pbv1 "github.com/linzhengen/hub/server/pb/system/resource/v1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockResourceUseCase struct {
	mock.Mock
}

func (m *mockResourceUseCase) Get(ctx context.Context, resourceId string) (*resource.Resource, error) {
	args := m.Called(ctx, resourceId)
	return args.Get(0).(*resource.Resource), args.Error(1)
}

func (m *mockResourceUseCase) Create(ctx context.Context, r *resource.Resource) (*resource.Resource, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*resource.Resource), args.Error(1)
}

func (m *mockResourceUseCase) Update(ctx context.Context, r *resource.Resource) (*resource.Resource, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*resource.Resource), args.Error(1)
}

func (m *mockResourceUseCase) Delete(ctx context.Context, resourceId string) error {
	args := m.Called(ctx, resourceId)
	return args.Error(0)
}

func (m *mockResourceUseCase) List(ctx context.Context, params *system.ListResourceQueryParams) ([]*resource.Resource, int64, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]*resource.Resource), args.Get(1).(int64), args.Error(2)
}

func (m *mockResourceUseCase) CreateMenu(ctx context.Context, r *resource.Resource) (*resource.Resource, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*resource.Resource), args.Error(1)
}

func (m *mockResourceUseCase) UpdateMenu(ctx context.Context, r *resource.Resource) (*resource.Resource, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*resource.Resource), args.Error(1)
}

func TestResourceHandler_UpdateResource_NilIdentifier(t *testing.T) {
	mockUC := new(mockResourceUseCase)
	handler := NewResourceHandler(mockUC)

	req := &pbv1.UpdateResourceRequest{
		Id:     "test-id",
		Name:   "test-name",
		Status: pbv1.Status_STATUS_ACTIVE,
		// Identifier is nil
	}

	mockUC.On("Update", mock.Anything, mock.MatchedBy(func(r *resource.Resource) bool {
		return r.Id == "test-id" && r.Identifier.Api == "" && r.Identifier.Category == ""
	})).Return(&resource.Resource{Id: "test-id"}, nil)

	_, err := handler.UpdateResource(context.Background(), req)
	require.NoError(t, err)
	mockUC.AssertExpectations(t)
}

func TestResourceHandler_CreateResource_NilIdentifier(t *testing.T) {
	mockUC := new(mockResourceUseCase)
	handler := NewResourceHandler(mockUC)

	req := &pbv1.CreateResourceRequest{
		Name: "test-name",
		// Identifier is nil
	}

	mockUC.On("Create", mock.Anything, mock.MatchedBy(func(r *resource.Resource) bool {
		return r.Name == "test-name" && r.Identifier.Api == "" && r.Identifier.Category == ""
	})).Return(&resource.Resource{Id: "test-id", Name: "test-name"}, nil)

	_, err := handler.CreateResource(context.Background(), req)
	require.NoError(t, err)
	mockUC.AssertExpectations(t)
}
