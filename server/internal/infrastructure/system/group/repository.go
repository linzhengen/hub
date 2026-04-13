package group

import (
	"context"

	"github.com/linzhengen/hub/server/internal/domain/contextx"
	"github.com/linzhengen/hub/server/internal/domain/system/group"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence/mysql"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence/mysql/sqlc"
)

func New(q *sqlc.Queries) group.Repository {
	return &repositoryImpl{q: q}
}

type repositoryImpl struct {
	q *sqlc.Queries
}

func (r repositoryImpl) FindOne(ctx context.Context, id string) (*group.Group, error) {
	g, err := contextx.FindOne[sqlc.Group](ctx, id, mysql.GetQ(ctx, r.q).SelectGroupById, mysql.GetQ(ctx, r.q).SelectGroupForUpdate)
	if err != nil {
		return nil, err
	}

	return &group.Group{
		Id:          g.ID,
		Name:        g.Name,
		Description: g.Description,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}, nil
}

func (r repositoryImpl) Create(ctx context.Context, g *group.Group) error {
	return mysql.GetQ(ctx, r.q).CreateGroup(ctx, sqlc.CreateGroupParams{
		ID:          g.Id,
		Name:        g.Name,
		Status:      string(g.Status),
		Description: g.Description,
	})
}

func (r repositoryImpl) Update(ctx context.Context, g *group.Group) error {
	return mysql.GetQ(ctx, r.q).UpdateGroup(ctx, sqlc.UpdateGroupParams{
		Name:        g.Name,
		Description: g.Description,
		ID:          g.Id,
		Status:      string(g.Status),
	})
}

func (r repositoryImpl) Delete(ctx context.Context, id string) error {
	return mysql.GetQ(ctx, r.q).DeleteGroup(ctx, id)
}
