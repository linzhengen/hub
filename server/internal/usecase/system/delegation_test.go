package system

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	agentDomain "github.com/linzhengen/hub/server/internal/domain/ai/agent"
	"github.com/linzhengen/hub/server/internal/domain/auth"
	"github.com/linzhengen/hub/server/internal/domain/system/delegation"
)

const (
	principalId    = "11111111-1111-1111-1111-111111111111"
	otherUserId    = "22222222-2222-2222-2222-222222222222"
	delegatedAgent = "33333333-3333-3333-3333-333333333333"
	homeOrgId      = "44444444-4444-4444-4444-444444444444"
	foreignOrgId   = "55555555-5555-5555-5555-555555555555"
)

// fakeDelegationRepo records what was written.
type fakeDelegationRepo struct {
	stored    *delegation.Delegation
	createErr error
	findErr   error
	revoked   []string
	listed    delegation.ListParams
}

func (f *fakeDelegationRepo) Create(_ context.Context, d *delegation.Delegation) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.stored = d
	return nil
}

func (f *fakeDelegationRepo) FindOne(_ context.Context, _ string) (*delegation.Delegation, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	if f.stored == nil {
		return nil, errors.New("no rows")
	}
	copied := *f.stored
	return &copied, nil
}

func (f *fakeDelegationRepo) List(
	_ context.Context,
	params delegation.ListParams,
) ([]*delegation.Delegation, int64, error) {
	f.listed = params
	return nil, 0, nil
}

func (f *fakeDelegationRepo) Revoke(_ context.Context, id string) (bool, error) {
	f.revoked = append(f.revoked, id)
	return true, nil
}

// fakeAgentRepo answers where an agent lives, which is what the tenant boundary
// is checked against.
type fakeAgentRepo struct {
	agentDomain.Repository

	orgId   string
	findErr error
}

func (f *fakeAgentRepo) FindOne(_ context.Context, id string) (*agentDomain.Agent, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	return &agentDomain.Agent{Id: id, OrgId: f.orgId}, nil
}

// scopedAuth reports a caller who can reach exactly the organizations given.
type scopedAuth struct {
	auth.Service

	orgIds []string
	all    bool
}

func (s scopedAuth) VisibleOrgs(_ context.Context, _, _ string) (auth.Scope, error) {
	return auth.Scope{All: s.all, OrgIds: s.orgIds}, nil
}

func newDelegationUseCase(
	repo *fakeDelegationRepo,
	agents *fakeAgentRepo,
	authSvc auth.Service,
) DelegationUseCase {
	return NewDelegationUseCase(passthroughTrans{}, repo, agents, authSvc)
}

func inAnHour() *time.Time {
	t := time.Now().Add(time.Hour)
	return &t
}

func wanted() *delegation.Delegation {
	return &delegation.Delegation{
		AgentId:       delegatedAgent,
		Reason:        "runs the nightly report",
		PermissionIds: []string{"perm-a"},
		ExpiresAt:     inAnHour(),
	}
}

// TestCreate_LendsOnlyTheCallersOwnAuthority is the load-bearing test.
//
// The principal is taken from the context and never from the request, so
// granting an agent somebody else's authority is not something the API can
// express. If this ever regressed, an administrator could hand an agent the
// rights of any user in the system without that user knowing.
func TestCreate_LendsOnlyTheCallersOwnAuthority(t *testing.T) {
	repo := &fakeDelegationRepo{}
	uc := newDelegationUseCase(repo, &fakeAgentRepo{orgId: homeOrgId}, scopedAuth{orgIds: []string{homeOrgId}})

	// The request names another user's id in every field it could; none of them
	// reach the stored row.
	asked := wanted()
	asked.PrincipalUserId = otherUserId
	asked.GrantedByUserId = otherUserId

	created, err := uc.Create(ctxAs(principalId), asked)

	require.NoError(t, err)
	assert.Equal(t, principalId, created.PrincipalUserId,
		"the principal must be the caller, never a field of the request")
	assert.Equal(t, principalId, created.GrantedByUserId)
	require.NotNil(t, repo.stored)
	assert.Equal(t, principalId, repo.stored.PrincipalUserId)
}

// The organization is taken from the agent rather than the caller, because that
// is what the delegation is answerable in.
func TestCreate_TakesTheOrganizationFromTheAgent(t *testing.T) {
	repo := &fakeDelegationRepo{}
	uc := newDelegationUseCase(repo, &fakeAgentRepo{orgId: homeOrgId}, scopedAuth{all: true})

	created, err := uc.Create(ctxAs(principalId), wanted())

	require.NoError(t, err)
	assert.Equal(t, homeOrgId, created.OrgId)
}

// An agent in a tenant the caller cannot reach reads as absent. Otherwise a
// guessed id would let somebody lend their authority across the boundary.
func TestCreate_RefusesAnAgentInAnotherTenant(t *testing.T) {
	repo := &fakeDelegationRepo{}
	uc := newDelegationUseCase(repo, &fakeAgentRepo{orgId: foreignOrgId}, scopedAuth{orgIds: []string{homeOrgId}})

	_, err := uc.Create(ctxAs(principalId), wanted())

	assert.ErrorIs(t, err, errAgentOutOfReach)
	assert.Nil(t, repo.stored)
}

