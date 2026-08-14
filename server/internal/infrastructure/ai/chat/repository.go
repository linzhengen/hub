package chat

import (
	"context"
	"encoding/json"

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

func (r repositoryImpl) FindSession(ctx context.Context, id, userId string) (*chatDomain.Session, error) {
	row, err := persistence.GetQ(ctx, r.q).SelectChatSessionById(ctx, id, userId)
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

func (r repositoryImpl) DeleteSession(ctx context.Context, id, userId string) error {
	return persistence.GetQ(ctx, r.q).DeleteChatSession(ctx, id, userId)
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

func (r repositoryImpl) ListMessages(ctx context.Context, sessionId, userId string) ([]*chatDomain.Message, error) {
	rows, err := persistence.GetQ(ctx, r.q).SelectChatMessagesBySessionId(ctx, sessionId, userId)
	if err != nil {
		return nil, err
	}
	messages := make([]*chatDomain.Message, len(rows))
	for i, row := range rows {
		messages[i] = convertMessage(row)
	}
	return messages, nil
}

func (r repositoryImpl) AddTokens(ctx context.Context, sessionId string, tokens int64) error {
	if tokens <= 0 {
		return nil
	}
	return persistence.GetQ(ctx, r.q).AddChatSessionTokens(ctx, sessionId, tokens)
}

func (r repositoryImpl) CreateProposal(ctx context.Context, p *chatDomain.ToolProposal) error {
	arguments := p.Arguments
	if arguments == "" {
		arguments = "{}"
	}
	row, err := persistence.GetQ(ctx, r.q).CreateChatToolProposal(
		ctx, p.SessionId, p.ToolName, json.RawMessage(arguments), p.Continuation)
	if err != nil {
		return err
	}
	p.Id = row.ID
	p.Status = chatDomain.ProposalStatus(row.Status)
	p.CreatedAt = row.CreatedAt
	return nil
}

func (r repositoryImpl) FindProposal(ctx context.Context, id, userId string) (*chatDomain.ToolProposal, error) {
	row, err := persistence.GetQ(ctx, r.q).SelectChatToolProposalById(ctx, id, userId)
	if err != nil {
		return nil, err
	}
	return convertProposal(row), nil
}

func (r repositoryImpl) Decide(
	ctx context.Context,
	id string,
	status chatDomain.ProposalStatus,
) (*chatDomain.ToolProposal, error) {
	row, err := persistence.GetQ(ctx, r.q).DecideChatToolProposal(ctx, id, string(status))
	if err != nil {
		return nil, err
	}
	return convertProposal(row), nil
}

func convertProposal(m *persistence.ChatToolProposalModel) *chatDomain.ToolProposal {
	p := &chatDomain.ToolProposal{
		Id:           m.ID,
		SessionId:    m.SessionID,
		ToolName:     m.ToolName,
		Arguments:    string(m.Arguments),
		Continuation: m.Continuation,
		Status:       chatDomain.ProposalStatus(m.Status),
		CreatedAt:    m.CreatedAt,
	}
	if m.DecidedAt.Valid {
		decided := m.DecidedAt.Time
		p.DecidedAt = &decided
	}
	return p
}

func convertSession(m *persistence.ChatSessionModel) *chatDomain.Session {
	return &chatDomain.Session{
		Id:         m.ID,
		UserId:     m.UserID,
		Title:      m.Title,
		CreatedAt:  m.CreatedAt,
		TokensUsed: m.TokensUsed,
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
