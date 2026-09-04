package auth

import (
	"context"
	"database/sql"
	"time"

	"github.com/linzhengen/hub/server/internal/domain/auth"
	"github.com/linzhengen/hub/server/internal/domain/system/organization"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence"
)

// expiry turns the nullable column into the domain's "nil means never".
func expiry(v sql.NullTime) *time.Time {
	if !v.Valid {
		return nil
	}
	t := v.Time
	return &t
}

// orgWide decides, once and here, whether a grant reached through this kind of
// organization answers about every organization.
//
// The rule itself lives in organization.Kind.AppliesEverywhere. It is applied
// where the graph is read rather than carried into the decision as a kind,
// because a decision should only ever have to ask "does this reach here?" - and
// because auth would otherwise have to know what kinds of organization exist.
func orgWide(kind string) bool {
	return organization.Kind(kind).AppliesEverywhere()
}

// Repository is the implementation of the auth.Repository interface
type Repository struct {
	q persistence.Querier
}

// NewRepository creates a new auth repository
func NewRepository(q persistence.Querier) auth.Repository {
	return &Repository{
		q: q,
	}
}

func (r *Repository) FindUserAuthorizedPolicies(ctx context.Context, userId string) ([]auth.Policy, error) {
	rows, err := persistence.GetQ(ctx, r.q).SelectUserAuthorizedPolicies(ctx, userId)
	if err != nil {
		return nil, err
	}

	var policies []auth.Policy
	for _, row := range rows {
		policies = append(policies, auth.Policy{
			Subject:   row.ID,
			Object:    row.Identifier,
			Action:    row.Verb,
			ExpiresAt: auth.Earliest(expiry(row.GroupExpiresAt), expiry(row.RoleExpiresAt)),
			OrgId:     row.OrgID,
			OrgWide:   orgWide(row.OrgKind),
		})
	}

	return policies, nil
}

func (r *Repository) Revision(ctx context.Context) (int64, error) {
	return persistence.GetQ(ctx, r.q).SelectRbacRevision(ctx)
}

func (r *Repository) FindUserAccessPaths(ctx context.Context, userId string) ([]auth.AccessPath, error) {
	rows, err := persistence.GetQ(ctx, r.q).SelectUserAccessPath(ctx, userId)
	if err != nil {
		return nil, err
	}

	paths := make([]auth.AccessPath, 0, len(rows))
	for _, row := range rows {
		paths = append(paths, auth.AccessPath{
			GroupId:      row.GroupID,
			GroupName:    row.GroupName,
			RoleId:       row.RoleID,
			RoleName:     row.RoleName,
			PermissionId: row.PermissionID,
			Object:       row.Identifier,
			Action:       row.Verb,
			ExpiresAt:    auth.Earliest(expiry(row.GroupExpiresAt), expiry(row.RoleExpiresAt)),
			OrgId:        row.OrgID,
			OrgName:      row.OrgName,
			OrgWide:      orgWide(row.OrgKind),
		})
	}
	return paths, nil
}

func (r *Repository) FindAccessPaths(ctx context.Context) ([]auth.AccessPath, error) {
	rows, err := persistence.GetQ(ctx, r.q).SelectAccessPath(ctx)
	if err != nil {
		return nil, err
	}

	paths := make([]auth.AccessPath, 0, len(rows))
	for _, row := range rows {
		paths = append(paths, auth.AccessPath{
			GroupId:      row.GroupID,
			GroupName:    row.GroupName,
			RoleId:       row.RoleID,
			RoleName:     row.RoleName,
			PermissionId: row.PermissionID,
			Object:       row.Identifier,
			Action:       row.Verb,
			ExpiresAt:    expiry(row.ExpiresAt),
			OrgId:        row.OrgID,
			OrgName:      row.OrgName,
			OrgWide:      orgWide(row.OrgKind),
		})
	}
	return paths, nil
}

func (r *Repository) FindMemberships(ctx context.Context) ([]auth.Membership, error) {
	rows, err := persistence.GetQ(ctx, r.q).SelectMembership(ctx)
	if err != nil {
		return nil, err
	}

	memberships := make([]auth.Membership, 0, len(rows))
	for _, row := range rows {
		memberships = append(memberships, auth.Membership{
			UserId:    row.ID,
			Username:  row.Username,
			GroupId:   row.GroupID,
			ExpiresAt: expiry(row.ExpiresAt),
		})
	}
	return memberships, nil
}
