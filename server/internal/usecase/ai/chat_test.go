package ai

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/linzhengen/hub/server/internal/domain/ai/chat"
	"github.com/linzhengen/hub/server/internal/domain/contextx"
)

const (
	owner     = "11111111-1111-1111-1111-111111111111"
	intruder  = "22222222-2222-2222-2222-222222222222"
	sessionId = "33333333-3333-3333-3333-333333333333"
	missingId = "44444444-4444-4444-4444-444444444444"
)

type call struct{ id, userId string }

// fakeRepo mirrors what the queries now do: every lookup is scoped by user, so
// a session belonging to somebody else is not found rather than found and
// refused. A fake that ignored userId would let the use case pass this suite
// while the real repository leaked, which is the whole point of the change.
type fakeRepo struct {
	session  *chat.Session
	messages []*chat.Message

	finds    []call
	deletes  []call
	listMsgs []call
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

type fakeService struct{}

func (fakeService) Send(_ context.Context, _ []*chat.Message) (<-chan chat.Delta, error) {
	ch := make(chan chat.Delta, 1)
	ch <- chat.Delta{Done: true}
	close(ch)
	return ch, nil
}

func newUseCase(session *chat.Session) (*fakeRepo, ChatUseCase) {
	repo := &fakeRepo{session: session}
	return repo, NewChatUseCase(repo, fakeService{})
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
