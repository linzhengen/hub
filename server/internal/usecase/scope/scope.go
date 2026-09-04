// Package scope answers one question - which organizations may this caller see
// - and is the only place that answers it.
//
// Enforce decides whether a caller may read a list at all; this decides what is
// in it. Both come from the same policies, so a listing cannot show data from a
// tenant the caller could not act in, and a grant that lapses disappears from
// both at the same moment.
//
// It lives in a package of its own because listings in more than one feature
// need it, and none of them should grow its own idea of what "visible" means.
package scope

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/linzhengen/hub/server/internal/domain/auth"
	"github.com/linzhengen/hub/server/internal/domain/contextx"
)

// VisibleOrgs is the organization filter every listing shares.
func VisibleOrgs(ctx context.Context, authSvc auth.Service) (auth.Scope, error) {
	userId, ok := contextx.GetUserID(ctx)
	if !ok {
		return auth.Scope{}, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	return authSvc.VisibleOrgs(ctx, userId, contextx.GetOrgID(ctx))
}

// Reaches reports whether one organization is inside a scope.
//
// A listing narrows itself with the scope; a request that names a single
// organization has to be checked against it instead, and that check has to mean
// the same thing. Creating an agent in a tenant the caller cannot reach would
// otherwise mint a working credential inside somebody else's boundary.
func Reaches(s auth.Scope, orgId string) bool {
	if s.All {
		return true
	}
	for _, id := range s.OrgIds {
		if id == orgId {
			return true
		}
	}
	return false
}
