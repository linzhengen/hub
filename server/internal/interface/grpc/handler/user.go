package handler

import (
	"context"

	"github.com/linzhengen/hub/server/internal/domain/system/resource/menu"
	"github.com/linzhengen/hub/server/internal/usecase"
	"github.com/linzhengen/hub/server/internal/usecase/system"
	pbgroupv1 "github.com/linzhengen/hub/server/pb/system/group/v1"
	resourcev1 "github.com/linzhengen/hub/server/pb/system/resource/v1"
	pbv1 "github.com/linzhengen/hub/server/pb/user/v1"

	"github.com/linzhengen/hub/server/internal/domain/user"
)

func NewUserHandler(
	userUseCase usecase.UserUseCase,
	groupUseCase system.GroupUseCase,
) pbv1.UserServiceServer {
	return &userHandler{
		userUseCase:  userUseCase,
		groupUseCase: groupUseCase,
	}
}

type userHandler struct {
	userUseCase  usecase.UserUseCase
	groupUseCase system.GroupUseCase
}

func (h userHandler) GetMe(ctx context.Context, _ *pbv1.GetMeRequest) (*pbv1.GetMeResponse, error) {
	u, err := h.userUseCase.Me(ctx)
	if err != nil {
		return nil, err
	}
	gs, _, err := h.groupUseCase.List(ctx, &system.ListGroupQueryParams{
		GroupIds: u.GroupIds,
	})
	if err != nil {
		return nil, err
	}
	var gsPb []*pbgroupv1.Group
	for _, v := range gs {
		gsPb = append(gsPb, groupDomainToPb(v))
	}
	return &pbv1.GetMeResponse{
		User:   userDomainToPb(u),
		Groups: gsPb,
	}, nil
}

func (h userHandler) GetMeMenus(ctx context.Context, _ *pbv1.GetMeMenusRequest) (*pbv1.GetMeMenusResponse, error) {
	menus, err := h.userUseCase.GetMeMenus(ctx)
	if err != nil {
		return nil, err
	}

	return &pbv1.GetMeMenusResponse{
		Menus: menusDomainToPb(buildMenuTree(menus)),
	}, nil
}

func buildMenuTree(menus []*menu.Menu) []*menu.Menu {
	menuMap := make(map[string]*menu.Menu)
	for _, menu := range menus {
		menuMap[menu.Id] = menu
	}

	var tree []*menu.Menu
	for _, menu := range menus {
		if menu.ParentId == "" {
			tree = append(tree, menu)
		} else {
			if parent, ok := menuMap[menu.ParentId]; ok {
				parent.Children = append(parent.Children, menu)
			}
		}
	}

	return tree
}

func (h userHandler) UpdateMe(ctx context.Context, request *pbv1.UpdateMeRequest) (*pbv1.UpdateMeResponse, error) {
	// Get the current user
	currentUser, err := h.userUseCase.Me(ctx)
	if err != nil {
		return nil, err
	}

	// Update user information
	currentUser.Username = request.Username
	currentUser.Email = request.Email

	// Call the Update method to update the user in DB and Keycloak
	var password *string
	if request.Password != "" {
		password = &request.Password
	}

	updatedUser, err := h.userUseCase.Update(ctx, currentUser, password)
	if err != nil {
		return nil, err
	}

	// Get the user's groups
	gs, _, err := h.groupUseCase.List(ctx, &system.ListGroupQueryParams{
		GroupIds: updatedUser.GroupIds,
	})
	if err != nil {
		return nil, err
	}

	// Convert groups to protobuf format
	var gsPb []*pbgroupv1.Group
	for _, v := range gs {
		gsPb = append(gsPb, groupDomainToPb(v))
	}

	// Return the updated user and groups
	return &pbv1.UpdateMeResponse{
		User:   userDomainToPb(updatedUser),
		Groups: gsPb,
	}, nil
}

func (h userHandler) GetUser(ctx context.Context, request *pbv1.GetUserRequest) (*pbv1.GetUserResponse, error) {
	u, err := h.userUseCase.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &pbv1.GetUserResponse{User: userDomainToPb(u)}, nil
}

