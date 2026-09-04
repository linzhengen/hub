package ai

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/linzhengen/hub/server/internal/domain/ai/agent"
	"github.com/linzhengen/hub/server/internal/domain/auth"
	"github.com/linzhengen/hub/server/internal/domain/contextx"
	"github.com/linzhengen/hub/server/internal/domain/system/organization"
	"github.com/linzhengen/hub/server/internal/domain/trans"
	"github.com/linzhengen/hub/server/internal/domain/user"
	"github.com/linzhengen/hub/server/internal/infrastructure/oidc/admin"
	"github.com/linzhengen/hub/server/internal/usecase/pagination"
	"github.com/linzhengen/hub/server/internal/usecase/scope"
	"github.com/linzhengen/hub/server/pkg/logger"
)

var (
	errInvalidAgentName = status.Error(
		codes.InvalidArgument,
		"an agent name must be 3-64 characters of lowercase letters, digits and hyphens, "+
			"starting and ending with a letter or digit",
	)
	// errOrgOutOfReach refuses work in a tenant the caller holds no live grant
	// in. Enforce says whether they may register agents at all; this says where.
	errOrgOutOfReach = status.Error(
		codes.PermissionDenied,
		"the organization is not one you hold access in",
	)
	// errParentInAnotherOrg keeps a lineage inside one tenant. A chain that
	// crossed the boundary would be a route for one customer's agent to act
	// under another's, which is the disclosure the boundary exists for.
	errParentInAnotherOrg = status.Error(
		codes.InvalidArgument,
		"a parent agent must belong to the same organization as its sub-agent",
	)
	// errAgentHasChildren refuses to remove an agent that still owns others.
	// The delete path removes the Keycloak client and then the `users` row; a
	// child's row would go with the parent's while the child's own credential
	// kept working, which is the one state this whole feature is ordered to
	// avoid.
	errAgentHasChildren = status.Error(
		codes.FailedPrecondition,
		"the agent still owns sub-agents: remove them first, so no credential is left without a record",
	)
	errAgentNotFound = status.Error(codes.NotFound, "agent not found")
)

// Credentials are what an agent needs to authenticate, returned once.
//
// The secret is not stored by hub and cannot be read back: Keycloak can issue a
// new one but not show the current one again in every configuration, so
// write-once is the behaviour that holds everywhere. Losing it means rotating
// it, which is the right cost for a credential.
type Credentials struct {
	ClientId string
	Secret   string
}

// AgentUseCase registers the programs that act on the API.
//
// It does not authenticate them. An agent is a Keycloak client and gets its
// token through the client credentials grant exactly as a service account does;
// what this adds is that hub knows the client exists, which tenant it belongs
// to, what it is for, who is answerable for it, and which hub user it acts as.
type AgentUseCase interface {
	Create(ctx context.Context, a *agent.Agent) (*agent.Agent, Credentials, error)
	Get(ctx context.Context, id string) (*agent.Agent, error)
	List(ctx context.Context, params agent.ListParams) ([]*agent.Agent, int64, error)
	RotateSecret(ctx context.Context, id string) (Credentials, error)
	Delete(ctx context.Context, id string) error
}

func NewAgentUseCase(
	transRepo trans.Repository,
	agentRepo agent.Repository,
	userRepo user.Repository,
	orgRepo organization.Repository,
	authSvc auth.Service,
	oidcAdmin admin.Client,
) AgentUseCase {
	return &agentUseCase{
		transRepo: transRepo,
		agentRepo: agentRepo,
		userRepo:  userRepo,
		orgRepo:   orgRepo,
		authSvc:   authSvc,
		oidcAdmin: oidcAdmin,
	}
}

type agentUseCase struct {
	transRepo trans.Repository
	agentRepo agent.Repository
	userRepo  user.Repository
	orgRepo   organization.Repository
	authSvc   auth.Service
	oidcAdmin admin.Client
}

