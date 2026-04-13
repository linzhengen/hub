package menu

import (
	_ "embed"

	"gopkg.in/yaml.v3"
)

//go:embed menus.yaml
var menuBytes []byte

type Menus struct {
	Menus []*Menu `yaml:"menus"`
}

type Menu struct {
	ID           string                 `yaml:"id"`
	Name         string                 `yaml:"name"`
	Type         string                 `yaml:"type"`
	Path         string                 `yaml:"path"`
	Component    string                 `yaml:"component"`
	DisplayOrder int                    `yaml:"displayOrder"`
	Description  string                 `yaml:"description,omitempty"`
	Metadata     map[string]interface{} `yaml:"metadata,omitempty"`
	Status       string                 `yaml:"status"`
	Children     []*Menu                `yaml:"children,omitempty"`
}

func SelectAllMenus() Menus {
	var menus Menus
	err := yaml.Unmarshal(menuBytes, &menus)
	if err != nil {
		return Menus{}
	}
	return menus
}
