package system

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/linzhengen/hub/server/internal/domain/ai/agent"
	"github.com/linzhengen/hub/server/internal/domain/auth"
	"github.com/linzhengen/hub/server/internal/domain/contextx"
	"github.com/linzhengen/hub/server/internal/domain/system/delegation"
	"github.com/linzhengen/hub/server/internal/domain/trans"
	"github.com/linzhengen/hub/server/internal/usecase/pagination"
	"github.com/linzhengen/hub/server/internal/usecase/scope"
)

var (
	// errDelegationExpiryInThePast refuses a term that is already over. It is
	// not merely useless: a row that looks like a grant and never was is the
	// kind of thing somebody finds later and misreads.
	errDelegationExpiryInThePast = status.Error(
		codes.InvalidArgument,
		"a delegation must expire in the future",
	)
	// errAgentOutOfReach refuses an agent in a tenant the caller holds no live
	// grant in - including one that does not exist, which reads the same so
	// that an id cannot be probed.
	errAgentOutOfReach = status.Error(
		codes.NotFound,
		"agent not found",
	)
	errDelegationNotFound = status.Error(codes.NotFound, "delegation not found")
)

// DelegationUseCase records which agent may act on whose behalf.
//
// Every delegation here lends the **caller's own** authority. There is no path
// through which one person grants an agent another person's authority: the
// principal is taken from the context and never from the request, so it is not
// something the API can express rather than something it refuses. The other
// route into the table - an approved access request - is somebody agreeing on
// their own behalf too, only asynchronously.
type DelegationUseCase interface {
	Create(ctx context.Context, d *delegation.Delegation) (*delegation.Delegation, error)
	List(ctx context.Context, params delegation.ListParams) ([]*delegation.Delegation, int64, error)
	Revoke(ctx context.Context, id string) (*delegation.Delegation, error)
}

func NewDelegationUseCase(
	transRepo trans.Repository,
	delegationRepo delegation.Repository,
	agentRepo agent.Repository,
	authSvc auth.Service,
) DelegationUseCase {
	return &delegationUseCase{
		transRepo:      transRepo,
		delegationRepo: delegationRepo,
		agentRepo:      agentRepo,
		authSvc:        authSvc,
	}
}

type delegationUseCase struct {
	transRepo      trans.Repository
	delegationRepo delegation.Repository
	agentRepo      agent.Repository
	authSvc        auth.Service
}

// Create lends the caller's authority to an agent, for a set of permissions and
// a term.
//
// It deliberately does **not** check that the caller holds the permissions they
// name. That check belongs to the decision, where the effective authority is
// `agent ∩ principal ∩ delegation` computed at the moment it is needed - and
// stating it here as well would be a second place for it to live, which is how
// a later change ends up removing the one that mattered. Naming a permission
// you do not hold produces a delegation that allows nothing, not one that
// grants you something.
func (uc delegationUseCase) Create(
	ctx context.Context,
	d *delegation.Delegation,
) (*delegation.Delegation, error) {
	principal, ok := contextx.GetUserID(ctx)
	if !ok {
		return nil, errNoUser
	}
	if d.ExpiresAt == nil || !d.ExpiresAt.After(time.Now()) {
		return nil, errDelegationExpiryInThePast
	}

	// The agent has to be one the caller can see. Otherwise an id guessed from
	// somewhere else would let a person lend their authority into another
	// tenant, which is a route across the boundary rather than through it.
	target, err := uc.agentRepo.FindOne(ctx, d.AgentId)
	if err != nil {
		return nil, errAgentOutOfReach
	}
	if err := uc.reachable(ctx, target.OrgId); err != nil {
		return nil, err
	}

	depth := d.MaxDepth
	if depth == 0 {
		// A client that says nothing means the agent it named and no further,
		// which is the narrowest chain and therefore the right default.
		depth = 1
	}

	created := delegation.Factory(
		target.Id, principal, principal, target.OrgId, d.Reason,
		d.PermissionIds, depth, d.ExpiresAt)

	if err := uc.transRepo.ExecTrans(ctx, func(ctx context.Context) error {
		return uc.delegationRepo.Create(ctx, created)
	}); err != nil {
		return nil, err
	}
	return uc.delegationRepo.FindOne(ctx, created.Id)
}

// List reads the delegations the caller can reach: those in an organization
// they hold a grant in, plus their own wherever the agent lives.
//
// The second half matters. A person must be able to see, and therefore revoke,
// what they granted - and losing access to the agent's organization is exactly
// the moment they would most want to.
func (uc delegationUseCase) List(
	ctx context.Context,
	params delegation.ListParams,
) ([]*delegation.Delegation, int64, error) {
	caller, ok := contextx.GetUserID(ctx)
	if !ok {
		return nil, 0, errNoUser
	}
	visible, err := scope.VisibleOrgs(ctx, uc.authSvc)
	if err != nil {
		return nil, 0, err
	}
	if !visible.All {
		params.OrgIds = visible.OrgIds
	}
	params.SelfUserId = caller

	page := pagination.New(params.Limit, params.Offset)
	params.Limit = uint32(page.Limit())   //nolint:gosec // pagination bounds it
	params.Offset = uint32(page.Offset()) //nolint:gosec // pagination bounds it
	return uc.delegationRepo.List(ctx, params)
}

// Revoke withdraws a delegation. The principal may always do it; so may anyone
// who can reach the organization the agent belongs to, because an administrator
// noticing an agent misbehaving should not have to find the person who granted
// it first.
//
// Revoking twice is not an error. The caller wanted it gone and it is gone; the
// second call simply is not a second event, and the audit log records the
// attempt either way.
func (uc delegationUseCase) Revoke(
	ctx context.Context,
	id string,
) (*delegation.Delegation, error) {
	caller, ok := contextx.GetUserID(ctx)
	if !ok {
		return nil, errNoUser
	}

	existing, err := uc.delegationRepo.FindOne(ctx, id)
	if err != nil {
		return nil, errDelegationNotFound
	}
	if existing.PrincipalUserId != caller {
		if err := uc.reachable(ctx, existing.OrgId); err != nil {
			return nil, errDelegationNotFound
		}
	}

	if err := uc.transRepo.ExecTrans(ctx, func(ctx context.Context) error {
		_, err := uc.delegationRepo.Revoke(ctx, id)
		return err
	}); err != nil {
		return nil, err
	}
	return uc.delegationRepo.FindOne(ctx, id)
}

// reachable refuses an organization the caller holds no live grant in. It
// answers NotFound rather than PermissionDenied: whether a tenant exists is
// itself something the boundary is meant to withhold.
func (uc delegationUseCase) reachable(ctx context.Context, orgId string) error {
	visible, err := scope.VisibleOrgs(ctx, uc.authSvc)
	if err != nil {
		return err
	}
	if !scope.Reaches(visible, orgId) {
		return errAgentOutOfReach
	}
	return nil
}
