package system

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/linzhengen/hub/server/internal/domain/auth"
	"github.com/linzhengen/hub/server/internal/usecase/system"
	pbv1 "github.com/linzhengen/hub/server/pb/system/access/v1"
)

func NewAccessHandler(accessUseCase system.AccessUseCase) pbv1.AccessServiceServer {
	return &accessHandler{accessUseCase: accessUseCase}
}

type accessHandler struct {
	accessUseCase system.AccessUseCase
}

func (h accessHandler) ExplainUserAccess(
	ctx context.Context,
	request *pbv1.ExplainUserAccessRequest,
) (*pbv1.ExplainUserAccessResponse, error) {
	paths, err := h.accessUseCase.ExplainUserAccess(ctx, request.UserId, request.Resource, request.Action)
	if err != nil {
		return nil, err
	}
	return &pbv1.ExplainUserAccessResponse{
		Allowed: len(paths) > 0,
		Paths:   accessPathsDomainToPb(paths),
	}, nil
}

func (h accessHandler) ListPrincipalsForOperation(
	ctx context.Context,
	request *pbv1.ListPrincipalsForOperationRequest,
) (*pbv1.ListPrincipalsForOperationResponse, error) {
	principals, err := h.accessUseCase.PrincipalsForOperation(ctx, request.Resource, request.Action)
	if err != nil {
		return nil, err
	}
	pbPrincipals := make([]*pbv1.Principal, 0, len(principals))
	for _, principal := range principals {
		pbPrincipals = append(pbPrincipals, &pbv1.Principal{
			UserId:   principal.UserId,
			Username: principal.Username,
			Paths:    accessPathsDomainToPb(principal.Paths),
		})
	}
	return &pbv1.ListPrincipalsForOperationResponse{Principals: pbPrincipals}, nil
}

// expiresAtToPb leaves the field absent for a route that does not end, rather
// than sending a zero time that a reader would have to know to ignore.
func expiresAtToPb(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func accessPathsDomainToPb(paths []auth.AccessPath) []*pbv1.AccessPath {
	pbPaths := make([]*pbv1.AccessPath, 0, len(paths))
	for _, path := range paths {
		pbPaths = append(pbPaths, &pbv1.AccessPath{
			GroupId:      path.GroupId,
			GroupName:    path.GroupName,
			RoleId:       path.RoleId,
			RoleName:     path.RoleName,
			PermissionId: path.PermissionId,
			Resource:     path.Object,
			Action:       path.Action,
			ExpiresAt:    expiresAtToPb(path.ExpiresAt),
		})
	}
	return pbPaths
}
