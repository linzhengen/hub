package ai

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	chatDomain "github.com/linzhengen/hub/server/internal/domain/ai/chat"
	aiUseCase "github.com/linzhengen/hub/server/internal/usecase/ai"
	pbchatv1 "github.com/linzhengen/hub/server/pb/ai/chat/v1"
)

func NewChatHandler(uc aiUseCase.ChatUseCase) pbchatv1.ChatServiceServer {
	return &chatHandler{uc: uc}
}

type chatHandler struct {
	uc aiUseCase.ChatUseCase
}

func (h *chatHandler) CreateSession(ctx context.Context, req *pbchatv1.CreateSessionRequest) (*pbchatv1.CreateSessionResponse, error) {
	session, err := h.uc.CreateSession(ctx, req.Title)
	if err != nil {
		return nil, err
	}
	return &pbchatv1.CreateSessionResponse{Session: sessionToPb(session)}, nil
}

func (h *chatHandler) ListSessions(ctx context.Context, _ *pbchatv1.ListSessionsRequest) (*pbchatv1.ListSessionsResponse, error) {
	sessions, err := h.uc.ListSessions(ctx)
	if err != nil {
		return nil, err
	}
	pbSessions := make([]*pbchatv1.Session, len(sessions))
	for i, s := range sessions {
		pbSessions[i] = sessionToPb(s)
	}
	return &pbchatv1.ListSessionsResponse{Sessions: pbSessions}, nil
}

func (h *chatHandler) DeleteSession(ctx context.Context, req *pbchatv1.DeleteSessionRequest) (*pbchatv1.DeleteSessionResponse, error) {
	if err := h.uc.DeleteSession(ctx, req.Id); err != nil {
		return nil, err
	}
	return &pbchatv1.DeleteSessionResponse{}, nil
}

func (h *chatHandler) SendMessage(req *pbchatv1.SendMessageRequest, stream pbchatv1.ChatService_SendMessageServer) error {
	deltas, err := h.uc.SendMessage(stream.Context(), req.Id, req.Content)
	if err != nil {
		return err
	}
	return forward(deltas, func(delta chatDomain.Delta) error {
		return stream.Send(&pbchatv1.SendMessageResponse{
			Delta:        delta.Text,
			Done:         delta.Done,
			ToolCall:     toolCallToPb(delta.Tool),
			ToolProposal: proposalToPb(delta.Proposal),
		})
	})
}

func (h *chatHandler) ConfirmToolCall(
	req *pbchatv1.ConfirmToolCallRequest,
	stream pbchatv1.ChatService_ConfirmToolCallServer,
) error {
	deltas, err := h.uc.ConfirmToolCall(stream.Context(), req.Id, req.Approved)
	if err != nil {
		return err
	}
	return forward(deltas, func(delta chatDomain.Delta) error {
		return stream.Send(&pbchatv1.ConfirmToolCallResponse{
			Delta:        delta.Text,
			Done:         delta.Done,
			ToolCall:     toolCallToPb(delta.Tool),
			ToolProposal: proposalToPb(delta.Proposal),
		})
	})
}

// forward relays the use case's frames, which SendMessage and ConfirmToolCall
// do identically - the second continues the answer the first stopped, so a
// difference between the two would show up as the conversation behaving
// differently either side of an approval.
//
// A proposal ends the stream: the caller has to answer it before there is any
// more to say, and grpc-gateway gives them no way to answer this one.
func forward(deltas <-chan chatDomain.Delta, send func(chatDomain.Delta) error) error {
	for delta := range deltas {
		if delta.Error != nil {
			return status.Errorf(codes.Internal, "stream error: %v", delta.Error)
		}
		if err := send(delta); err != nil {
			return err
		}
		if delta.Done || delta.Proposal != nil {
			return nil
		}
	}
	return nil
}

func toolCallToPb(t *chatDomain.ToolCall) *pbchatv1.ToolCall {
	if t == nil {
		return nil
	}
	return &pbchatv1.ToolCall{Name: t.Name, Arguments: t.Arguments}
}

func proposalToPb(p *chatDomain.ToolProposal) *pbchatv1.ToolProposal {
	if p == nil {
		return nil
	}
	return &pbchatv1.ToolProposal{Id: p.Id, Name: p.ToolName, Arguments: p.Arguments}
}

func (h *chatHandler) ListMessages(ctx context.Context, req *pbchatv1.ListMessagesRequest) (*pbchatv1.ListMessagesResponse, error) {
	messages, err := h.uc.ListMessages(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	pbMessages := make([]*pbchatv1.Message, len(messages))
	for i, m := range messages {
		pbMessages[i] = messageToPb(m)
	}
	return &pbchatv1.ListMessagesResponse{Messages: pbMessages}, nil
}

func sessionToPb(s *chatDomain.Session) *pbchatv1.Session {
	return &pbchatv1.Session{
		Id:        s.Id,
		UserId:    s.UserId,
		Title:     s.Title,
		CreatedAt: timestamppb.New(s.CreatedAt),
	}
}

func messageToPb(m *chatDomain.Message) *pbchatv1.Message {
	return &pbchatv1.Message{
		Id:        m.Id,
		SessionId: m.SessionId,
		Role:      string(m.Role),
		Content:   m.Content,
		CreatedAt: timestamppb.New(m.CreatedAt),
	}
}
