package ai

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/linzhengen/hub/server/config"
	"github.com/linzhengen/hub/server/internal/domain/ai/chat"
	"github.com/linzhengen/hub/server/internal/domain/contextx"
)

const (
	owner      = "11111111-1111-1111-1111-111111111111"
	intruder   = "22222222-2222-2222-2222-222222222222"
	sessionId  = "33333333-3333-3333-3333-333333333333"
	missingId  = "44444444-4444-4444-4444-444444444444"
	proposalId = "55555555-5555-5555-5555-555555555555"
)

type call struct{ id, userId string }

// fakeRepo mirrors what the queries now do: every lookup is scoped by user, so
// a session belonging to somebody else is not found rather than found and
// refused. A fake that ignored userId would let the use case pass this suite
// while the real repository leaked, which is the whole point of the change.
type fakeRepo struct {
	session  *chat.Session
	messages []*chat.Message

	proposal *chat.ToolProposal

	finds    []call
	deletes  []call
	listMsgs []call
	decided  []chat.ProposalStatus
	tokens   int64
}

func (r *fakeRepo) owns(id, userId string) bool {
	return r.session != nil && r.session.Id == id && r.session.UserId == userId
}

func (r *fakeRepo) FindSession(_ context.Context, id, userId string) (*chat.Session, error) {
	r.finds = append(r.finds, call{id, userId})
	if !r.owns(id, userId) {
		return nil, sql.ErrNoRows
	}
	return r.session, nil
}

func (r *fakeRepo) DeleteSession(_ context.Context, id, userId string) error {
	r.deletes = append(r.deletes, call{id, userId})
	return nil
}

func (r *fakeRepo) ListMessages(_ context.Context, sessionId, userId string) ([]*chat.Message, error) {
	r.listMsgs = append(r.listMsgs, call{sessionId, userId})
	if !r.owns(sessionId, userId) {
		return nil, nil
	}
	return r.messages, nil
}

func (r *fakeRepo) CreateSession(_ context.Context, s *chat.Session) error {
	s.Id = sessionId
	r.session = s
	return nil
}

func (r *fakeRepo) CreateMessage(_ context.Context, m *chat.Message) error {
	r.messages = append(r.messages, m)
	return nil
}

func (r *fakeRepo) ListSessions(_ context.Context, userId string) ([]*chat.Session, error) {
	if r.session == nil || r.session.UserId != userId {
		return nil, nil
	}
	return []*chat.Session{r.session}, nil
}

func (r *fakeRepo) AddTokens(_ context.Context, _ string, tokens int64) error {
	r.tokens += tokens
	return nil
}

func (r *fakeRepo) CreateProposal(_ context.Context, p *chat.ToolProposal) error {
	p.Id = proposalId
	p.Status = chat.ProposalPending
	r.proposal = p
	return nil
}

func (r *fakeRepo) FindProposal(_ context.Context, id, userId string) (*chat.ToolProposal, error) {
	if r.proposal == nil || r.proposal.Id != id || !r.owns(r.proposal.SessionId, userId) {
		return nil, sql.ErrNoRows
	}
	return r.proposal, nil
}

func (r *fakeRepo) Decide(_ context.Context, id string, s chat.ProposalStatus) (*chat.ToolProposal, error) {
	// Mirrors the query, which matches only a pending row: a second decision
	// updates nothing.
	if r.proposal == nil || r.proposal.Id != id || !r.proposal.Pending() {
		return nil, sql.ErrNoRows
	}
	r.proposal.Status = s
	r.decided = append(r.decided, s)
	return r.proposal, nil
}

// fakeService streams what the backend would. proposes makes the first Send
// stop on a change, the way the real client does when the model asks for a tool
// that writes.
type fakeService struct {
	proposes bool
	spends   int64
	resumed  int
	approved []bool
}

