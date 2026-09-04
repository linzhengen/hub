package system

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/linzhengen/hub/server/internal/domain/access"
	"github.com/linzhengen/hub/server/internal/domain/auth"
	"github.com/linzhengen/hub/server/internal/domain/contextx"
	"github.com/linzhengen/hub/server/internal/domain/trans"
	"github.com/linzhengen/hub/server/internal/domain/user/usergroup"
	"github.com/linzhengen/hub/server/internal/usecase/pagination"
)

var (
	errNoUser = status.Error(codes.Unauthenticated, "unauthenticated")
	// errAlreadyDecided is what losing the claim looks like to a caller. It is
	// FailedPrecondition rather than NotFound: the request exists, it has just
	// been settled already.
	errAlreadyDecided = status.Error(codes.FailedPrecondition, "this request has already been decided")
	errNotRequester   = status.Error(codes.PermissionDenied, "only the requester can cancel a request")
	errSelfApproval   = status.Error(codes.PermissionDenied, "a request cannot be decided by the person who raised it")
)

// AccessRequestUseCase is the request-and-approval path into a group.
//
// It exists so that being put in a group can be asked for and agreed to, rather
// than only done. The grant itself is the same write an administrator makes
// directly; what this adds is a reason, a term, and somebody other than the
// asker saying yes.
type AccessRequestUseCase interface {
	Create(
		ctx context.Context,
		subjectUserId, groupId, reason string,
		requestedUntil *time.Time,
		origin access.Origin,
		sessionId string,
	) (*access.Request, error)
	List(ctx context.Context, params access.ListParams) ([]*access.Request, int64, error)
	Cancel(ctx context.Context, id string) (*access.Request, error)
	Decide(ctx context.Context, id string, approved bool, comment string) (*access.Request, error)
}

func NewAccessRequestUseCase(
	transRepo trans.Repository,
	requestRepo access.Repository,
	userGroupRepo usergroup.Repository,
	authSvc auth.Service,
) AccessRequestUseCase {
	return &accessRequestUseCase{
		transRepo:     transRepo,
		requestRepo:   requestRepo,
		userGroupRepo: userGroupRepo,
		authSvc:       authSvc,
	}
}

type accessRequestUseCase struct {
	transRepo     trans.Repository
	requestRepo   access.Repository
	userGroupRepo usergroup.Repository
	authSvc       auth.Service
}

func (uc accessRequestUseCase) Create(
	ctx context.Context,
	subjectUserId, groupId, reason string,
	requestedUntil *time.Time,
	origin access.Origin,
	sessionId string,
) (*access.Request, error) {
	requester, ok := contextx.GetUserID(ctx)
	if !ok {
		return nil, errNoUser
	}
	// The requester is taken from the caller, never from the request body: a
	// request that could name somebody else as its author would let one person
	// put a colleague's name on an ask they never made.
	if subjectUserId == "" {
		subjectUserId = requester
	}

	r := access.Factory(requester, subjectUserId, groupId, reason, requestedUntil, origin, sessionId)
	if err := uc.requestRepo.Create(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

func (uc accessRequestUseCase) List(
	ctx context.Context,
	params access.ListParams,
) ([]*access.Request, int64, error) {
	// A request names a user, a group and a reason somebody wrote, so an
	// unnarrowed queue would show one tenant what another is asking for.
	scope, err := visibleOrgs(ctx, uc.authSvc)
	if err != nil {
		return nil, 0, err
	}
	if scope.Empty() {
		return nil, 0, nil
	}
	if !scope.All {
		params.OrgIds = scope.OrgIds
	}

	page := pagination.New(params.Limit, params.Offset)
	params.Limit = uint32(page.Limit())   //nolint:gosec // pagination bounds it
	params.Offset = uint32(page.Offset()) //nolint:gosec // pagination bounds it
	return uc.requestRepo.List(ctx, params)
}

// Cancel withdraws a request. Only the person who raised it may, and only while
// it is still pending - a decided request is a record, not a draft.
func (uc accessRequestUseCase) Cancel(ctx context.Context, id string) (*access.Request, error) {
	caller, ok := contextx.GetUserID(ctx)
	if !ok {
		return nil, errNoUser
	}

	existing, err := uc.requestRepo.FindOne(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing.RequesterUserId != caller {
		return nil, errNotRequester
	}

	cancelled, ok, err := uc.requestRepo.Decide(ctx, id, access.Decision{
		Status:          access.StatusCancelled,
		DecidedByUserId: caller,
		DecidedAt:       time.Now(),
	})
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errAlreadyDecided
	}
	return cancelled, nil
}

// Decide approves or rejects a request, and on approval performs the grant.
//
// The claim and the grant are one transaction, so a request can never be
// recorded as approved without the access existing, nor the access exist
// without the record of who agreed to it.
func (uc accessRequestUseCase) Decide(
	ctx context.Context,
	id string,
	approved bool,
	comment string,
) (*access.Request, error) {
	decider, ok := contextx.GetUserID(ctx)
	if !ok {
		return nil, errNoUser
	}

	var decided *access.Request
	if err := uc.transRepo.ExecTrans(ctx, func(ctx context.Context) error {
		existing, err := uc.requestRepo.FindOne(ctx, id)
		if err != nil {
			return err
		}
		// Checked before the claim so that a self-approval leaves the request
		// pending for somebody else to decide, rather than burning it.
		if !existing.DecidableBy(decider) {
			return errSelfApproval
		}

		outcome := access.StatusRejected
		if approved {
			outcome = access.StatusApproved
		}

		claimed, ok, err := uc.requestRepo.Decide(ctx, id, access.Decision{
			Status:          outcome,
			DecidedByUserId: decider,
			DecidedAt:       time.Now(),
			Comment:         comment,
		})
		if err != nil {
			return err
		}
		if !ok {
			return errAlreadyDecided
		}

		if approved {
			// The term asked for becomes the membership's expiry, so approving
			// "for a week" grants a week rather than for good.
			if err := uc.userGroupRepo.AssignGroup(
				ctx,
				claimed.SubjectUserId,
				claimed.GroupId,
				claimed.RequestedUntil,
			); err != nil {
				return err
			}
		}

		decided = claimed
		return nil
	}); err != nil {
		return nil, err
	}
	return decided, nil
}
