package system

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/linzhengen/hub/server/internal/domain/system/permission"
	"github.com/linzhengen/hub/server/internal/domain/system/resource"
	"github.com/linzhengen/hub/server/internal/domain/system/resource/api"
	"github.com/linzhengen/hub/server/internal/domain/system/resource/menu"
)

type ResourceService interface {
	CreateResourcesAndPermissionsFromAPI(ctx context.Context) error
	CreateResourcesAndPermissionsFromMenu(ctx context.Context) error
}

func NewResourceService(apiRepo api.Repository, menuRepo menu.Repository, resourceRepo resource.Repository, permissionRepo permission.Repository) ResourceService {
	return &resourceService{
		apiRepo:        apiRepo,
		menuRepo:       menuRepo,
		resourceRepo:   resourceRepo,
		permissionRepo: permissionRepo,
	}
}

type resourceService struct {
	apiRepo        api.Repository
	menuRepo       menu.Repository
	resourceRepo   resource.Repository
	permissionRepo permission.Repository
}

// convertMetadata converts map[string]interface{} to map[string]string
func convertMetadata(metadata map[string]interface{}) resource.Metadata {
	result := make(resource.Metadata)
	for k, v := range metadata {
		if str, ok := v.(string); ok {
			result[k] = str
		} else {
			// Convert non-string values to string representation
			result[k] = fmt.Sprintf("%v", v)
		}
	}
	return result
}

func (r resourceService) CreateResourcesAndPermissionsFromMenu(ctx context.Context) error {
	// Load menu definitions from menu.Repository
	menus, err := r.menuRepo.FindAll(ctx)
	if err != nil {
		return err
	}

	// Process each menu
	for _, menu := range menus {
		// Create resource with the identifier
		res := resource.Factory(
			menu.Name,
			menu.ParentId,
			resource.Identifier{
				Api:      "menu",
				Category: menu.Path,
			},
			resource.ResourceTypeMenu,
			menu.Path,
			menu.Component,
			int32(menu.DisplayOrder),
			menu.Description,
			convertMetadata(menu.Metadata),
		)
		res.Id = menu.Id
		if rs, err := r.resourceRepo.FindOneByIdentifier(ctx, res.Identifier.String()); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				if err := r.resourceRepo.Create(ctx, res); err != nil {
					return err
				}
			}
		} else {
			res.Id = rs.Id
			if err := r.resourceRepo.Update(ctx, res); err != nil {
				return err
			}
		}

		// Create a "view" permission for the menu
		perms, err := r.permissionRepo.FindByResourceId(ctx, res.Id)
		if err != nil {
			return err
		}

		// Create permission with the "view" verb
		perm, err := permission.Factory(
			"view",
			res.Id,
			"Auto-generated permission for "+menu.Identifier+" with verb view",
		)
		if err != nil {
			return err
		}

		found := false
		for _, v := range perms {
			if v.ResourceId == res.Id && v.Verb == perm.Verb {
				found = true
				break
			}
		}
		if found {
			continue
		}

		// Try to create the permission
		if err := r.permissionRepo.Create(ctx, perm); err != nil {
			return err
		}
	}

	return nil
}

func (r resourceService) CreateResourcesAndPermissionsFromAPI(ctx context.Context) error {
	// Load Proto definitions from api.Repository
	apis, err := r.apiRepo.FindAll(ctx)
	if err != nil {
		return err
	}

	// Group APIs by service and collect methods
	serviceToMethods := make(map[string][]string)
	for _, a := range apis {
		if _, exists := serviceToMethods[a.Service]; !exists {
			serviceToMethods[a.Service] = []string{}
		}
		serviceToMethods[a.Service] = append(serviceToMethods[a.Service], a.Method)
	}

	// Process each service
	for service, methods := range serviceToMethods {
		// Create resource with the identifier
		res := resource.Factory(
			service,
			"", // No parent
			resource.Identifier{
				Api:      "api",
				Category: service,
			},
			resource.ResourceTypeApi,
			"", // No path
			"", // No component
			0,  // No display order
			"Auto-generated resource for "+service,
			map[string]string{}, // No metadata
		)

		if rs, err := r.resourceRepo.FindOneByIdentifier(ctx, res.Identifier.String()); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				if err := r.resourceRepo.Create(ctx, res); err != nil {
					return err
				}
			}
		} else {
			res.Id = rs.Id
			if err := r.resourceRepo.Update(ctx, res); err != nil {
				return err
			}
		}
		perms, err := r.permissionRepo.FindByResourceId(ctx, res.Id)
		if err != nil {
			return err
		}
		for _, verb := range methods {
			// Create permission with the verb
			perm, err := permission.Factory(
				verb,
				res.Id,
				"Auto-generated permission for "+service+" with verb "+verb,
			)
			if err != nil {
				return err
			}
			found := false
			for _, v := range perms {
				if v.ResourceId == res.Id && v.Verb == perm.Verb {
					found = true
					break
				}
			}
			if found {
				continue
			}
			// Try to create the permission
			if err := r.permissionRepo.Create(ctx, perm); err != nil {
				return err
			}
		}
	}

	return nil
}