func (s *fakeService) Send(_ context.Context, _ []*chat.Message) (<-chan chat.Delta, error) {
	ch := make(chan chat.Delta, 3)
	if s.spends > 0 {
		ch <- chat.Delta{Tokens: s.spends}
	}
	if s.proposes {
		ch <- chat.Delta{Proposal: &chat.ToolProposal{
			ToolName:     "add_users_to_group",
			Arguments:    `{"id":"a-group"}`,
			Continuation: []byte("paused"),
			Status:       chat.ProposalPending,
		}}
		close(ch)
		return ch, nil
	}
	ch <- chat.Delta{Done: true}
	close(ch)
	return ch, nil
}

func (s *fakeService) Resume(
	_ context.Context,
	_ []*chat.Message,
	_ []byte,
	approved bool,
) (<-chan chat.Delta, error) {
	s.resumed++
	s.approved = append(s.approved, approved)
	ch := make(chan chat.Delta, 2)
	ch <- chat.Delta{Text: "done"}
	ch <- chat.Delta{Done: true}
	close(ch)
	return ch, nil
}

// budget is generous enough that the tests below are not about it; the ones
// that are set their own.
func testConfig(budget int64) config.EnvConfig {
	var cfg config.EnvConfig
	cfg.SessionTokenBudget = budget
	return cfg
}

func newUseCase(session *chat.Session) (*fakeRepo, ChatUseCase) {
	repo := &fakeRepo{session: session}
	return repo, NewChatUseCase(repo, &fakeService{}, testConfig(1_000_000))
}

// newProposingUseCase builds one whose backend stops on a change.
func newProposingUseCase(session *chat.Session) (*fakeRepo, *fakeService, ChatUseCase) {
	repo := &fakeRepo{session: session}
	svc := &fakeService{proposes: true}
	return repo, svc, NewChatUseCase(repo, svc, testConfig(1_000_000))
}

func ctxAs(userId string) context.Context {
	return contextx.WithUserID(context.Background(), userId)
}

func drain(ch <-chan chat.Delta) []chat.Delta {
	var deltas []chat.Delta
	for delta := range ch {
		deltas = append(deltas, delta)
	}
	return deltas
}

func ownedSession() *chat.Session {
	return &chat.Session{Id: sessionId, UserId: owner, Title: "owned"}
}

// A session someone else owns and a session that does not exist have to be
// indistinguishable, or the response tells a caller which ids are real.
func TestSomebodyElsesSessionIsNotFound(t *testing.T) {
	for _, tt := range []struct {
		name string
		id   string
	}{
		{"another user's session", sessionId},
		{"a session that does not exist", missingId},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := contextx.WithUserID(context.Background(), intruder)

			t.Run("ListMessages", func(t *testing.T) {
				_, uc := newUseCase(ownedSession())
				_, err := uc.ListMessages(ctx, tt.id)
				assert.Equal(t, codes.NotFound, status.Code(err), "got %v", err)
			})

			t.Run("DeleteSession", func(t *testing.T) {
				repo, uc := newUseCase(ownedSession())
				err := uc.DeleteSession(ctx, tt.id)
				assert.Equal(t, codes.NotFound, status.Code(err), "got %v", err)
				assert.Empty(t, repo.deletes, "nothing may be deleted for a session the caller does not own")
			})

			t.Run("SendMessage", func(t *testing.T) {
				repo, uc := newUseCase(ownedSession())
				_, err := uc.SendMessage(ctx, tt.id, "hello")
				assert.Equal(t, codes.NotFound, status.Code(err), "got %v", err)
				assert.Empty(t, repo.messages, "no message may be stored against a session the caller does not own")
			})
		})
	}
}

