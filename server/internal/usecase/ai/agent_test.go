package ai

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/linzhengen/hub/server/internal/domain/ai/agent"
	"github.com/linzhengen/hub/server/internal/domain/auth"
	"github.com/linzhengen/hub/server/internal/domain/system/organization"
	"github.com/linzhengen/hub/server/internal/domain/user"
	"github.com/linzhengen/hub/server/internal/infrastructure/oidc/admin"
)

const (
	creatorId    = "66666666-6666-6666-6666-666666666666"
	agentOrgId   = "11111111-1111-1111-1111-111111111111"
	otherOrgId   = "22222222-2222-2222-2222-222222222222"
	agentKcId    = "kc-client-uuid"
	agentUserId  = "77777777-7777-7777-7777-777777777777"
	issuedSecret = "s3cr3t-value"
)

// passthroughTrans runs the block inline. The transaction's job here is
// atomicity against the database, which the fakes do not have.
type passthroughTrans struct{}

func (passthroughTrans) ExecTrans(ctx context.Context, f func(context.Context) error) error {
	return f(ctx)
}

func (passthroughTrans) ExecTransWithLock(ctx context.Context, f func(context.Context) error) error {
	return f(ctx)
}

// fakeOIDCAdmin embeds the real interface so only the three calls an agent
// makes have to be written. Anything else would panic, which is the right
// outcome: this use case has no business touching the rest.
type fakeOIDCAdmin struct {
	admin.Client

	created       admin.ServiceAccountClient
	createErr     error
	createdFor    string
	rotated       string
	rotateErr     error
	deletedClient []string
	deleteErr     error
}

func (f *fakeOIDCAdmin) CreateServiceAccountClient(
	_ context.Context,
	clientId, _ string,
) (admin.ServiceAccountClient, error) {
	f.createdFor = clientId
	if f.createErr != nil {
		return admin.ServiceAccountClient{}, f.createErr
	}
	return f.created, nil
}

func (f *fakeOIDCAdmin) RotateClientSecret(_ context.Context, id string) (string, error) {
	if f.rotateErr != nil {
		return "", f.rotateErr
	}
	f.rotated = id
	return "rotated-value", nil
}

func (f *fakeOIDCAdmin) DeleteClient(_ context.Context, id string) error {
	f.deletedClient = append(f.deletedClient, id)
	return f.deleteErr
}

// fakeAgentRepo records what was stored.
type fakeAgentRepo struct {
	stored    *agent.Agent
	byId      map[string]*agent.Agent
	createErr error
	children  int64
	rotations []string
	deleted   []string
	listed    agent.ListParams
}

func (f *fakeAgentRepo) Create(_ context.Context, a *agent.Agent) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.stored = a
	return nil
}

func (f *fakeAgentRepo) FindOne(_ context.Context, id string) (*agent.Agent, error) {
	if found, ok := f.byId[id]; ok {
		copied := *found
		return &copied, nil
	}
	if f.stored == nil {
		return nil, errors.New("no rows")
	}
	copied := *f.stored
	return &copied, nil
}

func (f *fakeAgentRepo) List(
	_ context.Context,
	params agent.ListParams,
) ([]*agent.Agent, int64, error) {
	f.listed = params
	return nil, 0, nil
}

func (f *fakeAgentRepo) CountChildren(_ context.Context, _ string) (int64, error) {
	return f.children, nil
}

func (f *fakeAgentRepo) RecordSecretRotation(_ context.Context, id string) error {
	f.rotations = append(f.rotations, id)
	return nil
}

func (f *fakeAgentRepo) Delete(_ context.Context, id string) error {
	f.deleted = append(f.deleted, id)
	return nil
}

// fakeUserRepo records the `users` row a registration writes, which is the row
// the whole permission model hangs off.
type fakeUserRepo struct {
	user.Repository

	created   *user.User
	createErr error
	deleted   []string
}

func (f *fakeUserRepo) Create(_ context.Context, u *user.User) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = u
	return nil
}

func (f *fakeUserRepo) Delete(_ context.Context, id string) error {
	f.deleted = append(f.deleted, id)
	return nil
}

// fakeOrgRepo answers whether an organization exists.
type fakeOrgRepo struct {
	organization.Repository

	missing bool
}

func (f *fakeOrgRepo) FindOne(_ context.Context, id string) (*organization.Organization, error) {
	if f.missing {
		return nil, errors.New("no rows")
	}
	return &organization.Organization{Id: id}, nil
}

// fakeAuth answers what a caller can see. It is the tenant boundary, so the
// tests set it deliberately rather than leaving it wide open.
type fakeAuth struct {
	auth.Service

	scope auth.Scope
	err   error
}

