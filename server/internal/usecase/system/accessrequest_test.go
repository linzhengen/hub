package system

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/linzhengen/hub/server/internal/domain/access"
	"github.com/linzhengen/hub/server/internal/domain/contextx"
	"github.com/linzhengen/hub/server/internal/domain/user/usergroup"
)

const (
	requesterID = "11111111-1111-1111-1111-111111111111"
	subjectID   = "22222222-2222-2222-2222-222222222222"
	approverID  = "33333333-3333-3333-3333-333333333333"
	groupID     = "44444444-4444-4444-4444-444444444444"
	requestID   = "55555555-5555-5555-5555-555555555555"
)

// fakeRequestRepo records what was asked of it and answers from what it was
// given. Decide follows the real one: it settles a pending request once, and
// reports the claim lost after that.
type fakeRequestRepo struct {
	stored    *access.Request
	created   *access.Request
	findErr   error
	decideErr error
	decisions []access.Decision
}

func (f *fakeRequestRepo) Create(_ context.Context, r *access.Request) error {
	f.created = r
	return nil
}

func (f *fakeRequestRepo) FindOne(_ context.Context, _ string) (*access.Request, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	copied := *f.stored
	return &copied, nil
}

func (f *fakeRequestRepo) List(_ context.Context, _ access.ListParams) ([]*access.Request, int64, error) {
	return nil, 0, nil
}

func (f *fakeRequestRepo) Decide(
	_ context.Context,
	_ string,
	d access.Decision,
) (*access.Request, bool, error) {
	if f.decideErr != nil {
		return nil, false, f.decideErr
	}
	f.decisions = append(f.decisions, d)
	if !f.stored.Pending() {
		return nil, false, nil
	}
	f.stored.Status = d.Status
	f.stored.DecidedByUserId = d.DecidedByUserId
	f.stored.DecidedAt = &d.DecidedAt
	f.stored.DecisionComment = d.Comment
	settled := *f.stored
	return &settled, true, nil
}

// fakeUserGroupRepo records the grants a decision performed.
type fakeUserGroupRepo struct {
	usergroup.Repository
	assigned []assignment
	err      error
}

type assignment struct {
	userId    string
	groupId   string
	expiresAt *time.Time
}

func (f *fakeUserGroupRepo) AssignGroup(
	_ context.Context,
	userId, groupId string,
	expiresAt *time.Time,
) error {
	if f.err != nil {
		return f.err
	}
	f.assigned = append(f.assigned, assignment{userId, groupId, expiresAt})
	return nil
}

// passthroughTrans runs the block inline. The transaction's job here is
// atomicity against the database, which the fakes do not model.
type passthroughTrans struct{}

func (passthroughTrans) ExecTrans(ctx context.Context, f func(context.Context) error) error {
	return f(ctx)
}

func (passthroughTrans) ExecTransWithLock(ctx context.Context, f func(context.Context) error) error {
	return f(ctx)
}

func pendingRequest(until *time.Time) *access.Request {
	return &access.Request{
		Id:              requestID,
		RequesterUserId: requesterID,
		SubjectUserId:   subjectID,
		GroupId:         groupID,
		Reason:          "on call this week",
		RequestedUntil:  until,
		Status:          access.StatusPending,
		Origin:          access.OriginConsole,
	}
}

func ctxAs(userId string) context.Context {
	return contextx.WithUserID(context.Background(), userId)
}

func newUseCase(repo *fakeRequestRepo, groups *fakeUserGroupRepo) AccessRequestUseCase {
	return NewAccessRequestUseCase(passthroughTrans{}, repo, groups)
}

// TestDecide_SelfApprovalIsRefused is the check the authorization interceptor
// cannot make: an administrator holds the permission to grant the group, so
// nothing before this point stops them approving their own ask.
func TestDecide_SelfApprovalIsRefused(t *testing.T) {
	repo := &fakeRequestRepo{stored: pendingRequest(nil)}
	groups := &fakeUserGroupRepo{}

	_, err := newUseCase(repo, groups).Decide(ctxAs(requesterID), requestID, true, "")

	assert.ErrorIs(t, err, errSelfApproval)
	assert.Empty(t, groups.assigned, "no grant may be made")
	// Refusing must not consume the request: somebody else still has to decide
	// it.
	assert.Empty(t, repo.decisions, "the request must be left pending")
	assert.True(t, repo.stored.Pending())
}

func TestDecide_ApprovalGrantsForTheTermAsked(t *testing.T) {
	friday := time.Now().Add(72 * time.Hour)
	repo := &fakeRequestRepo{stored: pendingRequest(&friday)}
	groups := &fakeUserGroupRepo{}

	decided, err := newUseCase(repo, groups).Decide(ctxAs(approverID), requestID, true, "ok for this week")

	require.NoError(t, err)
	assert.Equal(t, access.StatusApproved, decided.Status)
	assert.Equal(t, approverID, decided.DecidedByUserId)
	// The whole point of carrying the term: approving "until Friday" must not
	// grant for good.
	assert.Equal(t, []assignment{{subjectID, groupID, &friday}}, groups.assigned)
}

