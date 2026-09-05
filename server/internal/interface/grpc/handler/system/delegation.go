package system

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/linzhengen/hub/server/internal/domain/system/delegation"
	"github.com/linzhengen/hub/server/internal/usecase/system"
	pbv1 "github.com/linzhengen/hub/server/pb/system/delegation/v1"
)

func NewDelegationHandler(delegationUseCase system.DelegationUseCase) pbv1.DelegationServiceServer {
	return &delegationHandler{delegationUseCase: delegationUseCase}
}

type delegationHandler struct {
	delegationUseCase system.DelegationUseCase
}

// CreateDelegation carries no principal from the request on purpose: the use
// case takes it from the authenticated caller, so lending somebody else's
// authority is not expressible rather than merely refused.
func (h delegationHandler) CreateDelegation(
	ctx context.Context,
	request *pbv1.CreateDelegationRequest,
) (*pbv1.CreateDelegationResponse, error) {
	created, err := h.delegationUseCase.Create(ctx, &delegation.Delegation{
		AgentId:       request.AgentId,
		Reason:        request.Reason,
		PermissionIds: request.PermissionIds,
		MaxDepth:      request.MaxDepth,
		ExpiresAt:     timestampPbToTimePtr(request.ExpiresAt),
	})
	if err != nil {
		return nil, err
	}
	return &pbv1.CreateDelegationResponse{Delegation: delegationDomainToPb(created)}, nil
}

func (h delegationHandler) ListDelegations(
	ctx context.Context,
	request *pbv1.ListDelegationsRequest,
) (*pbv1.ListDelegationsResponse, error) {
	items, total, err := h.delegationUseCase.List(ctx, delegation.ListParams{
		Limit:           request.Limit,
		Offset:          request.Offset,
		AgentId:         request.GetAgentId(),
		PrincipalUserId: request.GetPrincipalUserId(),
		IncludeRevoked:  request.IncludeRevoked,
	})
	if err != nil {
		return nil, err
	}
	pbItems := make([]*pbv1.Delegation, 0, len(items))
	for _, item := range items {
		pbItems = append(pbItems, delegationDomainToPb(item))
	}
	return &pbv1.ListDelegationsResponse{Delegations: pbItems, Total: total}, nil
}

func (h delegationHandler) RevokeDelegation(
	ctx context.Context,
	request *pbv1.RevokeDelegationRequest,
) (*pbv1.RevokeDelegationResponse, error) {
	revoked, err := h.delegationUseCase.Revoke(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &pbv1.RevokeDelegationResponse{Delegation: delegationDomainToPb(revoked)}, nil
}

func delegationDomainToPb(d *delegation.Delegation) *pbv1.Delegation {
	pb := &pbv1.Delegation{
		Id:              d.Id,
		AgentId:         d.AgentId,
		PrincipalUserId: d.PrincipalUserId,
		GrantedByUserId: d.GrantedByUserId,
		OrgId:           d.OrgId,
		Reason:          d.Reason,
		PermissionIds:   d.PermissionIds,
		MaxDepth:        d.MaxDepth,
		CreatedAt:       timestamppb.New(d.CreatedAt),
		UpdatedAt:       timestamppb.New(d.UpdatedAt),
	}
	if d.ExpiresAt != nil {
		pb.ExpiresAt = timestamppb.New(*d.ExpiresAt)
	}
	if d.RevokedAt != nil {
		pb.RevokedAt = timestamppb.New(*d.RevokedAt)
	}
	return pb
}