// The use case must hand its own user id to every scoped query - passing the
// session id alone is what let a session be read by anyone who knew it.
func TestEveryLookupIsScopedToTheCaller(t *testing.T) {
	ctx := contextx.WithUserID(context.Background(), owner)

	repo, uc := newUseCase(ownedSession())
	_, err := uc.ListMessages(ctx, sessionId)
	require.NoError(t, err)
	assert.Equal(t, []call{{sessionId, owner}}, repo.finds)
	assert.Equal(t, []call{{sessionId, owner}}, repo.listMsgs)

	repo, uc = newUseCase(ownedSession())
	require.NoError(t, uc.DeleteSession(ctx, sessionId))
	assert.Equal(t, []call{{sessionId, owner}}, repo.deletes)
}

func TestOwnerReachesTheirOwnSession(t *testing.T) {
	ctx := contextx.WithUserID(context.Background(), owner)
	repo, uc := newUseCase(ownedSession())
	repo.messages = []*chat.Message{{Id: "m1", SessionId: sessionId, Role: chat.RoleUser, Content: "hi"}}

	messages, err := uc.ListMessages(ctx, sessionId)
	require.NoError(t, err)
	assert.Len(t, messages, 1)

	deltas, err := uc.SendMessage(ctx, sessionId, "hello")
	require.NoError(t, err)
	for range deltas {
	}
}

// Without the authentication interceptor there is no user, which is a 401 -
// not the 500 a bare error would have produced.
func TestMissingUserIsUnauthenticated(t *testing.T) {
	ctx := context.Background()
	_, uc := newUseCase(ownedSession())

	_, err := uc.ListSessions(ctx)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))

	_, err = uc.CreateSession(ctx, "t")
	assert.Equal(t, codes.Unauthenticated, status.Code(err))

	assert.Equal(t, codes.Unauthenticated, status.Code(uc.DeleteSession(ctx, sessionId)))

	_, err = uc.ListMessages(ctx, sessionId)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))

	_, err = uc.SendMessage(ctx, sessionId, "hello")
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

// The point of the whole flow: a change the assistant wants to make does not
// happen when it asks. The stream stops, a proposal is stored, and the backend
// is not resumed until somebody answers.
func TestAProposedChangeIsNotMadeUntilItIsApproved(t *testing.T) {
	repo, svc, uc := newProposingUseCase(ownedSession())

	deltas, err := uc.SendMessage(ctxAs(owner), sessionId, "add them to the group")
	require.NoError(t, err)

	frames := drain(deltas)
	require.Len(t, frames, 1, "the stream stops on the proposal")
	require.NotNil(t, frames[0].Proposal)
	assert.Equal(t, "add_users_to_group", frames[0].Proposal.ToolName)
	assert.Equal(t, `{"id":"a-group"}`, frames[0].Proposal.Arguments,
		"the user is shown the real arguments, not a summary")
	assert.False(t, frames[0].Done, "the answer is not finished, it is waiting")

	require.NotNil(t, repo.proposal)
	assert.Equal(t, chat.ProposalPending, repo.proposal.Status)
	assert.Equal(t, sessionId, repo.proposal.SessionId)
	assert.Zero(t, svc.resumed, "nothing ran: the backend was never resumed")
}

func TestApprovingRunsTheChange(t *testing.T) {
	repo, svc, uc := newProposingUseCase(ownedSession())
	proposed, err := uc.SendMessage(ctxAs(owner), sessionId, "add them to the group")
	require.NoError(t, err)
	drain(proposed)

	drain(mustStream(t, uc, owner, true))

	assert.Equal(t, 1, svc.resumed)
	assert.Equal(t, []bool{true}, svc.approved, "the backend is told it was approved")
	assert.Equal(t, chat.ProposalApproved, repo.proposal.Status)
}

func TestDecliningRunsNothing(t *testing.T) {
	repo, svc, uc := newProposingUseCase(ownedSession())
	proposed, err := uc.SendMessage(ctxAs(owner), sessionId, "add them to the group")
	require.NoError(t, err)
	drain(proposed)

	drain(mustStream(t, uc, owner, false))

	assert.Equal(t, []bool{false}, svc.approved,
		"the backend is told the user declined, and declines to run the tool")
	assert.Equal(t, chat.ProposalRejected, repo.proposal.Status)
}

