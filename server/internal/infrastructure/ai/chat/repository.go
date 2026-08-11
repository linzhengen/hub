package chat

import (
	"context"

	chatDomain "github.com/linzhengen/hub/server/internal/domain/ai/chat"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence"
)

type repositoryImpl struct {
	q persistence.Querier
}

func New(q persistence.Querier) chatDomain.Repository {
	return &repositoryImpl{q: q}
}

func (r repositoryImpl) CreateSession(ctx context.Context, s *chatDomain.Session) error {
	row, err := persistence.GetQ(ctx, r.q).CreateChatSession(ctx, s.UserId, s.Title)
	if err != nil {
		return err
	}
	s.Id = row.ID
	s.CreatedAt = row.CreatedAt
	return nil
}

func (r repositoryImpl) FindSession(ctx context.Context, id string) (*chatDomain.Session, error) {
	row, err := persistence.GetQ(ctx, r.q).SelectChatSessionById(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertSession(row), nil
}

func (r repositoryImpl) ListSessions(ctx context.Context, userId string) ([]*chatDomain.Session, error) {
	rows, err := persistence.GetQ(ctx, r.q).SelectChatSessionsByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}
	sessions := make([]*chatDomain.Session, len(rows))
	for i, row := range rows {
		sessions[i] = convertSession(row)
	}
	return sessions, nil
}

func (r repositoryImpl) DeleteSession(ctx context.Context, id string) error {
	return persistence.GetQ(ctx, r.q).DeleteChatSession(ctx, id)
}

func (r repositoryImpl) CreateMessage(ctx context.Context, m *chatDomain.Message) error {
	row, err := persistence.GetQ(ctx, r.q).CreateChatMessage(ctx, m.SessionId, string(m.Role), m.Content)
	if err != nil {
		return err
	}
	m.Id = row.ID
	m.CreatedAt = row.CreatedAt
	return nil
}

func (r repositoryImpl) ListMessages(ctx context.Context, sessionId string) ([]*chatDomain.Message, error) {
	rows, err := persistence.GetQ(ctx, r.q).SelectChatMessagesBySessionId(ctx, sessionId)
	if err != nil {
		return nil, err
	}
	messages := make([]*chatDomain.Message, len(rows))
	for i, row := range rows {
		messages[i] = convertMessage(row)
	}
	return messages, nil
}

func convertSession(m *persistence.ChatSessionModel) *chatDomain.Session {
	return &chatDomain.Session{
		Id:        m.ID,
		UserId:    m.UserID,
		Title:     m.Title,
		CreatedAt: m.CreatedAt,
	}
}

func convertMessage(m *persistence.ChatMessageModel) *chatDomain.Message {
	return &chatDomain.Message{
		Id:        m.ID,
		SessionId: m.SessionID,
		Role:      chatDomain.Role(m.Role),
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
}
