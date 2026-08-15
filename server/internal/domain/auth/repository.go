package auth

import "context"

type Repository interface {
	FindUserAuthorizedPolicies(ctx context.Context, userId string) ([]Policy, error)
	// Revision changes whenever anything an authorization decision reads
	// changes. It exists so a cache can ask "is what I hold still true?" in one
	// cheap query instead of repeating the permission join.
	Revision(ctx context.Context) (int64, error)

	// FindUserAccessPaths returns every route by which one user holds a
	// permission, unfiltered. It answers the same join as
	// FindUserAuthorizedPolicies but keeps the groups and roles that a policy
	// throws away.
	FindUserAccessPaths(ctx context.Context, userId string) ([]AccessPath, error)
	// FindAccessPaths returns every route in the graph, without the user
	// dimension: one row per group-role-permission edge, however many people
	// are in the group.
	//
	// Splitting the users off is what keeps the "who can do this" query from
	// returning users times permissions rows. The paths are narrowed first,
	// and only the groups that survive are looked up.
	FindAccessPaths(ctx context.Context) ([]AccessPath, error)
	// FindMemberships returns which active users are in which groups.
	FindMemberships(ctx context.Context) ([]Membership, error)
}
