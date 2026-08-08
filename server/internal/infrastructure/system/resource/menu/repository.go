package menu

import (
	"context"
	"time"

	"github.com/linzhengen/hub/server/internal/domain/system/resource/menu"
	yamlMenu "github.com/linzhengen/hub/server/internal/infrastructure/persistence/yaml/menu"
)

func New() menu.Repository {
	return &repositoryImpl{}
}

type repositoryImpl struct {
}

// convertYamlMenuToDomainMenu converts a YAML menu to a domain menu
func convertYamlMenuToDomainMenu(yamlMenu *yamlMenu.Menu, parentId string) menu.Menu {
	domainMenu := menu.Menu{
		Id:           yamlMenu.ID,
		ParentId:     parentId,
		Name:         yamlMenu.Name,
		Type:         yamlMenu.Type,
		Path:         yamlMenu.Path,
		Component:    yamlMenu.Component,
		DisplayOrder: yamlMenu.DisplayOrder,
		Description:  yamlMenu.Description,
		Metadata:     yamlMenu.Metadata,
		Status:       yamlMenu.Status,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Children:     make([]*menu.Menu, 0),
	}

	// Convert children if any
	if yamlMenu.Children != nil {
		for _, child := range yamlMenu.Children {
			childMenu := convertYamlMenuToDomainMenu(child, yamlMenu.ID)
			domainMenu.Children = append(domainMenu.Children, &childMenu)
		}
	}

	return domainMenu
}

func (r repositoryImpl) FindAll(ctx context.Context) (menu.Menus, error) {
	// Get menus from YAML
	yamlMenus := yamlMenu.SelectAllMenus()

	// Convert to domain menus
	domainMenus := make(menu.Menus, 0)

	// Process top-level menus
	for _, m := range yamlMenus.Menus {
		domainMenu := convertYamlMenuToDomainMenu(m, "")
		domainMenus = append(domainMenus, domainMenu)

		// Process children recursively and add them to the flat list
		var processChildren func([]*menu.Menu)
		processChildren = func(children []*menu.Menu) {
			for _, child := range children {
				domainMenus = append(domainMenus, *child)
				if len(child.Children) > 0 {
					processChildren(child.Children)
				}
			}
		}

		if len(domainMenu.Children) > 0 {
			processChildren(domainMenu.Children)
		}
	}

	return domainMenus, nil
}