func TestDecide_RejectionGrantsNothing(t *testing.T) {
	repo := &fakeRequestRepo{stored: pendingRequest(nil)}
	groups := &fakeUserGroupRepo{}

	decided, err := newUseCase(repo, groups).Decide(ctxAs(approverID), requestID, false, "ask your manager")

	require.NoError(t, err)
	assert.Equal(t, access.StatusRejected, decided.Status)
	assert.Equal(t, "ask your manager", decided.DecisionComment)
	assert.Empty(t, groups.assigned)
}

// TestDecide_OneApprovalIsOneGrant covers the double click, the retry and the
// replayed request: the second decision claims no row, so it neither grants
// again nor rewrites who agreed.
func TestDecide_OneApprovalIsOneGrant(t *testing.T) {
	repo := &fakeRequestRepo{stored: pendingRequest(nil)}
	groups := &fakeUserGroupRepo{}
	uc := newUseCase(repo, groups)

	_, err := uc.Decide(ctxAs(approverID), requestID, true, "")
	require.NoError(t, err)

	_, err = uc.Decide(ctxAs(approverID), requestID, true, "")

	assert.ErrorIs(t, err, errAlreadyDecided)
	assert.Len(t, groups.assigned, 1, "the access must be granted exactly once")
}

func TestDecide_FailedGrantLeavesNothingApproved(t *testing.T) {
	repo := &fakeRequestRepo{stored: pendingRequest(nil)}
	groups := &fakeUserGroupRepo{err: errors.New("insert failed")}

	_, err := newUseCase(repo, groups).Decide(ctxAs(approverID), requestID, true, "")

	// The use case reports the failure; the transaction is what rolls the
	// claimed row back, so the caller never sees an approval without access.
	assert.Error(t, err)
	assert.Empty(t, groups.assigned)
}

func TestCancel_OnlyTheRequesterMay(t *testing.T) {
	repo := &fakeRequestRepo{stored: pendingRequest(nil)}
	groups := &fakeUserGroupRepo{}

	_, err := newUseCase(repo, groups).Cancel(ctxAs(approverID), requestID)

	assert.ErrorIs(t, err, errNotRequester)
	assert.True(t, repo.stored.Pending())
}

func TestCancel_ByTheRequester(t *testing.T) {
	repo := &fakeRequestRepo{stored: pendingRequest(nil)}
	groups := &fakeUserGroupRepo{}

	cancelled, err := newUseCase(repo, groups).Cancel(ctxAs(requesterID), requestID)

	require.NoError(t, err)
	assert.Equal(t, access.StatusCancelled, cancelled.Status)
	assert.Empty(t, groups.assigned)
}

// TestCreate_RequesterComesFromTheCaller stops one person putting a colleague's
// name on an ask they never made.
func TestCreate_RequesterComesFromTheCaller(t *testing.T) {
	repo := &fakeRequestRepo{}
	groups := &fakeUserGroupRepo{}

	created, err := newUseCase(repo, groups).Create(
		ctxAs(requesterID), subjectID, groupID, "on call", nil, access.OriginConsole, "")

	require.NoError(t, err)
	assert.Equal(t, requesterID, created.RequesterUserId)
	assert.Equal(t, subjectID, created.SubjectUserId)
	assert.Equal(t, access.StatusPending, created.Status)
}

// A request with no subject is a request for yourself, which is the common case
// from the console.
func TestCreate_SubjectDefaultsToTheRequester(t *testing.T) {
	repo := &fakeRequestRepo{}
	groups := &fakeUserGroupRepo{}

	created, err := newUseCase(repo, groups).Create(
		ctxAs(requesterID), "", groupID, "on call", nil, access.OriginConsole, "")

	require.NoError(t, err)
	assert.Equal(t, requesterID, created.SubjectUserId)
}

func TestCreate_WithoutAUserIsRefused(t *testing.T) {
	repo := &fakeRequestRepo{}

	_, err := newUseCase(repo, &fakeUserGroupRepo{}).Create(
		context.Background(), subjectID, groupID, "on call", nil, access.OriginConsole, "")

	assert.ErrorIs(t, err, errNoUser)
	assert.Nil(t, repo.created)
}

func TestDecidableBy(t *testing.T) {
	r := pendingRequest(nil)
	assert.False(t, r.DecidableBy(requesterID), "the person who asked may not decide")
	assert.True(t, r.DecidableBy(approverID))
}
