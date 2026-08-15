package system

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/linzhengen/hub/server/internal/domain/access"
	"github.com/linzhengen/hub/server/internal/domain/audit"
	"github.com/linzhengen/hub/server/internal/usecase/system"
	pbv1 "github.com/linzhengen/hub/server/pb/system/access/v1"
)

func NewAccessRequestHandler(accessRequestUseCase system.AccessRequestUseCase) pbv1.AccessRequestServiceServer {
	return &accessRequestHandler{accessRequestUseCase: accessRequestUseCase}
}

type accessRequestHandler struct {
	accessRequestUseCase system.AccessRequestUseCase
}

func (h accessRequestHandler) CreateAccessRequest(
	ctx context.Context,
	request *pbv1.CreateAccessRequestRequest,
) (*pbv1.CreateAccessRequestResponse, error) {
	origin, sessionId := originOf(ctx)
	r, err := h.accessRequestUseCase.Create(
		ctx,
		request.GetSubjectUserId(),
		request.GroupId,
		request.Reason,
		timestampPbToTimePtr(request.RequestedUntil),
		origin,
		sessionId,
	)
	if err != nil {
		return nil, err
	}
	return &pbv1.CreateAccessRequestResponse{AccessRequest: accessRequestDomainToPb(r)}, nil
}

func (h accessRequestHandler) ListAccessRequests(
	ctx context.Context,
	request *pbv1.ListAccessRequestsRequest,
) (*pbv1.ListAccessRequestsResponse, error) {
	items, total, err := h.accessRequestUseCase.List(ctx, access.ListParams{
		Limit:           request.Limit,
		Offset:          request.Offset,
		RequesterUserId: request.GetRequesterUserId(),
		SubjectUserId:   request.GetSubjectUserId(),
		GroupId:         request.GetGroupId(),
		Status:          requestStatusPbToDomain(request.Status),
	})
	if err != nil {
		return nil, err
	}
	pbItems := make([]*pbv1.AccessRequest, 0, len(items))
	for _, item := range items {
		pbItems = append(pbItems, accessRequestDomainToPb(item))
	}
	return &pbv1.ListAccessRequestsResponse{AccessRequests: pbItems, Total: total}, nil
}

func (h accessRequestHandler) CancelAccessRequest(
	ctx context.Context,
	request *pbv1.CancelAccessRequestRequest,
) (*pbv1.CancelAccessRequestResponse, error) {
	r, err := h.accessRequestUseCase.Cancel(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &pbv1.CancelAccessRequestResponse{AccessRequest: accessRequestDomainToPb(r)}, nil
}

func (h accessRequestHandler) DecideAccessRequest(
	ctx context.Context,
	request *pbv1.DecideAccessRequestRequest,
) (*pbv1.DecideAccessRequestResponse, error) {
	r, err := h.accessRequestUseCase.Decide(ctx, request.Id, request.Approved, request.Comment)
	if err != nil {
		return nil, err
	}
	return &pbv1.DecideAccessRequestResponse{AccessRequest: accessRequestDomainToPb(r)}, nil
}

// originOf reads the surface from the same context mark the audit log reads, so
// a request raised by the assistant is labelled by the one thing that actually
// knows - the tool executor that dispatched it - rather than by a field the
// caller could set to anything.
func originOf(ctx context.Context) (access.Origin, string) {
	source := audit.FromContext(ctx)
	if source.Channel == audit.ChannelAIChat {
		return access.OriginAIChat, source.SessionId
	}
	return access.OriginConsole, ""
}

func accessRequestDomainToPb(r *access.Request) *pbv1.AccessRequest {
	return &pbv1.AccessRequest{
		Id:              r.Id,
		RequesterUserId: r.RequesterUserId,
		SubjectUserId:   r.SubjectUserId,
		GroupId:         r.GroupId,
		Reason:          r.Reason,
		RequestedUntil:  expiresAtToPb(r.RequestedUntil),
		Status:          requestStatusDomainToPb(r.Status),
		Origin:          requestOriginDomainToPb(r.Origin),
		SessionId:       r.SessionId,
		DecidedByUserId: r.DecidedByUserId,
		DecidedAt:       expiresAtToPb(r.DecidedAt),
		DecisionComment: r.DecisionComment,
		CreatedAt:       timestamppb.New(r.CreatedAt),
		UpdatedAt:       timestamppb.New(r.UpdatedAt),
	}
}

func requestStatusDomainToPb(s access.Status) pbv1.RequestStatus {
	switch s {
	case access.StatusPending:
		return pbv1.RequestStatus_REQUEST_STATUS_PENDING
	case access.StatusApproved:
		return pbv1.RequestStatus_REQUEST_STATUS_APPROVED
	case access.StatusRejected:
		return pbv1.RequestStatus_REQUEST_STATUS_REJECTED
	case access.StatusCancelled:
		return pbv1.RequestStatus_REQUEST_STATUS_CANCELLED
	default:
		return pbv1.RequestStatus_REQUEST_STATUS_UNSPECIFIED
	}
}

// requestStatusPbToDomain maps a filter. UNSPECIFIED means "any", so it maps to
// the empty status the use case reads as no filter.
func requestStatusPbToDomain(s pbv1.RequestStatus) access.Status {
	switch s {
	case pbv1.RequestStatus_REQUEST_STATUS_PENDING:
		return access.StatusPending
	case pbv1.RequestStatus_REQUEST_STATUS_APPROVED:
		return access.StatusApproved
	case pbv1.RequestStatus_REQUEST_STATUS_REJECTED:
		return access.StatusRejected
	case pbv1.RequestStatus_REQUEST_STATUS_CANCELLED:
		return access.StatusCancelled
	default:
		return ""
	}
}

func requestOriginDomainToPb(o access.Origin) pbv1.RequestOrigin {
	switch o {
	case access.OriginConsole:
		return pbv1.RequestOrigin_REQUEST_ORIGIN_CONSOLE
	case access.OriginCLI:
		return pbv1.RequestOrigin_REQUEST_ORIGIN_CLI
	case access.OriginAIChat:
		return pbv1.RequestOrigin_REQUEST_ORIGIN_AI_CHAT
	default:
		return pbv1.RequestOrigin_REQUEST_ORIGIN_UNSPECIFIED
	}
}
