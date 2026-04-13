package menu

import "time"

type Menu struct {
	Id           string
	ParentId     string
	Name         string
	Identifier   string
	Type         string
	Path         string
	Component    string
	DisplayOrder int
	Description  string
	Metadata     map[string]interface{}
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Children     []*Menu
}

type Menus []Menu