// One approval means one change. Sending the same id twice must not run the
// tool a second time - a retry, a double click or a replayed request all look
// alike from here.
func TestAProposalIsOnlyAnsweredOnce(t *testing.T) {
	_, svc, uc := newProposingUseCase(ownedSession())
	proposed, err := uc.SendMessage(ctxAs(owner), sessionId, "add them to the group")
	require.NoError(t, err)
	drain(proposed)
	drain(mustStream(t, uc, owner, true))

	_, err = uc.ConfirmToolCall(ctxAs(owner), proposalId, true)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))
	assert.Equal(t, 1, svc.resumed, "the tool must not be run again")
}

// A proposal is answerable only by the person whose conversation produced it.
// Somebody else approving it would be an approval the owner never gave.
func TestOnlyTheOwnerCanAnswerAProposal(t *testing.T) {
	_, svc, uc := newProposingUseCase(ownedSession())
	proposed, err := uc.SendMessage(ctxAs(owner), sessionId, "add them to the group")
	require.NoError(t, err)
	drain(proposed)

	_, err = uc.ConfirmToolCall(ctxAs(intruder), proposalId, true)
	assert.Equal(t, codes.NotFound, status.Code(err),
		"somebody else's proposal is absent, not refused")
	assert.Zero(t, svc.resumed)
}

func TestAnsweringAnUnknownProposalIsNotFound(t *testing.T) {
	_, svc, uc := newProposingUseCase(ownedSession())
	_, err := uc.ConfirmToolCall(ctxAs(owner), missingId, true)
	assert.Equal(t, codes.NotFound, status.Code(err))
	assert.Zero(t, svc.resumed)
}

func mustStream(t *testing.T, uc ChatUseCase, userId string, approved bool) <-chan chat.Delta {
	t.Helper()
	deltas, err := uc.ConfirmToolCall(ctxAs(userId), proposalId, approved)
	require.NoError(t, err)
	return deltas
}

// A conversation that has spent its allowance stops answering. The check runs
// before the message is stored, so an unanswerable message does not join the
// history and get paid for again on the next turn.
func TestASessionThatHasSpentItsBudgetStopsAnswering(t *testing.T) {
	session := ownedSession()
	session.TokensUsed = 900
	repo := &fakeRepo{session: session}
	uc := NewChatUseCase(repo, &fakeService{}, testConfig(800))

	_, err := uc.SendMessage(ctxAs(owner), sessionId, "one more thing")

	assert.Equal(t, codes.ResourceExhausted, status.Code(err))
	assert.Empty(t, repo.messages, "the message was not stored")
}

func TestASessionWithinItsBudgetCarriesOn(t *testing.T) {
	session := ownedSession()
	session.TokensUsed = 100
	repo := &fakeRepo{session: session}
	uc := NewChatUseCase(repo, &fakeService{}, testConfig(800))

	deltas, err := uc.SendMessage(ctxAs(owner), sessionId, "one more thing")
	require.NoError(t, err)
	drain(deltas)

	assert.Len(t, repo.messages, 1)
}

// What a round cost is accounted for, and is not part of the answer: a frame
// reporting it must not reach the client as text.
func TestTokensAreAccountedForAndNotStreamed(t *testing.T) {
	repo := &fakeRepo{session: ownedSession()}
	svc := &fakeService{spends: 250}
	uc := NewChatUseCase(repo, svc, testConfig(1_000_000))

	deltas, err := uc.SendMessage(ctxAs(owner), sessionId, "hello")
	require.NoError(t, err)
	frames := drain(deltas)

	assert.Equal(t, int64(250), repo.tokens)
	for _, frame := range frames {
		assert.Zero(t, frame.Tokens, "an accounting frame reached the client")
	}
}