func (f fakeAuth) VisibleOrgs(_ context.Context, _, _ string) (auth.Scope, error) {
	return f.scope, f.err
}

func inOrg(ids ...string) fakeAuth {
	return fakeAuth{scope: auth.Scope{OrgIds: ids}}
}

func registeredClient() admin.ServiceAccountClient {
	return admin.ServiceAccountClient{
		KeycloakID: agentKcId,
		UserID:     agentUserId,
		Secret:     issuedSecret,
	}
}

type deps struct {
	agents *fakeAgentRepo
	users  *fakeUserRepo
	orgs   *fakeOrgRepo
	oidc   *fakeOIDCAdmin
	authz  fakeAuth
}

func newDeps() *deps {
	return &deps{
		agents: &fakeAgentRepo{},
		users:  &fakeUserRepo{},
		orgs:   &fakeOrgRepo{},
		oidc:   &fakeOIDCAdmin{created: registeredClient()},
		authz:  inOrg(agentOrgId),
	}
}

func (d *deps) useCase() AgentUseCase {
	return NewAgentUseCase(passthroughTrans{}, d.agents, d.users, d.orgs, d.authz, d.oidc)
}

func newAgent() *agent.Agent {
	return &agent.Agent{OrgId: agentOrgId, Name: "support-triage", Description: "triages support mail"}
}

// TestCreate_WritesTheUserRowKeycloakWillAuthenticateAs is the load-bearing
// test of the whole feature.
//
// An agent authenticates with a client credentials token whose `sub` is the
// Keycloak client's own service account user. The authorization interceptor
// looks the caller up by exactly that id, so unless the `users` row carries it,
// the agent authenticates as nobody: it joins no groups, holds no roles, and
// appears in the audit log as an id nothing matches.
//
// Everything else about an agent - groups, roles, expiring grants, audit, and
// the delegation chain that comes later - is the ordinary user machinery, and
// it all hangs off this one field.
func TestCreate_WritesTheUserRowKeycloakWillAuthenticateAs(t *testing.T) {
	d := newDeps()

	created, credentials, err := d.useCase().Create(ctxAs(creatorId), newAgent())

	require.NoError(t, err)

	require.NotNil(t, d.users.created)
	assert.Equal(t, agentUserId, d.users.created.Id,
		"the users row must carry the id Keycloak puts in the token's subject")
	assert.Equal(t, "service-account-hub-agent-support-triage", d.users.created.Username)
	assert.Equal(t, "hub-agent-support-triage@service-account.invalid", d.users.created.Email)
	assert.Equal(t, user.Active, d.users.created.Status)

	require.NotNil(t, d.agents.stored)
	assert.Equal(t, agentUserId, d.agents.stored.UserId)
	assert.Equal(t, agentKcId, d.agents.stored.KeycloakId)
	assert.Equal(t, agentOrgId, d.agents.stored.OrgId)
	assert.Equal(t, creatorId, d.agents.stored.CreatedByUserId,
		"an agent with no identifiable controller is the one nobody turns off")
	assert.Equal(t, "hub-agent-support-triage", d.agents.stored.ClientId)
	assert.Equal(t, agentUserId, created.UserId)

	// The secret comes back exactly here and nowhere else.
	assert.Equal(t, "hub-agent-support-triage", credentials.ClientId)
	assert.Equal(t, issuedSecret, credentials.Secret)
}

// The name reaches Keycloak as a client id, so a bad one is refused before the
// round trip rather than after it.
func TestCreate_RefusesABadNameWithoutTouchingKeycloak(t *testing.T) {
	d := newDeps()
	bad := newAgent()
	bad.Name = "Support Triage"

	_, _, err := d.useCase().Create(ctxAs(creatorId), bad)

	assert.ErrorIs(t, err, errInvalidAgentName)
	assert.Empty(t, d.oidc.createdFor, "no client may be registered for a name that was refused")
	assert.Nil(t, d.users.created)
	assert.Nil(t, d.agents.stored)
}

func TestCreate_WithoutAUserIsRefused(t *testing.T) {
	d := newDeps()

	_, _, err := d.useCase().Create(context.Background(), newAgent())

	assert.ErrorIs(t, err, errNoUser)
	assert.Empty(t, d.oidc.createdFor)
}