// Create registers an agent and returns its credentials once.
//
// Everything that can be refused is refused before Keycloak is touched, because
// the client is what carries a working credential: the ordering is what keeps a
// rejected request from leaving one behind.
//
// The client is then made first, because hub's rows are derived from what it
// returns - the id of the user the client acts as, above all. If storing those
// rows fails, the client is removed again: a client nobody has a record of is a
// working credential nobody can see.
func (uc agentUseCase) Create(
	ctx context.Context,
	a *agent.Agent,
) (*agent.Agent, Credentials, error) {
	creator, ok := contextx.GetUserID(ctx)
	if !ok {
		return nil, Credentials{}, errNoUser
	}
	if !agent.ValidName(a.Name) {
		return nil, Credentials{}, errInvalidAgentName
	}
	if err := uc.checkOrgReachable(ctx, a.OrgId); err != nil {
		return nil, Credentials{}, err
	}
	if err := uc.checkOrgExists(ctx, a.OrgId); err != nil {
		return nil, Credentials{}, err
	}
	if err := uc.checkParent(ctx, a.ParentAgentId, a.OrgId); err != nil {
		return nil, Credentials{}, err
	}

	clientId := agent.ClientIdFor(a.Name)
	created, err := uc.oidcAdmin.CreateServiceAccountClient(ctx, clientId, a.Description)
	if err != nil {
		return nil, Credentials{}, err
	}

	registered := agent.Factory(
		a.OrgId, a.Name, a.Description, clientId,
		created.KeycloakID, created.UserID, a.ParentAgentId, creator)

	if err := uc.transRepo.ExecTrans(ctx, func(ctx context.Context) error {
		// The agent gets a row in `users` like anybody else, so that groups,
		// roles, expiring grants and the audit log all work on it unchanged. It
		// is created here rather than left to the interceptor's provisioning: a
		// client credentials token carries no email, and `users.email` is
		// unique and not null, so the first agent would take the empty address
		// and the second would collide with it.
		if err := uc.userRepo.Create(ctx, &user.User{
			Id:        created.UserID,
			Username:  agent.UsernameFor(a.Name),
			Email:     agent.EmailFor(a.Name),
			Status:    user.Active,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}); err != nil {
			return err
		}
		return uc.agentRepo.Create(ctx, registered)
	}); err != nil {
		uc.removeClient(ctx, created.KeycloakID)
		return nil, Credentials{}, err
	}

	return registered, Credentials{ClientId: clientId, Secret: created.Secret}, nil
}

