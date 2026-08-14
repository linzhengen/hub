package chat

import "context"

// Repository reads and writes chat sessions and their messages.
//
// Every method that addresses a session by id also takes the id of the user it
// must belong to, and the queries filter on it. A session belonging to someone
// else is therefore not "found and then refused" but simply absent, which is
// what keeps one user's conversations private even from a caller that reaches
// the repository without passing through the use case.
type Repository interface {
	CreateSession(ctx context.Context, s *Session) error
	FindSession(ctx context.Context, id, userId string) (*Session, error)
	ListSessions(ctx context.Context, userId string) ([]*Session, error)
	DeleteSession(ctx context.Context, id, userId string) error

	CreateMessage(ctx context.Context, m *Message) error
	ListMessages(ctx context.Context, sessionId, userId string) ([]*Message, error)

	// AddTokens accounts for what a conversation spent. It adds rather than
	// sets, so two streams on one session both count.
	AddTokens(ctx context.Context, sessionId string, tokens int64) error

	CreateProposal(ctx context.Context, p *ToolProposal) error
	FindProposal(ctx context.Context, id, userId string) (*ToolProposal, error)
	// Decide records the user's answer, and only for a proposal still pending.
	// It returns sql.ErrNoRows for one already decided, so two decisions racing
	// cannot both run the change.
	Decide(ctx context.Context, id string, status ProposalStatus) (*ToolProposal, error)
}
