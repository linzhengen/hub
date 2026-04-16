package permission

import (
	"context"

	"github.com/linzhengen/hub/v1/server/internal/domain/contextx"
	"github.com/linzhengen/hub/v1/server/internal/domain/system/permission"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence/mysql"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence/mysql/sqlc"
)

func New(q *sqlc.Queries) permission.Repository {
	return &repositoryImpl{q: q}
}

type repositoryImpl struct {
	q *sqlc.Queries
}

func (r repositoryImpl) FindOne(ctx context.Context, id string) (*permission.Permission, error) {
	p, err := contextx.FindOne[sqlc.Permission](ctx, id, mysql.GetQ(ctx, r.q).SelectPermissionById, mysql.GetQ(ctx, r.q).SelectPermissionForUpdate)
	if err != nil {
		return nil, err
	}
	return &permission.Permission{
		Id:          p.ID,
		Verb:        permission.Verb(p.Verb),
		ResourceId:  p.ResourceID,
		Description: p.Description,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}, nil
}

func (r repositoryImpl) FindByResourceId(ctx context.Context, resourceId string) ([]*permission.Permission, error) {
	ps, err := mysql.GetQ(ctx, r.q).SelectPermissionByResourceId(ctx, resourceId)
	if err != nil {
		return nil, err
	}
	var result []*permission.Permission
	for _, p := range ps {
		result = append(result, &permission.Permission{
			Id:          p.ID,
			Verb:        permission.Verb(p.Verb),
			ResourceId:  p.ResourceID,
			Description: p.Description,
			CreatedAt:   p.CreatedAt,
		})
	}
	return result, nil
}

func (r repositoryImpl) Create(ctx context.Context, p *permission.Permission) error {
	return mysql.GetQ(ctx, r.q).CreatePermission(ctx, sqlc.CreatePermissionParams{
		ID:          p.Id,
		Verb:        string(p.Verb),
		ResourceID:  p.ResourceId,
		Description: p.Description,
	})
}

func (r repositoryImpl) Update(ctx context.Context, p *permission.Permission) error {
	return mysql.GetQ(ctx, r.q).UpdatePermission(ctx, sqlc.UpdatePermissionParams{
		Verb:        string(p.Verb),
		ResourceID:  p.ResourceId,
		Description: p.Description,
		ID:          p.Id,
	})
}

func (r repositoryImpl) Delete(ctx context.Context, id string) error {
	return mysql.GetQ(ctx, r.q).DeleteGroup(ctx, id)
}
