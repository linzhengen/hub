package permission

import (
	"context"

	"github.com/linzhengen/hub/v1/server/internal/domain/contextx"
	"github.com/linzhengen/hub/v1/server/internal/domain/system/permission"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence"
)

func New(q persistence.Querier) permission.Repository {
	return &repositoryImpl{q: q}
}

type repositoryImpl struct {
	q persistence.Querier
}

func (r repositoryImpl) FindOne(ctx context.Context, id string) (*permission.Permission, error) {
	q := persistence.GetQ(ctx, r.q)
	var p *persistence.PermissionModel
	var err error
	if contextx.FromTransLock(ctx) {
		p, err = q.SelectPermissionForUpdate(ctx, id)
	} else {
		p, err = q.SelectPermissionById(ctx, id)
	}
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
	ps, err := persistence.GetQ(ctx, r.q).SelectPermissionByResourceId(ctx, resourceId)
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
	return persistence.GetQ(ctx, r.q).CreatePermission(ctx, p.Id, string(p.Verb), p.ResourceId, p.Description)
}

func (r repositoryImpl) Update(ctx context.Context, p *permission.Permission) error {
	return persistence.GetQ(ctx, r.q).UpdatePermission(ctx, p.Id, string(p.Verb), p.ResourceId, p.Description)
}

func (r repositoryImpl) Delete(ctx context.Context, id string) error {
	return persistence.GetQ(ctx, r.q).DeletePermission(ctx, id)
}
