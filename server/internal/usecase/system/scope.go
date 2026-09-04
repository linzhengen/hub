package system

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/linzhengen/hub/server/internal/domain/auth"
	"github.com/linzhengen/hub/server/internal/domain/contextx"
)

// visibleOrgs is the organization filter every listing shares.
//
// Enforce decides whether a caller may read a list at all; this decides what is
// in it. Both come from the same policies, so a listing cannot show data from a
// tenant the caller could not act in - and a grant that lapses disappears from
// both at the same moment.
//
// It is a free function rather than a method because six listings need it and
// none of them should each grow their own idea of what "visible" means.
func visibleOrgs(ctx context.Context, authSvc auth.Service) (auth.Scope, error) {
	userId, ok := contextx.GetUserID(ctx)
	if !ok {
		return auth.Scope{}, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	return authSvc.VisibleOrgs(ctx, userId, contextx.GetOrgID(ctx))
}
