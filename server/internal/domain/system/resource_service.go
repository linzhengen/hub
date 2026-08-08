package system

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"

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
			// Anything other than "not found" means we do not know whether the
			// resource exists. Carrying on would create permissions against a
			// resource that was never written.
			if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			if err := r.resourceRepo.Create(ctx, res); err != nil {
				return err
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

// apiVerb is one permission to register under an API resource.
type apiVerb struct {
	verb        string
	description string
}

// apiResource collects the verbs that share a resource identifier.
type apiResource struct {
	identifier resource.Identifier
	name       string
	verbs      []apiVerb
}

func (r resourceService) CreateResourcesAndPermissionsFromAPI(ctx context.Context) error {
	// Load Proto definitions from api.Repository
	apis, err := r.apiRepo.FindAll(ctx)
	if err != nil {
		return err
	}

	grouped, order, err := groupAPIsByResource(apis)
	if err != nil {
		return err
	}

	// Iterate in discovery order rather than over the map, so a run that fails
	// part way through is reproducible.
	for _, identifier := range order {
		if err := r.importAPIResource(ctx, grouped[identifier]); err != nil {
			return err
		}
	}

	return nil
}

// groupAPIsByResource buckets rpcs under the resource they are guarded by.
//
// Public rpcs are skipped: authorization never consults a permission for them,
// so registering one would only add a record that looks meaningful in the RBAC
// screens and grants nothing.
func groupAPIsByResource(apis api.APIs) (map[string]*apiResource, []string, error) {
	grouped := map[string]*apiResource{}
	var order []string

	for _, a := range apis {
		if a.Public {
			continue
		}

		entry, ok := grouped[a.Resource]
		if !ok {
			identifier := resource.Identifier{}
			if err := identifier.Scan(a.Resource); err != nil {
				return nil, nil, fmt.Errorf("api %s.%s has an unusable resource identifier %q: %w",
					a.Service, a.Method, a.Resource, err)
			}
			entry = &apiResource{identifier: identifier, name: a.Service}
			grouped[a.Resource] = entry
			order = append(order, a.Resource)
		}
		entry.verbs = append(entry.verbs, apiVerb{
			verb:        a.Action,
			description: permissionDescription(a),
		})
	}

	return grouped, order, nil
}

func permissionDescription(a *api.API) string {
	if a.Summary != "" {
		return a.Summary
	}
	return "Auto-generated permission for " + a.Service + " with verb " + a.Action
}

func (r resourceService) importAPIResource(ctx context.Context, entry *apiResource) error {
	res := resource.Factory(
		entry.name,
		"", // No parent
		entry.identifier,
		resource.ResourceTypeApi,
		"", // No path
		"", // No component
		0,  // No display order
		"Auto-generated resource for "+entry.name,
		map[string]string{}, // No metadata
	)

	if rs, err := r.resourceRepo.FindOneByIdentifier(ctx, res.Identifier.String()); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if err := r.resourceRepo.Create(ctx, res); err != nil {
			return err
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

	for _, v := range entry.verbs {
		perm, err := permission.Factory(v.verb, res.Id, v.description)
		if err != nil {
			return err
		}
		if slices.ContainsFunc(perms, func(existing *permission.Permission) bool {
			return existing.ResourceId == res.Id && existing.Verb == perm.Verb
		}) {
			continue
		}
		if err := r.permissionRepo.Create(ctx, perm); err != nil {
			return err
		}
	}

	return nil
}
