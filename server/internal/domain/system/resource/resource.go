package resource

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/linzhengen/hub/server/pkg/uuid"
)

type Status string

const (
	Active   Status = "Active"
	InActive Status = "Inactive"
)

type ResourceType string

const (
	ResourceTypeMenu ResourceType = "menu"
	ResourceTypeApi  ResourceType = "api"
)

const IdentifierFormat = "%s.%s"

type Identifier struct {
	Api      string
	Category string
}

func (i *Identifier) String() string {
	return fmt.Sprintf(IdentifierFormat, i.Api, i.Category)
}

func (i *Identifier) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		index := strings.Index(v, ".")
		if index < 0 {
			return fmt.Errorf("invalid format for Identifier")
		}
		i.Api = v[:index]
		i.Category = v[index+1:]
		return nil
	default:
		return fmt.Errorf("unsupported type for Identifier: %T", value)
	}
}

type Metadata map[string]string

type Resource struct {
	Id           string
	ParentId     string
	Name         string
	Identifier   Identifier
	Type         ResourceType
	Path         string
	Component    string
	DisplayOrder int32
	Description  string
	Metadata     Metadata
	Status       Status
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (v Metadata) JsonRawMessage() (json.RawMessage, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func Factory(
	name string,
	parentId string,
	identifier Identifier,
	resourceType ResourceType,
	path string,
	component string,
	displayOrder int32,
	description string,
	metadata Metadata,
) *Resource {
	return &Resource{
		Id:           uuid.MustUUID().String(),
		ParentId:     parentId,
		Name:         name,
		Identifier:   identifier,
		Type:         resourceType,
		Path:         path,
		Component:    component,
		DisplayOrder: displayOrder,
		Description:  description,
		Metadata:     metadata,
		Status:       Active,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func (r *Resource) ValidateForMenu() error {
	if r.Name == "" {
		return fmt.Errorf("name is required, error: %w", ErrInvalidRequest)
	}
	if r.Path == "" {
		return fmt.Errorf("path is required, error: %w", ErrInvalidRequest)
	}
	if r.Component == "" {
		return fmt.Errorf("component is required, error: %w", ErrInvalidRequest)
	}
	return nil
}
