package system

import (
	"context"

	"github.com/linzhengen/hub/v1/server/internal/usecase/system"

	"github.com/linzhengen/hub/v1/server/internal/domain/system/group"
	pbv1 "github.com/linzhengen/hub/v1/server/pb/system/group/v1"
)

func NewGroupHandler(
	groupUseCase system.GroupUseCase,
) pbv1.GroupServiceServer {
	return &groupHandler{groupUseCase: groupUseCase}
}

type groupHandler struct {
	groupUseCase system.GroupUseCase
}

func (h groupHandler) CreateGroup(ctx context.Context, request *pbv1.CreateGroupRequest) (*pbv1.CreateGroupResponse, error) {
	g, err := h.groupUseCase.Create(ctx, group.Factory(
		request.Name,
		request.Description,
	))
	if err != nil {
		return nil, err
	}
	return &pbv1.CreateGroupResponse{Group: groupDomainToPb(g)}, nil
}

func (h groupHandler) DeleteGroup(ctx context.Context, request *pbv1.DeleteGroupRequest) (*pbv1.DeleteGroupResponse, error) {
	if err := h.groupUseCase.Delete(ctx, request.Id); err != nil {
		return nil, err
	}
	return &pbv1.DeleteGroupResponse{}, nil
}

func (h groupHandler) GetGroup(ctx context.Context, request *pbv1.GetGroupRequest) (*pbv1.GetGroupResponse, error) {
	g, err := h.groupUseCase.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &pbv1.GetGroupResponse{Group: groupDomainToPb(g)}, nil
}

func (h groupHandler) ListGroup(ctx context.Context, request *pbv1.ListGroupRequest) (*pbv1.ListGroupResponse, error) {
	status, _ := toGroupDomainStatus(request.Status)
	items, total, err := h.groupUseCase.List(ctx, &system.ListGroupQueryParams{
		Limit:     request.Limit,
		Offset:    request.Offset,
		GroupIds:  request.GroupIds,
		GroupName: request.GroupName,
		Status:    status,
		RoleIds:   request.RoleIds,
	})
	if err != nil {
		return nil, err
	}
	var pbItems []*pbv1.Group
	for _, item := range items {
		pbItems = append(pbItems, groupDomainToPb(item))
	}
	return &pbv1.ListGroupResponse{
		Groups: pbItems,
		Total:  total,
	}, nil
}

func (h groupHandler) UpdateGroup(ctx context.Context, request *pbv1.UpdateGroupRequest) (*pbv1.UpdateGroupResponse, error) {
	status, err := toGroupDomainStatus(request.Status)
	if err != nil {
		return nil, err
	}
	g, err := h.groupUseCase.Update(ctx, &group.Group{
		Id:          request.Id,
		Name:        request.Name,
		Description: request.Description,
		Status:      status,
	})
	if err != nil {
		return nil, err
	}
	return &pbv1.UpdateGroupResponse{Group: groupDomainToPb(g)}, nil
}

func (h groupHandler) AssignRole(ctx context.Context, request *pbv1.AssignRoleRequest) (*pbv1.AssignRoleResponse, error) {
	g, err := h.groupUseCase.AssignRole(ctx, request.Id, request.RoleId)
	if err != nil {
		return nil, err
	}
	return &pbv1.AssignRoleResponse{Group: groupDomainToPb(g)}, nil
}

func (h groupHandler) AddUsersToGroup(ctx context.Context, request *pbv1.AddUsersToGroupRequest) (*pbv1.AddUsersToGroupResponse, error) {
	_, err := h.groupUseCase.AddUsersToGroup(ctx, request.GroupId, request.UserIds)
	if err != nil {
		return nil, err
	}
	return &pbv1.AddUsersToGroupResponse{}, nil
}

func (h groupHandler) RemoveUsersFromGroup(ctx context.Context, request *pbv1.RemoveUsersFromGroupRequest) (*pbv1.RemoveUsersFromGroupResponse, error) {
	_, err := h.groupUseCase.RemoveUsersFromGroup(ctx, request.GroupId, request.UserIds)
	if err != nil {
		return nil, err
	}
	return &pbv1.RemoveUsersFromGroupResponse{}, nil
}

func (h groupHandler) AssignRolesToGroup(ctx context.Context, request *pbv1.AssignRolesToGroupRequest) (*pbv1.AssignRolesToGroupResponse, error) {
	_, err := h.groupUseCase.AssignRolesToGroup(ctx, request.GroupId, request.RoleIds)
	if err != nil {
		return nil, err
	}
	return &pbv1.AssignRolesToGroupResponse{}, nil
}