// A term that is already over is refused rather than stored: a row that looks
// like a grant and never was is what somebody finds later and misreads.
func TestCreate_RefusesATermAlreadyOver(t *testing.T) {
	past := time.Now().Add(-time.Minute)
	for _, tt := range []struct {
		name    string
		expires *time.Time
	}{
		{name: "in the past", expires: &past},
		{name: "absent", expires: nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeDelegationRepo{}
			uc := newDelegationUseCase(repo, &fakeAgentRepo{orgId: homeOrgId}, scopedAuth{all: true})

			asked := wanted()
			asked.ExpiresAt = tt.expires

			_, err := uc.Create(ctxAs(principalId), asked)

			assert.ErrorIs(t, err, errDelegationExpiryInThePast)
			assert.Nil(t, repo.stored)
		})
	}
}

// A client that says nothing about depth gets the narrowest chain, not the
// widest: the agent it named and no further.
func TestCreate_DefaultsToTheNarrowestChain(t *testing.T) {
	repo := &fakeDelegationRepo{}
	uc := newDelegationUseCase(repo, &fakeAgentRepo{orgId: homeOrgId}, scopedAuth{all: true})

	asked := wanted()
	asked.MaxDepth = 0

	created, err := uc.Create(ctxAs(principalId), asked)

	require.NoError(t, err)
	assert.Equal(t, uint32(1), created.MaxDepth)
}

func TestCreateDelegation_WithoutAUserIsRefused(t *testing.T) {
	repo := &fakeDelegationRepo{}
	uc := newDelegationUseCase(repo, &fakeAgentRepo{orgId: homeOrgId}, scopedAuth{all: true})

	_, err := uc.Create(context.Background(), wanted())

	assert.ErrorIs(t, err, errNoUser)
	assert.Nil(t, repo.stored)
}

func storedDelegation() *delegation.Delegation {
	return delegation.Factory(
		delegatedAgent, principalId, principalId, homeOrgId, "runs the nightly report",
		[]string{"perm-a"}, 1, inAnHour())
}

// The principal may always revoke, wherever the agent lives.
func TestRevoke_ByThePrincipal(t *testing.T) {
	repo := &fakeDelegationRepo{stored: storedDelegation()}
	uc := newDelegationUseCase(repo, &fakeAgentRepo{}, scopedAuth{orgIds: []string{foreignOrgId}})

	_, err := uc.Revoke(ctxAs(principalId), "any")

	require.NoError(t, err)
	assert.Equal(t, []string{"any"}, repo.revoked)
}

// So may somebody who can reach the agent's organization: an administrator
// noticing an agent misbehaving should not have to find the person who granted
// it first.
func TestRevoke_ByAnAdministratorOfTheAgentsTenant(t *testing.T) {
	repo := &fakeDelegationRepo{stored: storedDelegation()}
	uc := newDelegationUseCase(repo, &fakeAgentRepo{}, scopedAuth{orgIds: []string{homeOrgId}})

	_, err := uc.Revoke(ctxAs(otherUserId), "any")

	require.NoError(t, err)
	assert.Equal(t, []string{"any"}, repo.revoked)
}

// Anybody else is told it does not exist, rather than that they may not.
func TestRevoke_ByAStrangerIsRefused(t *testing.T) {
	repo := &fakeDelegationRepo{stored: storedDelegation()}
	uc := newDelegationUseCase(repo, &fakeAgentRepo{}, scopedAuth{orgIds: []string{foreignOrgId}})

	_, err := uc.Revoke(ctxAs(otherUserId), "any")

	assert.ErrorIs(t, err, errDelegationNotFound)
	assert.Empty(t, repo.revoked)
}

// TestList_AlwaysShowsTheCallersOwnDelegations covers the case a person most
// needs: losing access to the agent's organization must not hide, and therefore
// make unrevokable, what they themselves granted.
func TestList_AlwaysShowsTheCallersOwnDelegations(t *testing.T) {
	repo := &fakeDelegationRepo{}
	uc := newDelegationUseCase(repo, &fakeAgentRepo{}, scopedAuth{orgIds: []string{homeOrgId}})

	_, _, err := uc.List(ctxAs(principalId), delegation.ListParams{})

	require.NoError(t, err)
	assert.Equal(t, []string{homeOrgId}, repo.listed.OrgIds)
	assert.Equal(t, principalId, repo.listed.SelfUserId,
		"a person must be able to see, and so revoke, what they granted")
}

// A platform grant answers about every organization, so the listing is not
// narrowed: nil is every organization and is not an empty slice.
func TestList_DoesNotNarrowAPlatformGrant(t *testing.T) {
	repo := &fakeDelegationRepo{}
	uc := newDelegationUseCase(repo, &fakeAgentRepo{}, scopedAuth{all: true})

	_, _, err := uc.List(ctxAs(principalId), delegation.ListParams{})

	require.NoError(t, err)
	assert.Nil(t, repo.listed.OrgIds)
}

// Revoked delegations are left out unless asked for: the usual question is what
// may act for me now.
func TestList_LeavesOutRevokedByDefault(t *testing.T) {
	repo := &fakeDelegationRepo{}
	uc := newDelegationUseCase(repo, &fakeAgentRepo{}, scopedAuth{all: true})

	_, _, err := uc.List(ctxAs(principalId), delegation.ListParams{})

	require.NoError(t, err)
	assert.False(t, repo.listed.IncludeRevoked)
}
