package resource

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/linzhengen/hub/v1/server/internal/domain/contextx"
	"github.com/linzhengen/hub/v1/server/internal/domain/system/resource"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence/mysql"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence/mysql/sqlc"
)

func New(q *sqlc.Queries) resource.Repository {
	return &repositoryImpl{q: q}
}

type repositoryImpl struct {
	q *sqlc.Queries
}

func (r repositoryImpl) FindOne(ctx context.Context, id string) (*resource.Resource, error) {
	res, err := contextx.FindOne[sqlc.Resource](ctx, id, mysql.GetQ(ctx, mysql.GetQ(ctx, r.q)).SelectResourceById, mysql.GetQ(ctx, mysql.GetQ(ctx, r.q)).SelectResourceForUpdate)
	if err != nil {
		return nil, err
	}
	var identifier resource.Identifier
	if err := identifier.Scan(res.Identifier); err != nil {
		return nil, err
	}
	var metadata map[string]string
	if err := json.Unmarshal(res.Metadata, &metadata); err != nil {
		return nil, err
	}
	return &resource.Resource{
		Id:           res.ID,
		ParentId:     res.ParentID,
		Name:         res.Name,
		Identifier:   identifier,
		Type:         resource.ResourceType(res.Type),
		Path:         res.Path.String,
		Component:    res.Component.String,
		DisplayOrder: res.DisplayOrder.Int32,
		Description:  res.Description,
		Metadata:     metadata,
		Status:       resource.Status(res.Status),
		CreatedAt:    res.CreatedAt,
		UpdatedAt:    res.UpdatedAt,
	}, nil
}

func (r repositoryImpl) FindOneByIdentifier(ctx context.Context, identifier string) (*resource.Resource, error) {
	res, err := mysql.GetQ(ctx, r.q).SelectResourceByIdentifier(ctx, identifier)
	if err != nil {
		return nil, err
	}
	var i resource.Identifier
	if err := i.Scan(res.Identifier); err != nil {
		return nil, err
	}
	var metadata map[string]string
	if err := json.Unmarshal(res.Metadata, &metadata); err != nil {
		return nil, err
	}
	return &resource.Resource{
		Id:           res.ID,
		ParentId:     res.ParentID,
		Name:         res.Name,
		Identifier:   i,
		Type:         resource.ResourceType(res.Type),
		Path:         res.Path.String,
		Component:    res.Component.String,
		DisplayOrder: res.DisplayOrder.Int32,
		Description:  res.Description,
		Metadata:     metadata,
		Status:       resource.Status(res.Status),
		CreatedAt:    res.CreatedAt,
		UpdatedAt:    res.UpdatedAt,
	}, nil
}

func (r repositoryImpl) Create(ctx context.Context, res *resource.Resource) error {
	md, err := res.Metadata.JsonRawMessage()
	if err != nil {
		return err
	}
	if err := mysql.GetQ(ctx, r.q).CreateResource(ctx, sqlc.CreateResourceParams{
		ID:           res.Id,
		ParentID:     res.ParentId,
		Name:         res.Name,
		Identifier:   res.Identifier.String(),
		Type:         string(res.Type),
		Path:         sql.NullString{String: res.Path, Valid: res.Path != ""},
		Component:    sql.NullString{String: res.Component, Valid: res.Component != ""},
		DisplayOrder: sql.NullInt32{Int32: int32(res.DisplayOrder), Valid: res.DisplayOrder != 0},
		Description:  res.Description,
		Metadata:     md,
		Status:       string(res.Status),
	}); err != nil {
		return err
	}
	return nil
}

func (r repositoryImpl) Update(ctx context.Context, res *resource.Resource) error {
	md, err := res.Metadata.JsonRawMessage()
	if err != nil {
		return err
	}
	return mysql.GetQ(ctx, r.q).UpdateResource(ctx, sqlc.UpdateResourceParams{
		ParentID:     res.ParentId,
		Name:         res.Name,
		Identifier:   res.Identifier.String(),
		Type:         string(res.Type),
		Path:         sql.NullString{String: res.Path, Valid: res.Path != ""},
		Component:    sql.NullString{String: res.Component, Valid: res.Component != ""},
		DisplayOrder: sql.NullInt32{Int32: int32(res.DisplayOrder), Valid: res.DisplayOrder != 0},
		Description:  res.Description,
		Metadata:     md,
		Status:       string(res.Status),
		ID:           res.Id,
	})
}

func (r repositoryImpl) Delete(ctx context.Context, id string) error {
	return mysql.GetQ(ctx, r.q).DeleteResource(ctx, id)
}