func (h userHandler) ListUser(ctx context.Context, request *pbv1.ListUserRequest) (*pbv1.ListUserResponse, error) {
	status, _ := toUserDomainStatus(request.Status)
	params := &usecase.ListUserQueryParams{
		Limit:      request.Limit,
		Offset:     request.Offset,
		UserIds:    request.UserIds,
		UserEmails: request.UserEmails,
		UserName:   request.UserName,
		Status:     status,
	}

	// Only add group_ids filter if it's not empty
	if len(request.GroupIds) > 0 {
		params.GroupIds = request.GroupIds
	}

	items, total, err := h.userUseCase.List(ctx, params)
	if err != nil {
		return nil, err
	}
	var pbItems []*pbv1.User
	for _, v := range items {
		pbItems = append(pbItems, userDomainToPb(v))
	}
	return &pbv1.ListUserResponse{Users: pbItems, Total: total}, nil
}

func (h userHandler) UpdateUser(ctx context.Context, request *pbv1.UpdateUserRequest) (*pbv1.UpdateUserResponse, error) {
	status, err := toUserDomainStatus(request.Status)
	if err != nil {
		return nil, err
	}
	u, err := h.userUseCase.Update(ctx, &user.User{
		Id:       request.Id,
		Username: request.Username,
		Email:    request.Email,
		Status:   status,
		GroupIds: request.GroupIds,
	}, request.Password)
	if err != nil {
		return nil, err
	}
	return &pbv1.UpdateUserResponse{User: userDomainToPb(u)}, nil
}

func (h userHandler) DeleteUser(ctx context.Context, request *pbv1.DeleteUserRequest) (*pbv1.DeleteUserResponse, error) {
	if err := h.userUseCase.Delete(ctx, request.Id); err != nil {
		return nil, err
	}
	return &pbv1.DeleteUserResponse{}, nil
}

func (h userHandler) AssignGroup(ctx context.Context, request *pbv1.AssignGroupRequest) (*pbv1.AssignGroupResponse, error) {
	u, err := h.userUseCase.AssignGroup(ctx, request.Id, request.GroupId)
	if err != nil {
		return nil, err
	}
	return &pbv1.AssignGroupResponse{User: userDomainToPb(u)}, nil
}

func (h userHandler) UnassignGroup(ctx context.Context, request *pbv1.UnassignGroupRequest) (*pbv1.UnassignGroupResponse, error) {
	u, err := h.userUseCase.UnassignGroup(ctx, request.Id, request.GroupId)
	if err != nil {
		return nil, err
	}
	return &pbv1.UnassignGroupResponse{User: userDomainToPb(u)}, nil
}

func (h userHandler) CreateUser(ctx context.Context, request *pbv1.CreateUserRequest) (*pbv1.CreateUserResponse, error) {
	u, err := h.userUseCase.Create(ctx, request.Username, request.Email, request.Password, request.GroupIds)
	if err != nil {
		return nil, err
	}
	return &pbv1.CreateUserResponse{User: userDomainToPb(u)}, nil
}

func menusDomainToPb(menus []*menu.Menu) []*pbv1.Menu {
	var pbMenus []*pbv1.Menu
	for _, menu := range menus {
		pbMenus = append(pbMenus, menuDomainToPb(menu))
	}
	return pbMenus
}

func menuDomainToPb(menu *menu.Menu) *pbv1.Menu {
	meta := &pbv1.MenuMeta{}
	if menu.Metadata != nil {
		if icon, ok := menu.Metadata["icon"].(string); ok {
			meta.Icon = icon
		}
		if keepAlive, ok := menu.Metadata["keepAlive"].(string); ok {
			meta.KeepAlive = keepAlive == "true"
		}
		if order, ok := menu.Metadata["order"].(float64); ok {
			meta.Order = uint32(order)
		}
		if title, ok := menu.Metadata["title"].(string); ok {
			meta.Title = title
		}
		if authority, ok := menu.Metadata["authority"].(string); ok {
			meta.Authority = authority
		}
		if badge, ok := menu.Metadata["badge"].(string); ok {
			meta.Badge = badge
		}
		if hideInMenu, ok := menu.Metadata["hideInMenu"].(string); ok {
			meta.HideInMenu = hideInMenu == "true"
		}
	}
	return &pbv1.Menu{
		Name:       menu.Name,
		Identifier: menu.Identifier,
		Path:       menu.Path,
		Component:  menu.Component,
		Type:       resourcev1.Type(resourcev1.Type_value[string(menu.Type)]),
		Meta:       meta,
		Children:   menusDomainToPb(menu.Children),
	}
}