// TestCreate_RefusesAnOrganizationTheCallerCannotReach is the tenant boundary.
//
// Enforce says whether the caller may register agents at all; it does not say
// where. Without this check, anybody allowed to create an agent could create
// one inside another customer's organization - a working credential, holding
// that tenant's grants, that the tenant never asked for.
func TestCreate_RefusesAnOrganizationTheCallerCannotReach(t *testing.T) {
	d := newDeps()
	d.authz = inOrg(otherOrgId)

	_, _, err := d.useCase().Create(ctxAs(creatorId), newAgent())

	assert.ErrorIs(t, err, errOrgOutOfReach)
	assert.Empty(t, d.oidc.createdFor, "no credential may be minted for a tenant that was refused")
}

// A platform grant answers about every organization, so the reachability check
// cannot catch an id that is not an organization at all. Without the existence
// check that becomes a foreign key violation after a client has been made.
func TestCreate_RefusesAnOrganizationThatDoesNotExist(t *testing.T) {
	d := newDeps()
	d.authz = fakeAuth{scope: auth.Scope{All: true}}
	d.orgs.missing = true

	_, _, err := d.useCase().Create(ctxAs(creatorId), newAgent())

	assert.Error(t, err)
	assert.Empty(t, d.oidc.createdFor)
}

// A lineage that crossed the boundary would be a route for one customer's agent
// to act under another's, which is the disclosure the boundary is for.
func TestCreate_RefusesAParentInAnotherOrganization(t *testing.T) {
	const parentId = "33333333-3333-3333-3333-333333333333"
	d := newDeps()
	d.agents.byId = map[string]*agent.Agent{
		parentId: {Id: parentId, OrgId: otherOrgId},
	}
	sub := newAgent()
	sub.ParentAgentId = parentId

	_, _, err := d.useCase().Create(ctxAs(creatorId), sub)

	assert.ErrorIs(t, err, errParentInAnotherOrg)
	assert.Empty(t, d.oidc.createdFor)
}

func TestCreate_RecordsLineageForASubAgent(t *testing.T) {
	const parentId = "33333333-3333-3333-3333-333333333333"
	d := newDeps()
	d.agents.byId = map[string]*agent.Agent{
		parentId: {Id: parentId, OrgId: agentOrgId},
	}
	sub := newAgent()
	sub.ParentAgentId = parentId

	created, _, err := d.useCase().Create(ctxAs(creatorId), sub)

	require.NoError(t, err)
	assert.Equal(t, parentId, created.ParentAgentId)
	assert.False(t, created.Root())
}

// TestCreate_RemovesTheClientWhenItCannotBeRecorded covers the failure that
// would otherwise be the worst of the two: a Keycloak client that works, with
// no row in hub saying it exists. Nobody could find it to turn it off.
func TestCreate_RemovesTheClientWhenItCannotBeRecorded(t *testing.T) {
	for _, tt := range []struct {
		name  string
		setup func(*deps)
	}{
		{
			name:  "the users row fails",
			setup: func(d *deps) { d.users.createErr = errors.New("insert failed") },
		},
		{
			name:  "the agent row fails",
			setup: func(d *deps) { d.agents.createErr = errors.New("insert failed") },
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			d := newDeps()
			tt.setup(d)

			_, _, err := d.useCase().Create(ctxAs(creatorId), newAgent())

			assert.Error(t, err)
			assert.Equal(t, []string{agentKcId}, d.oidc.deletedClient,
				"a client hub could not record must not be left working")
		})
	}
}

// TestDelete_TakesTheCredentialBeforeTheRecord fixes the order.
//
// Deleting the Keycloak client is what actually stops the agent
// authenticating. Dropping hub's rows first would, on a failure, leave a
// working credential with no record of it - the same state the create path goes
// out of its way to avoid.
func TestDelete_TakesTheCredentialBeforeTheRecord(t *testing.T) {
	d := newDeps()
	d.agents.stored = storedAgent()
	d.oidc.deleteErr = errors.New("keycloak unreachable")

	err := d.useCase().Delete(ctxAs(creatorId), "any")

	assert.Error(t, err)
	assert.Empty(t, d.users.deleted,
		"hub's rows must survive a failure to remove the credential")
}

// Removing the user row is what removes the access: the group memberships and
// the agent row both cascade from it.
func TestDelete_RemovesTheUserTheAgentActedAs(t *testing.T) {
	d := newDeps()
	d.agents.stored = storedAgent()

	err := d.useCase().Delete(ctxAs(creatorId), "any")

	require.NoError(t, err)
	assert.Equal(t, []string{agentKcId}, d.oidc.deletedClient)
	assert.Equal(t, []string{agentUserId}, d.users.deleted)
}

