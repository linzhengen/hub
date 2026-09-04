package group

import (
	"context"

	"github.com/linzhengen/hub/server/internal/domain/contextx"
	"github.com/linzhengen/hub/server/internal/domain/system/group"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence"
)

func New(q persistence.Querier) group.Repository {
	return &repositoryImpl{q: q}
}

type repositoryImpl struct {
	q persistence.Querier
}

func (r repositoryImpl) FindOne(ctx context.Context, id string) (*group.Group, error) {
	q := persistence.GetQ(ctx, r.q)
	var g *persistence.GroupModel
	var err error
	if contextx.FromTransLock(ctx) {
		g, err = q.SelectGroupForUpdate(ctx, id)
	} else {
		g, err = q.SelectGroupById(ctx, id)
	}
	if err != nil {
		return nil, err
	}

	return &group.Group{
		Id:          g.ID,
		Name:        g.Name,
		Description: g.Description,
		Status:      group.Status(g.Status),
		OrgId:       g.OrgID,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}, nil
}

func (r repositoryImpl) Create(ctx context.Context, g *group.Group) error {
	return persistence.GetQ(ctx, r.q).CreateGroup(ctx, g.Id, g.Name, string(g.Status), g.Description, g.OrgId)
}

func (r repositoryImpl) Update(ctx context.Context, g *group.Group) error {
	return persistence.GetQ(ctx, r.q).UpdateGroup(ctx, g.Name, g.Description, string(g.Status), g.Id)
}

func (r repositoryImpl) Delete(ctx context.Context, id string) error {
	return persistence.GetQ(ctx, r.q).DeleteGroup(ctx, id)
}
