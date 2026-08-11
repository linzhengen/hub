package ai

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/linzhengen/hub/server/internal/domain/ai/chat"
	"github.com/linzhengen/hub/server/internal/domain/contextx"
	"github.com/linzhengen/hub/server/pkg/logger"
)

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
		return nil, fmt.Errorf("user not found in context")
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
		return nil, fmt.Errorf("user not found in context")
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
		return fmt.Errorf("user not found in context")
	}
	session, err := uc.repo.FindSession(ctx, sessionId)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("session not found")
		}
		logger.Errorf("DeleteSession find: %v", err)
		return err
	}
	if session.UserId != userId {
		return fmt.Errorf("permission denied")
	}
	if err := uc.repo.DeleteSession(ctx, sessionId); err != nil {
		logger.Errorf("DeleteSession delete: %v", err)
		return err
	}
	return nil
}

func (uc *chatUseCase) SendMessage(ctx context.Context, sessionId, content string) (<-chan chat.Delta, error) {
	userId, ok := contextx.GetUserID(ctx)
	if !ok {
		return nil, fmt.Errorf("user not found in context")
	}

	session, err := uc.repo.FindSession(ctx, sessionId)
	if err != nil {
		logger.Errorf("SendMessage find session: %v", err)
		return nil, err
	}
	if session.UserId != userId {
		return nil, fmt.Errorf("permission denied")
	}

	userMsg := &chat.Message{SessionId: sessionId, Role: chat.RoleUser, Content: content}
	if err := uc.repo.CreateMessage(ctx, userMsg); err != nil {
		logger.Errorf("SendMessage save user message: %v", err)
		return nil, err
	}

	history, err := uc.repo.ListMessages(ctx, sessionId)
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
		return nil, fmt.Errorf("user not found in context")
	}
	session, err := uc.repo.FindSession(ctx, sessionId)
	if err != nil {
		logger.Errorf("ListMessages find session: %v", err)
		return nil, err
	}
	if session.UserId != userId {
		return nil, fmt.Errorf("permission denied")
	}
	messages, err := uc.repo.ListMessages(ctx, sessionId)
	if err != nil {
		logger.Errorf("ListMessages: %v", err)
		return nil, err
	}
	return messages, nil
}
