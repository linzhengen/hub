package ai

import (
	"context"
	"database/sql"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/linzhengen/hub/server/config"
	"github.com/linzhengen/hub/server/internal/domain/ai/chat"
	"github.com/linzhengen/hub/server/internal/domain/audit"
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

// errProposalNotFound covers both "no such proposal" and "somebody else's",
// which the repository makes indistinguishable on purpose.
var errProposalNotFound = status.Error(codes.NotFound, "proposal not found")

// errProposalDecided is what a second answer to the same proposal gets. It is
// FailedPrecondition rather than an error the client can retry: the decision
// stands, and re-sending it must not run the change again.
var errProposalDecided = status.Error(codes.FailedPrecondition, "this proposal has already been answered")

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
	// ConfirmToolCall answers a proposed change and takes the conversation back
	// up. The change runs here, and only when approved is true.
	ConfirmToolCall(ctx context.Context, proposalId string, approved bool) (<-chan chat.Delta, error)
	ListMessages(ctx context.Context, sessionId string) ([]*chat.Message, error)
}

// NewChatUseCase builds the chat use case. tokenBudget caps what one session
// may spend; zero or less means no cap.
func NewChatUseCase(repo chat.Repository, svc chat.Service, cfg config.EnvConfig) ChatUseCase {
	return &chatUseCase{repo: repo, svc: svc, tokenBudget: cfg.SessionTokenBudget}
}

type chatUseCase struct {
	repo        chat.Repository
	svc         chat.Service
	tokenBudget int64
}

// errBudgetSpent stops a conversation that has spent its allowance. It is
// ResourceExhausted so a client can tell it apart from a fault: nothing is
// broken, this conversation is finished, and a new one starts with a fresh
// budget.
var errBudgetSpent = status.Error(codes.ResourceExhausted,
	"this conversation has used its token budget; start a new chat to carry on")

// withinBudget reports whether the session may spend more.
func (uc *chatUseCase) withinBudget(session *chat.Session) bool {
	return uc.tokenBudget <= 0 || session.TokensUsed < uc.tokenBudget
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

	session, err := uc.findOwnedSession(ctx, sessionId, userId)
	if err != nil {
		return nil, err
	}
	// Checked before the message is stored: a message that cannot be answered
	// should not join the history and be paid for on the next turn.
	if !uc.withinBudget(session) {
		return nil, errBudgetSpent
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

	// Everything the assistant does from here runs on the user's behalf, so the
	// context carries which conversation asked. Any tool call it makes is
	// recorded against this user with the ai_chat channel and this session id -
	// the assistant is the route, the user is still the actor.
	ctx = audit.NewContext(ctx, audit.Source{
		Channel:   audit.ChannelAIChat,
		SessionId: sessionId,
	})

	upstream, err := uc.svc.Send(ctx, history)
	if err != nil {
		logger.Errorf("SendMessage LLM send: %v", err)
		return nil, err
	}
	return uc.relay(ctx, sessionId, upstream), nil
}

// ConfirmToolCall answers a proposed change.
//
// The decision is recorded before anything runs, and only a proposal still
// pending can be decided, so an approval cannot be replayed into a second
// change by sending the same id twice.
func (uc *chatUseCase) ConfirmToolCall(
	ctx context.Context,
	proposalId string,
	approved bool,
) (<-chan chat.Delta, error) {
	userId, ok := contextx.GetUserID(ctx)
	if !ok {
		return nil, errNoUser
	}

	proposal, err := uc.repo.FindProposal(ctx, proposalId, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errProposalNotFound
		}
		logger.Errorf("ConfirmToolCall find: %v", err)
		return nil, err
	}
	if !proposal.Pending() {
		return nil, errProposalDecided
	}

	status := chat.ProposalRejected
	if approved {
		status = chat.ProposalApproved
	}
	// Claiming the proposal first is what makes one approval mean one change:
	// the update matches only a pending row, so a second decision arriving at
	// the same moment finds nothing to claim and is refused here rather than
	// running the tool again.
	decided, err := uc.repo.Decide(ctx, proposalId, status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errProposalDecided
		}
		logger.Errorf("ConfirmToolCall decide: %v", err)
		return nil, err
	}

	history, err := uc.repo.ListMessages(ctx, decided.SessionId, userId)
	if err != nil {
		logger.Errorf("ConfirmToolCall load history: %v", err)
		return nil, err
	}

	ctx = audit.NewContext(ctx, audit.Source{
		Channel:    audit.ChannelAIChat,
		SessionId:  decided.SessionId,
		ApprovalId: decided.Id,
	})

	upstream, err := uc.svc.Resume(ctx, history, decided.Continuation, approved)
	if err != nil {
		logger.Errorf("ConfirmToolCall resume: %v", err)
		return nil, err
	}
	return uc.relay(ctx, decided.SessionId, upstream), nil
}

// relay forwards the backend's stream to the caller, storing the answer once it
// is complete and persisting a proposal so the pause survives the request.
//
// The stream has to end for the user to be able to answer it - grpc-gateway's
// server-streaming carries nothing back - so a proposal is not a message in the
// conversation but a row that outlives it.
func (uc *chatUseCase) relay(
	ctx context.Context,
	sessionId string,
	upstream <-chan chat.Delta,
) <-chan chat.Delta {
	out := make(chan chat.Delta, 64)
	go func() {
		defer close(out)
		var full string
		for delta := range upstream {
			if delta.Tokens > 0 {
				// Recorded as it is reported, so a conversation that fails or
				// stops for approval has still paid for what it spent. It is
				// not forwarded: what a round cost is not part of the answer.
				if err := uc.repo.AddTokens(ctx, sessionId, delta.Tokens); err != nil {
					logger.Errorf("relay account for tokens: %v", err)
				}
				continue
			}
			if delta.Error != nil {
				out <- delta
				return
			}
			if delta.Proposal != nil {
				uc.pause(ctx, sessionId, full, delta, out)
				return
			}
			full += delta.Text
			out <- delta
		}
		if err := uc.saveAnswer(ctx, sessionId, full); err != nil {
			out <- chat.Delta{Error: err}
			return
		}
		out <- chat.Delta{Done: true}
	}()
	return out
}

// pause stores the proposal and ends the stream. Whatever the assistant said on
// the way to asking is stored too, so the transcript still reads as a
// conversation once the page is reloaded.
func (uc *chatUseCase) pause(
	ctx context.Context,
	sessionId, said string,
	delta chat.Delta,
	out chan<- chat.Delta,
) {
	proposal := delta.Proposal
	proposal.SessionId = sessionId
	if err := uc.repo.CreateProposal(ctx, proposal); err != nil {
		logger.Errorf("relay save proposal: %v", err)
		out <- chat.Delta{Error: err}
		return
	}
	if said != "" {
		if err := uc.saveAnswer(ctx, sessionId, said); err != nil {
			out <- chat.Delta{Error: err}
			return
		}
	}
	out <- delta
}

func (uc *chatUseCase) saveAnswer(ctx context.Context, sessionId, content string) error {
	if content == "" {
		return nil
	}
	message := &chat.Message{SessionId: sessionId, Role: chat.RoleAssistant, Content: content}
	if err := uc.repo.CreateMessage(ctx, message); err != nil {
		logger.Errorf("save assistant message: %v", err)
		return err
	}
	return nil
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
