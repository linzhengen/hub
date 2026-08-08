package system

import (
	"context"

	"github.com/linzhengen/hub/server/internal/domain/system/permission"
	"github.com/linzhengen/hub/server/internal/usecase/system"
	pbv1 "github.com/linzhengen/hub/server/pb/system/permission/v1"
)

func NewPermissionHandler(
	permissionUseCase system.PermissionUseCase,
) pbv1.PermissionServiceServer {
	return &permissionHandler{
		permissionUseCase: permissionUseCase,
	}
}

type permissionHandler struct {
	permissionUseCase system.PermissionUseCase
}

func (h permissionHandler) CreatePermission(ctx context.Context, request *pbv1.CreatePermissionRequest) (*pbv1.CreatePermissionResponse, error) {
	f, err := permission.Factory(
		request.ResourceId,
		request.Verb,
		request.Description,
	)
	if err != nil {
		return nil, err
	}
	p, err := h.permissionUseCase.Create(ctx, f)
	if err != nil {
		return nil, err
	}
	return &pbv1.CreatePermissionResponse{Permission: permissionDomainToPb(p)}, nil
}

func (h permissionHandler) DeletePermission(ctx context.Context, request *pbv1.DeletePermissionRequest) (*pbv1.DeletePermissionResponse, error) {
	if err := h.permissionUseCase.Delete(ctx, request.Id); err != nil {
		return nil, err
	}
	return &pbv1.DeletePermissionResponse{}, nil
}

func (h permissionHandler) GetPermission(ctx context.Context, request *pbv1.GetPermissionRequest) (*pbv1.GetPermissionResponse, error) {
	p, err := h.permissionUseCase.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &pbv1.GetPermissionResponse{Permission: permissionDomainToPb(p)}, nil
}

func (h permissionHandler) ListPermission(ctx context.Context, request *pbv1.ListPermissionRequest) (*pbv1.ListPermissionResponse, error) {
	items, total, err := h.permissionUseCase.List(ctx, &system.ListPermissionQueryParams{
		Limit:          request.Limit,
		Offset:         request.Offset,
		PermissionIds:  request.PermissionIds,
		PermissionName: request.PermissionName,
	})
	if err != nil {
		return nil, err
	}
	var pbItems []*pbv1.Permission
	for _, item := range items {
		pbItems = append(pbItems, permissionDomainToPb(item))
	}
	return &pbv1.ListPermissionResponse{
		Permissions: pbItems,
		Total:       total,
	}, nil
}

func (h permissionHandler) UpdatePermission(ctx context.Context, request *pbv1.UpdatePermissionRequest) (*pbv1.UpdatePermissionResponse, error) {
	p, err := h.permissionUseCase.Update(ctx, &permission.Permission{
		Id:          request.Id,
		Verb:        permission.Verb(request.Verb),
		Description: request.Description,
	})
	if err != nil {
		return nil, err
	}
	return &pbv1.UpdatePermissionResponse{Permission: permissionDomainToPb(p)}, nil
}
