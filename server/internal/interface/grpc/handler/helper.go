package handler

import (
	"fmt"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/linzhengen/hub/server/internal/domain/system/group"
	pbgroupv1 "github.com/linzhengen/hub/server/pb/system/group/v1"

	"github.com/linzhengen/hub/server/internal/domain/user"
	pbuserv1 "github.com/linzhengen/hub/server/pb/user/v1"
)

func toUserPbStatus(s user.Status) pbuserv1.User_Status {
	switch s {
	case user.Active:
		return pbuserv1.User_STATUS_ACTIVE
	case user.InActive:
		return pbuserv1.User_STATUS_INACTIVE
	default:
		return pbuserv1.User_STATUS_UNSPECIFIED
	}
}

func toUserDomainStatus(status pbuserv1.User_Status) (user.Status, error) {
	switch status {
	case pbuserv1.User_STATUS_ACTIVE:
		return user.Active, nil
	case pbuserv1.User_STATUS_INACTIVE:
		return user.InActive, nil
	default:
		return "", fmt.Errorf("unknown user status: %s", status.String())
	}
}

func userDomainToPb(m *user.User) *pbuserv1.User {
	return &pbuserv1.User{
		Id:        m.Id,
		Username:  m.Username,
		Email:     m.Email,
		Status:    toUserPbStatus(m.Status),
		CreatedAt: timestamppb.New(m.CreatedAt),
		UpdatedAt: timestamppb.New(m.UpdatedAt),
		GroupIds:  m.GroupIds,
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
	}
}

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
