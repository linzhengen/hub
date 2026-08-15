package system

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/linzhengen/hub/server/internal/domain/audit"
	"github.com/linzhengen/hub/server/internal/usecase/system"
	pbv1 "github.com/linzhengen/hub/server/pb/system/audit/v1"
)

func NewAuditHandler(auditUseCase system.AuditUseCase) pbv1.AuditServiceServer {
	return &auditHandler{auditUseCase: auditUseCase}
}

type auditHandler struct {
	auditUseCase system.AuditUseCase
}

func (h auditHandler) ListAuditLog(
	ctx context.Context,
	request *pbv1.ListAuditLogRequest,
) (*pbv1.ListAuditLogResponse, error) {
	items, total, err := h.auditUseCase.List(ctx, &system.ListAuditLogQueryParams{
		Limit:       request.Limit,
		Offset:      request.Offset,
		ActorUserId: request.GetActorUserId(),
		Resource:    request.Resource,
		Action:      request.Action,
		TargetId:    request.TargetId,
		Channel:     channelPbToDomain(request.Channel),
		Since:       timestampPbToTime(request.Since),
		Until:       timestampPbToTime(request.Until),
	})
	if err != nil {
		return nil, err
	}
	pbItems := make([]*pbv1.AuditLog, 0, len(items))
	for _, item := range items {
		pbItems = append(pbItems, auditDomainToPb(item))
	}
	return &pbv1.ListAuditLogResponse{AuditLogs: pbItems, Total: total}, nil
}

func auditDomainToPb(e *audit.Entry) *pbv1.AuditLog {
	return &pbv1.AuditLog{
		Id:          e.Id,
		ActorUserId: e.ActorUserId,
		Channel:     channelDomainToPb(e.Channel),
		SessionId:   e.SessionId,
		Resource:    e.Resource,
		Action:      e.Action,
		TargetId:    e.TargetId,
		Arguments:   string(e.Arguments),
		Succeeded:   e.Succeeded,
		Error:       e.Error,
		Client:      e.Client,
		CreatedAt:   timestamppb.New(e.CreatedAt),
	}
}

func channelDomainToPb(c audit.Channel) pbv1.Channel {
	switch c {
	case audit.ChannelAPI:
		return pbv1.Channel_CHANNEL_API
	case audit.ChannelAIChat:
		return pbv1.Channel_CHANNEL_AI_CHAT
	default:
		return pbv1.Channel_CHANNEL_UNSPECIFIED
	}
}

// channelPbToDomain maps a filter. CHANNEL_UNSPECIFIED means "any", so it maps
// to the empty channel the use case reads as no filter.
func channelPbToDomain(c pbv1.Channel) audit.Channel {
	switch c {
	case pbv1.Channel_CHANNEL_API:
		return audit.ChannelAPI
	case pbv1.Channel_CHANNEL_AI_CHAT:
		return audit.ChannelAIChat
	default:
		return ""
	}
}

// timestampPbToTimePtr maps an absent expiry to nil, which is how every layer
// below spells "this grant does not end".
func timestampPbToTimePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

func timestampPbToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}
