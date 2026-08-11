package chat

import "context"

type Repository interface {
	CreateSession(ctx context.Context, s *Session) error
	FindSession(ctx context.Context, id string) (*Session, error)
	ListSessions(ctx context.Context, userId string) ([]*Session, error)
	DeleteSession(ctx context.Context, id string) error

	CreateMessage(ctx context.Context, m *Message) error
	ListMessages(ctx context.Context, sessionId string) ([]*Message, error)
}