// TestDelete_RefusesAnAgentThatStillOwnsSubAgents keeps the delete from
// orphaning credentials.
//
// The delete path removes the Keycloak client and then the `users` row, whose
// cascade takes the agent row. A child's row would go with the parent's while
// the child's own Keycloak client kept working - a live credential with nothing
// in hub recording it, which is the exact state everything else here is ordered
// to avoid.
func TestDelete_RefusesAnAgentThatStillOwnsSubAgents(t *testing.T) {
	d := newDeps()
	d.agents.stored = storedAgent()
	d.agents.children = 2

	err := d.useCase().Delete(ctxAs(creatorId), "any")

	assert.ErrorIs(t, err, errAgentHasChildren)
	assert.Empty(t, d.oidc.deletedClient, "nothing may be removed until the children are")
	assert.Empty(t, d.users.deleted)
}

// An agent in a tenant the caller cannot reach reads as absent rather than as
// forbidden: saying "you cannot reach that organization" about a guessed id
// would confirm the agent exists.
func TestGet_HidesAnAgentInAnotherTenant(t *testing.T) {
	d := newDeps()
	d.agents.stored = storedAgent()
	d.authz = inOrg(otherOrgId)

	_, err := d.useCase().Get(ctxAs(creatorId), "any")

	assert.ErrorIs(t, err, errAgentNotFound)
}

// The boundary has to hold for every path that reads an agent, not only the
// listing - and rotating a secret is a read followed by a credential change.
func TestRotateSecret_HonoursTheTenantBoundary(t *testing.T) {
	d := newDeps()
	d.agents.stored = storedAgent()
	d.authz = inOrg(otherOrgId)

	_, err := d.useCase().RotateSecret(ctxAs(creatorId), "any")

	assert.ErrorIs(t, err, errAgentNotFound)
	assert.Empty(t, d.oidc.rotated, "no secret may be replaced in a tenant the caller cannot reach")
}

func TestRotateSecret_IssuesAgainstTheStoredClientAndStampsIt(t *testing.T) {
	d := newDeps()
	d.agents.stored = storedAgent()

	credentials, err := d.useCase().RotateSecret(ctxAs(creatorId), "any")

	require.NoError(t, err)
	assert.Equal(t, agentKcId, d.oidc.rotated)
	// The client id is unchanged by a rotation, and is returned alongside so the
	// operator has the whole pair to paste in.
	assert.Equal(t, "hub-agent-support-triage", credentials.ClientId)
	assert.Equal(t, "rotated-value", credentials.Secret)
	// hub never holds the secret, so when it was issued is the only thing it can
	// report about its age - and an age nobody can see is one nobody acts on.
	assert.Equal(t, []string{d.agents.stored.Id}, d.agents.rotations)
}

// TestList_NarrowsToTheOrganizationsTheCallerHolds keeps one tenant from
// learning which agents another runs.
func TestList_NarrowsToTheOrganizationsTheCallerHolds(t *testing.T) {
	d := newDeps()

	_, _, err := d.useCase().List(ctxAs(creatorId), agent.ListParams{})

	require.NoError(t, err)
	assert.Equal(t, []string{agentOrgId}, d.agents.listed.OrgIds)
}

// A platform grant answers about every organization, so the listing is not
// narrowed at all - nil is every organization, and is not the same as an empty
// slice, which admits nothing.
func TestList_DoesNotNarrowAPlatformGrant(t *testing.T) {
	d := newDeps()
	d.authz = fakeAuth{scope: auth.Scope{All: true}}

	_, _, err := d.useCase().List(ctxAs(creatorId), agent.ListParams{})

	require.NoError(t, err)
	assert.Nil(t, d.agents.listed.OrgIds)
}

// Naming an organization the caller cannot reach returns nothing rather than
// everything they can reach: the filter must not widen the answer.
func TestList_ReturnsNothingForAnUnreachableOrganization(t *testing.T) {
	d := newDeps()

	agents, total, err := d.useCase().List(ctxAs(creatorId), agent.ListParams{OrgId: otherOrgId})

	require.NoError(t, err)
	assert.Empty(t, agents)
	assert.Zero(t, total)
}

// A caller holding no live grant anywhere sees no agents, rather than all of
// them through an unfiltered query.
func TestList_ReturnsNothingForAnEmptyScope(t *testing.T) {
	d := newDeps()
	d.authz = fakeAuth{}

	agents, total, err := d.useCase().List(ctxAs(creatorId), agent.ListParams{})

	require.NoError(t, err)
	assert.Empty(t, agents)
	assert.Zero(t, total)
}

func storedAgent() *agent.Agent {
	return agent.Factory(
		agentOrgId, "support-triage", "", "hub-agent-support-triage",
		agentKcId, agentUserId, "", creatorId)
}
