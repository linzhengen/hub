package system

import (
	"fmt"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/linzhengen/hub/v1/server/internal/domain/system/group"
	"github.com/linzhengen/hub/v1/server/internal/domain/system/permission"
	"github.com/linzhengen/hub/v1/server/internal/domain/system/resource"
	"github.com/linzhengen/hub/v1/server/internal/domain/system/role"
	pbgroupv1 "github.com/linzhengen/hub/v1/server/pb/system/group/v1"
	pbpermissionv1 "github.com/linzhengen/hub/v1/server/pb/system/permission/v1"
	pbresourcev1 "github.com/linzhengen/hub/v1/server/pb/system/resource/v1"
	pbrolev1 "github.com/linzhengen/hub/v1/server/pb/system/role/v1"
)

func toGroupPbStatus(s group.Status) pbgroupv1.Group_Status {
	switch s {
	case group.Active:
		return pbgroupv1.Group_STATUS_ACTIVE
	case group.InActive:
		return pbgroupv1.Group_STATUS_INACTIVE
	default:
		return pbgroupv1.Group_STATUS_UNSPECIFIED
	}
}

func toGroupDomainStatus(status pbgroupv1.Group_Status) (group.Status, error) {
	switch status {
	case pbgroupv1.Group_STATUS_ACTIVE:
		return group.Active, nil
	case pbgroupv1.Group_STATUS_INACTIVE:
		return group.InActive, nil
	default:
		return "", fmt.Errorf("unknown group status: %s", status.String())
	}
}

func groupDomainToPb(m *group.Group) *pbgroupv1.Group {
	return &pbgroupv1.Group{
		Id:          m.Id,
		Name:        m.Name,
		Status:      toGroupPbStatus(m.Status),
		Description: m.Description,
		CreatedAt:   timestamppb.New(m.CreatedAt),
		UpdatedAt:   timestamppb.New(m.UpdatedAt),

		RoleIds: m.RoleIds,
	}
}

func roleDomainToPb(m *role.Role) *pbrolev1.Role {
	return &pbrolev1.Role{
		Id:            m.Id,
		Name:          m.Name,
		Description:   m.Description,
		CreatedAt:     timestamppb.New(m.CreatedAt),
		UpdatedAt:     timestamppb.New(m.UpdatedAt),
		PermissionIds: m.PermissionIds,
	}
}

func permissionDomainToPb(m *permission.Permission) *pbpermissionv1.Permission {
	return &pbpermissionv1.Permission{
		Id:          m.Id,
		Verb:        string(m.Verb),
		ResourceId:  m.ResourceId,
		Description: m.Description,
		CreatedAt:   timestamppb.New(m.CreatedAt),
		UpdatedAt:   timestamppb.New(m.UpdatedAt),
	}
}

func nullableStr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func toResourcePbStatus(s resource.Status) pbresourcev1.Status {
	switch s {
	case resource.Active:
		return pbresourcev1.Status_STATUS_ACTIVE
	case resource.InActive:
		return pbresourcev1.Status_STATUS_INACTIVE
	default:
		return pbresourcev1.Status_STATUS_UNSPECIFIED
	}
}

func toResourcePbType(t resource.ResourceType) pbresourcev1.Type {
	switch t {
	case resource.ResourceTypeApi:
		return pbresourcev1.Type_TYPE_API
	case resource.ResourceTypeMenu:
		return pbresourcev1.Type_TYPE_MENU
	default:
		return pbresourcev1.Type_TYPE_UNSPECIFIED
	}
}

func resourceDomainToPb(m *resource.Resource) *pbresourcev1.Resource {
	var pbIdentifier *pbresourcev1.Identifier
	if m.Identifier.Api != "" || m.Identifier.Category != "" {
		pbIdentifier = &pbresourcev1.Identifier{
			Api:      m.Identifier.Api,
			Category: m.Identifier.Category,
		}
	}
	return &pbresourcev1.Resource{
		Id:           m.Id,
		ParentId:     m.ParentId,
		Name:         m.Name,
		Identifier:   pbIdentifier,
		Type:         toResourcePbType(m.Type),
		Path:         m.Path,
		Component:    nullableStr(m.Component),
		DisplayOrder: m.DisplayOrder,
		Description:  nullableStr(m.Description),
		Metadata:     m.Metadata,
		Status:       toResourcePbStatus(m.Status),
		CreatedAt:    timestamppb.New(m.CreatedAt),
		UpdatedAt:    timestamppb.New(m.UpdatedAt),
	}
}

func toResourceDomainStatus(status pbresourcev1.Status) (resource.Status, error) {
	switch status {
	case pbresourcev1.Status_STATUS_ACTIVE:
		return resource.Active, nil
	case pbresourcev1.Status_STATUS_INACTIVE:
		return resource.InActive, nil
	default:
		return "", fmt.Errorf("unknown resource status: %s", status.String())
	}
}
