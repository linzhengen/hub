package resource

import (
	"context"

	"github.com/linzhengen/hub/v1/server/internal/domain/contextx"
	"github.com/linzhengen/hub/v1/server/internal/domain/system/resource"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence"
)

func New(q persistence.Querier) resource.Repository {
	return &repositoryImpl{q: q}
}

type repositoryImpl struct {
	q persistence.Querier
}

func (r repositoryImpl) FindOne(ctx context.Context, id string) (*resource.Resource, error) {
	q := persistence.GetQ(ctx, r.q)
	var res *persistence.ResourceModel
	var err error
	if contextx.FromTransLock(ctx) {
		res, err = q.SelectResourceForUpdate(ctx, id)
	} else {
		res, err = q.SelectResourceById(ctx, id)
	}
	if err != nil {
		return nil, err
	}
	var identifier resource.Identifier
	if err := identifier.Scan(res.Identifier); err != nil {
		return nil, err
	}

	return &resource.Resource{
		Id:           res.ID,
		ParentId:     res.ParentID,
		Name:         res.Name,
		Identifier:   identifier,
		Type:         resource.ResourceType(res.Type),
		Path:         res.Path,
		Component:    res.Component,
		DisplayOrder: res.DisplayOrder,
		Description:  res.Description,
		Metadata:     res.Metadata,
		Status:       resource.Status(res.Status),
		CreatedAt:    res.CreatedAt,
		UpdatedAt:    res.UpdatedAt,
	}, nil
}

func (r repositoryImpl) FindOneByIdentifier(ctx context.Context, identifier string) (*resource.Resource, error) {
	res, err := persistence.GetQ(ctx, r.q).SelectResourceByIdentifier(ctx, identifier)
	if err != nil {
		return nil, err
	}
	var i resource.Identifier
	if err := i.Scan(res.Identifier); err != nil {
		return nil, err
	}

	return &resource.Resource{
		Id:           res.ID,
		ParentId:     res.ParentID,
		Name:         res.Name,
		Identifier:   i,
		Type:         resource.ResourceType(res.Type),
		Path:         res.Path,
		Component:    res.Component,
		DisplayOrder: res.DisplayOrder,
		Description:  res.Description,
		Metadata:     res.Metadata,
		Status:       resource.Status(res.Status),
		CreatedAt:    res.CreatedAt,
		UpdatedAt:    res.UpdatedAt,
	}, nil
}

func (r repositoryImpl) Create(ctx context.Context, res *resource.Resource) error {
	return persistence.GetQ(ctx, r.q).CreateResource(ctx, res.Id, res.ParentId, res.Name, res.Identifier.String(), string(res.Type), res.Path, res.Component, int32(res.DisplayOrder), res.Description, res.Metadata, string(res.Status))
}

func (r repositoryImpl) Update(ctx context.Context, res *resource.Resource) error {
	return persistence.GetQ(ctx, r.q).UpdateResource(ctx, res.ParentId, res.Name, res.Identifier.String(), string(res.Type), res.Path, res.Component, int32(res.DisplayOrder), res.Description, res.Metadata, string(res.Status), res.Id)
}

func (r repositoryImpl) Delete(ctx context.Context, id string) error {
	return persistence.GetQ(ctx, r.q).DeleteResource(ctx, id)
}
