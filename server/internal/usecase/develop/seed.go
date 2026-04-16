package develop

import (
	"context"
	"fmt"
	"time"

	"github.com/linzhengen/hub/v1/server/db/seeds"
	"github.com/linzhengen/hub/v1/server/internal/domain/system/group"
	"github.com/linzhengen/hub/v1/server/internal/domain/system/group/grouprole"
	"github.com/linzhengen/hub/v1/server/internal/domain/system/permission"
	"github.com/linzhengen/hub/v1/server/internal/domain/system/resource"
	"github.com/linzhengen/hub/v1/server/internal/domain/system/role"
	"github.com/linzhengen/hub/v1/server/internal/domain/system/role/rolepermission"
	"github.com/linzhengen/hub/v1/server/internal/domain/trans"
	"github.com/linzhengen/hub/v1/server/internal/domain/user"
	"github.com/linzhengen/hub/v1/server/internal/domain/user/usergroup"
	"github.com/linzhengen/hub/v1/server/pkg/logger"
)

type SeedUseCase interface {
	Seed(ctx context.Context) error
}

func NewSeedUseCase(
	trans trans.Repository,
	userRepo user.Repository,
	userGroupRepo usergroup.Repository,
	roleRepo role.Repository,
	permissionRepo permission.Repository,
	rolePermissionRepo rolepermission.Repository,
	resourceRepo resource.Repository,
	groupRepo group.Repository,
	groupRoleRepo grouprole.Repository,
) SeedUseCase {
	return &seedUseCase{
		trans:              trans,
		userRepo:           userRepo,
		userGroupRepo:      userGroupRepo,
		roleRepo:           roleRepo,
		permissionRepo:     permissionRepo,
		rolePermissionRepo: rolePermissionRepo,
		resourceRepo:       resourceRepo,
		groupRepo:          groupRepo,
		groupRoleRepo:      groupRoleRepo,
	}
}

type seedUseCase struct {
	trans              trans.Repository
	userRepo           user.Repository
	userGroupRepo      usergroup.Repository
	roleRepo           role.Repository
	permissionRepo     permission.Repository
	rolePermissionRepo rolepermission.Repository
	resourceRepo       resource.Repository
	groupRepo          group.Repository
	groupRoleRepo      grouprole.Repository
}

func (m seedUseCase) Seed(ctx context.Context) error {
	logger.Info("start seed")
	data, err := seeds.ParseSeed()
	if err != nil {
		return err
	}
	if err := m.trans.ExecTrans(ctx, func(ctx context.Context) error {
		if err := m.delete(ctx, data); err != nil {
			return fmt.Errorf("failed to delete seed data: %w", err)
		}
		if err := m.insert(ctx, data); err != nil {
			return fmt.Errorf("failed to insert seed data: %w", err)
		}
		return nil
	}); err != nil {
		logger.Errorf("failed seed, err: %s", err)
		return err
	}
	logger.Info("seed execute successfully")
	return nil
}

func (m seedUseCase) insert(ctx context.Context, data *seeds.Seed) error {
	for _, v := range data.Users {
		if err := m.userRepo.Create(ctx, &user.User{
			Id:        v.Id,
			Username:  v.Name,
			Email:     v.Email,
			Status:    user.Status(v.Status),
			UpdatedAt: time.Now(),
			CreatedAt: time.Now(),
		}); err != nil {
			return fmt.Errorf("failed seed user, err: %w", err)
		}
	}
	for _, v := range data.Roles {
		if err := m.roleRepo.Create(ctx, &role.Role{
			Id:          v.Id,
			Name:        v.Name,
			Description: v.Description,
			UpdatedAt:   time.Now(),
			CreatedAt:   time.Now(),
		}); err != nil {
			return fmt.Errorf("failed seed role, err: %w", err)
		}
	}
	for _, v := range data.Groups {
		if err := m.groupRepo.Create(ctx, &group.Group{
			Id:          v.Id,
			Name:        v.Name,
			Status:      group.Status(v.Status),
			Description: v.Description,
			UpdatedAt:   time.Now(),
			CreatedAt:   time.Now(),
		}); err != nil {
			return fmt.Errorf("failed seed group, err: %w", err)
		}
	}
	for _, v := range data.Resources {
		identifier := new(resource.Identifier)
		if err := identifier.Scan(v.Identifier); err != nil {
			return err
		}
		if err := m.resourceRepo.Create(ctx, &resource.Resource{
			Id:           v.Id,
			Name:         v.Name,
			Identifier:   *identifier,
			Type:         resource.ResourceType(v.Type),
			Path:         v.Path,
			Component:    v.Component,
			DisplayOrder: v.DisplayOrder,
			Description:  v.Description,
			Metadata:     v.Metadata,
			Status:       resource.Status(v.Status),
			UpdatedAt:    time.Now(),
			CreatedAt:    time.Now(),
		}); err != nil {
			return fmt.Errorf("failed seed resource, err: %w", err)
		}
	}
	for _, v := range data.Permissions {
		if err := m.permissionRepo.Create(ctx, &permission.Permission{
			Id:          v.Id,
			Verb:        permission.Verb(v.Verb),
			ResourceId:  v.ResourceId,
			Description: v.Description,
			UpdatedAt:   time.Now(),
			CreatedAt:   time.Now(),
		}); err != nil {
			return fmt.Errorf("failed seed permission, err: %w", err)
		}
	}
	for _, v := range data.UserGroups {
		if err := m.userGroupRepo.AssignGroup(ctx, v.UserId, v.GroupId); err != nil {
			return fmt.Errorf("failed seed user group, err: %w", err)
		}
	}
	for _, v := range data.RolePermissions {
		if err := m.rolePermissionRepo.AssignPermission(ctx, v.RoleId, v.PermissionId); err != nil {
			return fmt.Errorf("failed seed role permission, err: %w", err)
		}
	}
	for _, v := range data.GroupRoles {
		if err := m.groupRoleRepo.AssignRole(ctx, v.GroupId, v.RoleId); err != nil {
			return fmt.Errorf("failed seed group role, err: %w", err)
		}
	}
	return nil
}

func (m seedUseCase) delete(ctx context.Context, data *seeds.Seed) error {
	for _, v := range data.Users {
		if err := m.userRepo.Delete(ctx, v.Id); err != nil {
			return fmt.Errorf("failed to delete user %s: %w", v.Id, err)
		}
	}
	for _, v := range data.Roles {
		if err := m.roleRepo.Delete(ctx, v.Id); err != nil {
			return fmt.Errorf("failed to delete role %s: %w", v.Id, err)
		}
	}
	for _, v := range data.Groups {
		if err := m.groupRepo.Delete(ctx, v.Id); err != nil {
			return fmt.Errorf("failed to delete group %s: %w", v.Id, err)
		}
	}
	for _, v := range data.Permissions {
		if err := m.permissionRepo.Delete(ctx, v.Id); err != nil {
			return fmt.Errorf("failed to delete permission %s: %w", v.Id, err)
		}
	}
	for _, v := range data.Resources {
		if err := m.resourceRepo.Delete(ctx, v.Id); err != nil {
			return fmt.Errorf("failed to delete resource %s: %w", v.Id, err)
		}
	}
	return nil
}
