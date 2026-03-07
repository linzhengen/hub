package seeds

import (
	_ "embed"

	"gopkg.in/yaml.v3"
)

//go:embed seed.yaml
var yamlSeed []byte

type Seed struct {
	Users []struct {
		Id     string `yaml:"id"`
		Name   string `yaml:"name"`
		Email  string `yaml:"email"`
		Status string `yaml:"status"`
	} `yaml:"users"`
	Groups []struct {
		Id          string `yaml:"id"`
		Name        string `yaml:"name"`
		Status      string `yaml:"status"`
		Description string `yaml:"description"`
	} `yaml:"groups"`
	UserGroups []struct {
		UserId  string `yaml:"user_id"`
		GroupId string `yaml:"group_id"`
	} `yaml:"user_groups"`
	Resources []struct {
		Id           string            `yaml:"id"`
		ParentId     string            `yaml:"parent_id"`
		Name         string            `yaml:"name"`
		Identifier   string            `yaml:"identifier"`
		Type         string            `yaml:"type"`
		Path         string            `yaml:"path"`
		Component    string            `yaml:"component"`
		DisplayOrder int32             `yaml:"display_order"`
		Description  string            `yaml:"description"`
		Metadata     map[string]string `yaml:"metadata"`
		Status       string            `yaml:"status"`
	} `yaml:"resources"`
	Permissions []struct {
		Id          string `yaml:"id"`
		Verb        string `yaml:"verb"`
		ResourceId  string `yaml:"resource_id"`
		Description string `yaml:"description"`
	} `yaml:"permissions"`
	Roles []struct {
		Id          string `yaml:"id"`
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Status      string `yaml:"status"`
	} `yaml:"roles"`
	RolePermissions []struct {
		RoleId       string `yaml:"role_id"`
		PermissionId string `yaml:"permission_id"`
	} `yaml:"role_permissions"`
	GroupRoles []struct {
		GroupId string `yaml:"group_id"`
		RoleId  string `yaml:"role_id"`
	} `yaml:"group_roles"`
}

func ParseSeed() (*Seed, error) {
	var seed Seed
	if err := yaml.Unmarshal(yamlSeed, &seed); err != nil {
		return nil, err
	}
	return &seed, nil
}
