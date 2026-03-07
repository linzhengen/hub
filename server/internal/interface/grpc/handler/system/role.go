package system

import (
	"context"

	"github.com/linzhengen/hub/server/internal/domain/system/role"
	"github.com/linzhengen/hub/server/internal/usecase/system"
	pbv1 "github.com/linzhengen/hub/server/pb/system/role/v1"
)

func NewRoleHandler(
	roleUseCase system.RoleUseCase,
	permissionUseCase system.PermissionUseCase,
) pbv1.RoleServiceServer {
	return &roleHandler{
		roleUseCase:       roleUseCase,
		permissionUseCase: permissionUseCase,
	}
}

type roleHandler struct {
	roleUseCase       system.RoleUseCase
	permissionUseCase system.PermissionUseCase
}

func (h roleHandler) CreateRole(ctx context.Context, request *pbv1.CreateRoleRequest) (*pbv1.CreateRoleResponse, error) {
	r, err := h.roleUseCase.Create(ctx, role.Factory(
		request.Name,
		request.Description,
	))
	if err != nil {
		return nil, err
	}
	return &pbv1.CreateRoleResponse{Role: roleDomainToPb(r)}, nil
}

func (h roleHandler) DeleteRole(ctx context.Context, request *pbv1.DeleteRoleRequest) (*pbv1.DeleteRoleResponse, error) {
	if err := h.roleUseCase.Delete(ctx, request.Id); err != nil {
		return nil, err
	}
	return &pbv1.DeleteRoleResponse{}, nil
}

func (h roleHandler) GetRole(ctx context.Context, request *pbv1.GetRoleRequest) (*pbv1.GetRoleResponse, error) {
	r, err := h.roleUseCase.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &pbv1.GetRoleResponse{Role: roleDomainToPb(r)}, nil
}

func (h roleHandler) ListRole(ctx context.Context, request *pbv1.ListRoleRequest) (*pbv1.ListRoleResponse, error) {
	params := &system.ListRoleQueryParams{
		Limit:    request.Limit,
		Offset:   request.Offset,
		RoleIds:  request.RoleIds,
		RoleName: request.RoleName,
	}

	// Only add permission_ids filter if it's not empty
	if len(request.PermissionIds) > 0 {
		params.PermissionIds = request.PermissionIds
	}

	items, total, err := h.roleUseCase.List(ctx, params)
	if err != nil {
		return nil, err
	}
	var pbItems []*pbv1.Role
	for _, v := range items {
		pbItems = append(pbItems, roleDomainToPb(v))
	}
	return &pbv1.ListRoleResponse{Roles: pbItems, Total: total}, nil
}

func (h roleHandler) UpdateRole(ctx context.Context, request *pbv1.UpdateRoleRequest) (*pbv1.UpdateRoleResponse, error) {
	r, err := h.roleUseCase.Update(ctx, &role.Role{
		Id:          request.Id,
		Name:        request.Name,
		Description: request.Description,
	})
	if err != nil {
		return nil, err
	}
	return &pbv1.UpdateRoleResponse{Role: roleDomainToPb(r)}, nil
}

func (h roleHandler) AssignPermission(ctx context.Context, request *pbv1.AssignPermissionRequest) (*pbv1.AssignPermissionResponse, error) {
	r, err := h.roleUseCase.AssignPermission(ctx, request.Id, request.PermissionId)
	if err != nil {
		return nil, err
	}
	return &pbv1.AssignPermissionResponse{Role: roleDomainToPb(r)}, nil
}

func (h roleHandler) AddPermissionsToRole(ctx context.Context, request *pbv1.AddPermissionsToRoleRequest) (*pbv1.AddPermissionsToRoleResponse, error) {
	_, err := h.roleUseCase.AddPermissionsToRole(ctx, request.RoleId, request.PermissionIds)
	if err != nil {
		return nil, err
	}
	return &pbv1.AddPermissionsToRoleResponse{}, nil
}

func (h roleHandler) RemovePermissionsFromRole(ctx context.Context, request *pbv1.RemovePermissionsFromRoleRequest) (*pbv1.RemovePermissionsFromRoleResponse, error) {
	_, err := h.roleUseCase.RemovePermissionsFromRole(ctx, request.RoleId, request.PermissionIds)
	if err != nil {
		return nil, err
	}
	return &pbv1.RemovePermissionsFromRoleResponse{}, nil
}