// Get reads one agent, and only if the caller can reach its tenant. A boundary
// that held for listings but not for a direct read would not be a boundary.
func (uc agentUseCase) Get(ctx context.Context, id string) (*agent.Agent, error) {
	found, err := uc.agentRepo.FindOne(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := uc.checkOrgReachable(ctx, found.OrgId); err != nil {
		// Saying "you cannot reach that organization" about an id the caller
		// guessed would confirm the agent exists, so an unreachable one reads
		// as absent.
		return nil, errAgentNotFound
	}
	return found, nil
}

// List reads the agents the caller can reach. Listing every agent would tell
// one customer which agents another runs.
func (uc agentUseCase) List(
	ctx context.Context,
	params agent.ListParams,
) ([]*agent.Agent, int64, error) {
	visible, err := scope.VisibleOrgs(ctx, uc.authSvc)
	if err != nil {
		return nil, 0, err
	}
	if visible.Empty() {
		return nil, 0, nil
	}
	if params.OrgId != "" && !scope.Reaches(visible, params.OrgId) {
		return nil, 0, nil
	}
	if !visible.All {
		params.OrgIds = visible.OrgIds
	}

	page := pagination.New(params.Limit, params.Offset)
	params.Limit = uint32(page.Limit())   //nolint:gosec // pagination bounds it
	params.Offset = uint32(page.Offset()) //nolint:gosec // pagination bounds it
	return uc.agentRepo.List(ctx, params)
}

// RotateSecret issues a new secret and invalidates the old one, which is the
// only remedy for one that leaked or was lost.
//
// The rotation is stamped on the row afterwards. hub never holds the secret, so
// when it was last issued is the only thing it can say about its age - and a
// credential whose age nobody can see is one nobody rotates.
func (uc agentUseCase) RotateSecret(ctx context.Context, id string) (Credentials, error) {
	found, err := uc.Get(ctx, id)
	if err != nil {
		return Credentials{}, err
	}

	secret, err := uc.oidcAdmin.RotateClientSecret(ctx, found.KeycloakId)
	if err != nil {
		return Credentials{}, err
	}
	if err := uc.agentRepo.RecordSecretRotation(ctx, found.Id); err != nil {
		// The secret has already been replaced in Keycloak, so failing here
		// would report a rotation as not having happened while the old
		// credential was already dead. The stamp is worth less than that.
		logger.Errorf("agent: rotated the secret for %s but could not record when: %v", found.Id, err)
	}
	return Credentials{ClientId: found.ClientId, Secret: secret}, nil
}

// Delete removes the agent's identity entirely.
//
// The Keycloak client goes first, because that is what actually stops the agent
// authenticating: dropping hub's rows while leaving the client would take the
// record away and leave the credential working, which is the worse of the two
// failures to be left with.
//
// Removing the user row cascades to its group memberships and to the agent row,
// so nothing is left holding permissions.
func (uc agentUseCase) Delete(ctx context.Context, id string) error {
	found, err := uc.Get(ctx, id)
	if err != nil {
		return err
	}

	children, err := uc.agentRepo.CountChildren(ctx, found.Id)
	if err != nil {
		return err
	}
	if children > 0 {
		return errAgentHasChildren
	}

	if err := uc.oidcAdmin.DeleteClient(ctx, found.KeycloakId); err != nil {
		return err
	}

	return uc.transRepo.ExecTrans(ctx, func(ctx context.Context) error {
		return uc.userRepo.Delete(ctx, found.UserId)
	})
}

// checkOrgReachable refuses a tenant the caller holds no live grant in.
func (uc agentUseCase) checkOrgReachable(ctx context.Context, orgId string) error {
	visible, err := scope.VisibleOrgs(ctx, uc.authSvc)
	if err != nil {
		return err
	}
	if !scope.Reaches(visible, orgId) {
		return errOrgOutOfReach
	}
	return nil
}

// checkOrgExists is only reached by a caller whose grants answer about every
// organization, for whom scope.Reaches is true of an id that is not an
// organization at all. Without it their typo becomes a foreign key violation
// after a Keycloak client has already been registered.
func (uc agentUseCase) checkOrgExists(ctx context.Context, orgId string) error {
	_, err := uc.orgRepo.FindOne(ctx, orgId)
	return err
}

// checkParent settles that the named parent exists and is in the same tenant.
// It is a lookup rather than a foreign key so that the answer arrives before a
// Keycloak client is registered for a request that was never going to be
// stored.
func (uc agentUseCase) checkParent(ctx context.Context, parentId, orgId string) error {
	if parentId == "" {
		return nil
	}
	parent, err := uc.agentRepo.FindOne(ctx, parentId)
	if err != nil {
		return err
	}
	if parent.OrgId != orgId {
		return errParentInAnotherOrg
	}
	return nil
}

// removeClient undoes a client that hub could not record. It is best effort:
// the caller is already returning an error, and there is nothing useful to do
// with a second one except say so.
func (uc agentUseCase) removeClient(ctx context.Context, keycloakId string) {
	if err := uc.oidcAdmin.DeleteClient(ctx, keycloakId); err != nil {
		logger.Errorf(
			"agent: could not remove keycloak client %s after failing to record it; "+
				"it exists in keycloak with no row in hub: %v",
			keycloakId, err)
	}
}
