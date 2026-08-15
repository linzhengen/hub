package system

import (
	"context"

	"github.com/linzhengen/hub/server/internal/domain/auth"
)

// AccessUseCase reads the authorization graph. It offers no write: the graph is
// edited through the group, role and permission use cases, and a second way in
// would be a second thing to audit.
type AccessUseCase interface {
	ExplainUserAccess(ctx context.Context, userId, resource, action string) ([]auth.AccessPath, error)
	PrincipalsForOperation(ctx context.Context, resource, action string) ([]auth.Principal, error)
}

func NewAccessUseCase(authSvc auth.Service) AccessUseCase {
	return &accessUseCase{authSvc: authSvc}
}

type accessUseCase struct {
	authSvc auth.Service
}

func (uc accessUseCase) ExplainUserAccess(
	ctx context.Context,
	userId, resource, action string,
) ([]auth.AccessPath, error) {
	return uc.authSvc.Explain(ctx, auth.Request{
		Subject: userId,
		Object:  resource,
		Action:  action,
	})
}

// PrincipalsForOperation asks about an operation rather than a person, so the
// request carries no subject.
func (uc accessUseCase) PrincipalsForOperation(
	ctx context.Context,
	resource, action string,
) ([]auth.Principal, error) {
	return uc.authSvc.PrincipalsFor(ctx, auth.Request{
		Object: resource,
		Action: action,
	})
}
