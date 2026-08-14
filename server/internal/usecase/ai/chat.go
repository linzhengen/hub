package ai

import (
	"context"
	"database/sql"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/linzhengen/hub/server/internal/domain/ai/chat"
	"github.com/linzhengen/hub/server/internal/domain/contextx"
	"github.com/linzhengen/hub/server/pkg/logger"
)

// errNoUser is what a caller sees when the authentication interceptor has not
// run. It is a status rather than a plain error so the gateway answers 401
// instead of turning it into a 500 that reads like an outage.
var errNoUser = status.Error(codes.Unauthenticated, "unauthenticated")

// errSessionNotFound answers both "no such session" and "somebody else's
// session" - the repository scopes its queries by user, so the two are the same
// row-not-found and telling them apart would leak which ids exist.
var errSessionNotFound = status.Error(codes.NotFound, "chat session not found")

// findOwnedSession loads a session the user owns, mapping the empty result to
// errSessionNotFound.
func (uc *chatUseCase) findOwnedSession(ctx context.Context, sessionId, userId string) (*chat.Session, error) {
	session, err := uc.repo.FindSession(ctx, sessionId, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errSessionNotFound
		}
		logger.Errorf("findOwnedSession: %v", err)
		return nil, err
	}
	return session, nil
}

type ChatUseCase interface {
	CreateSession(ctx context.Context, title string) (*chat.Session, error)
	ListSessions(ctx context.Context) ([]*chat.Session, error)
	DeleteSession(ctx context.Context, sessionId string) error
	SendMessage(ctx context.Context, sessionId, content string) (<-chan chat.Delta, error)
	ListMessages(ctx context.Context, sessionId string) ([]*chat.Message, error)
}

func NewChatUseCase(repo chat.Repository, svc chat.Service) ChatUseCase {
	return &chatUseCase{repo: repo, svc: svc}
}

type chatUseCase struct {
	repo chat.Repository
	svc  chat.Service
}

func (uc *chatUseCase) CreateSession(ctx context.Context, title string) (*chat.Session, error) {
	userId, ok := contextx.GetUserID(ctx)
	if !ok {
		return nil, errNoUser
	}
	s := &chat.Session{UserId: userId, Title: title}
	if err := uc.repo.CreateSession(ctx, s); err != nil {
		logger.Errorf("CreateSession: %v", err)
		return nil, err
	}
	return s, nil
}

func (uc *chatUseCase) ListSessions(ctx context.Context) ([]*chat.Session, error) {
	userId, ok := contextx.GetUserID(ctx)
	if !ok {
		return nil, errNoUser
	}
	sessions, err := uc.repo.ListSessions(ctx, userId)
	if err != nil {
		logger.Errorf("ListSessions: %v", err)
		return nil, err
	}
	return sessions, nil
}

func (uc *chatUseCase) DeleteSession(ctx context.Context, sessionId string) error {
	userId, ok := contextx.GetUserID(ctx)
	if !ok {
		return errNoUser
	}
	if _, err := uc.findOwnedSession(ctx, sessionId, userId); err != nil {
		return err
	}
	if err := uc.repo.DeleteSession(ctx, sessionId, userId); err != nil {
		logger.Errorf("DeleteSession delete: %v", err)
		return err
	}
	return nil
}

func (uc *chatUseCase) SendMessage(ctx context.Context, sessionId, content string) (<-chan chat.Delta, error) {
	userId, ok := contextx.GetUserID(ctx)
	if !ok {
		return nil, errNoUser
	}

	if _, err := uc.findOwnedSession(ctx, sessionId, userId); err != nil {
		return nil, err
	}

	userMsg := &chat.Message{SessionId: sessionId, Role: chat.RoleUser, Content: content}
	if err := uc.repo.CreateMessage(ctx, userMsg); err != nil {
		logger.Errorf("SendMessage save user message: %v", err)
		return nil, err
	}

	history, err := uc.repo.ListMessages(ctx, sessionId, userId)
	if err != nil {
		logger.Errorf("SendMessage load history: %v", err)
		return nil, err
	}

	upstream, err := uc.svc.Send(ctx, history)
	if err != nil {
		logger.Errorf("SendMessage LLM send: %v", err)
		return nil, err
	}

	out := make(chan chat.Delta, 64)
	go func() {
		defer close(out)
		var full string
		for delta := range upstream {
			if delta.Error != nil {
				out <- delta
				return
			}
			full += delta.Text
			out <- delta
		}
		assistantMsg := &chat.Message{SessionId: sessionId, Role: chat.RoleAssistant, Content: full}
		if err := uc.repo.CreateMessage(ctx, assistantMsg); err != nil {
			logger.Errorf("SendMessage save assistant message: %v", err)
			out <- chat.Delta{Error: err}
			return
		}
		out <- chat.Delta{Done: true}
	}()

	return out, nil
}

func (uc *chatUseCase) ListMessages(ctx context.Context, sessionId string) ([]*chat.Message, error) {
	userId, ok := contextx.GetUserID(ctx)
	if !ok {
		return nil, errNoUser
	}
	if _, err := uc.findOwnedSession(ctx, sessionId, userId); err != nil {
		return nil, err
	}
	messages, err := uc.repo.ListMessages(ctx, sessionId, userId)
	if err != nil {
		logger.Errorf("ListMessages: %v", err)
		return nil, err
	}
	return messages, nil
}
