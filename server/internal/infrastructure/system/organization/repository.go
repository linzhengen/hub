package organization

import (
	"context"

	"github.com/linzhengen/hub/server/internal/domain/system/organization"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence"
)

func New(q persistence.Querier) organization.Repository {
	return &repositoryImpl{q: q}
}

type repositoryImpl struct {
	q persistence.Querier
}

func toDomain(m *persistence.OrganizationModel) *organization.Organization {
	return &organization.Organization{
		Id:          m.ID,
		Name:        m.Name,
		Slug:        m.Slug,
		Kind:        organization.Kind(m.Kind),
		Description: m.Description,
		Status:      organization.Status(m.Status),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func (r repositoryImpl) Create(ctx context.Context, o *organization.Organization) error {
	return persistence.GetQ(ctx, r.q).CreateOrganization(
		ctx, o.Id, o.Name, o.Slug, string(o.Kind), o.Description, string(o.Status),
	)
}

func (r repositoryImpl) FindOne(ctx context.Context, id string) (*organization.Organization, error) {
	m, err := persistence.GetQ(ctx, r.q).SelectOrganization(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomain(m), nil
}

func (r repositoryImpl) FindBySlug(ctx context.Context, slug string) (*organization.Organization, error) {
	m, err := persistence.GetQ(ctx, r.q).SelectOrganizationBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return toDomain(m), nil
}

func (r repositoryImpl) Update(ctx context.Context, o *organization.Organization) error {
	return persistence.GetQ(ctx, r.q).UpdateOrganization(
		ctx, o.Name, o.Slug, o.Description, string(o.Status), o.Id,
	)
}

func (r repositoryImpl) Delete(ctx context.Context, id string) error {
	return persistence.GetQ(ctx, r.q).DeleteOrganization(ctx, id)
}

// FindByUser is the membership question answered from the graph rather than
// from a membership table: belonging to an organization is having a group in
// it, which is the same edge an authorization decision reads.
func (r repositoryImpl) FindByUser(ctx context.Context, userId string) ([]*organization.Organization, error) {
	rows, err := persistence.GetQ(ctx, r.q).SelectUserOrganization(ctx, userId)
	if err != nil {
		return nil, err
	}
	items := make([]*organization.Organization, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomain(row))
	}
	return items, nil
}
