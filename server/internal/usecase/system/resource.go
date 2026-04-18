package system

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/doug-martin/goqu/v9"

	"github.com/linzhengen/hub/v1/server/internal/domain/system/permission"
	"github.com/linzhengen/hub/v1/server/internal/domain/system/resource"
	"github.com/linzhengen/hub/v1/server/internal/domain/trans"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence/postgres"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence/postgres/sqlc"

	"github.com/linzhengen/hub/v1/server/pkg/logger"
)

type ResourceUseCase interface {
	Get(ctx context.Context, resourceId string) (*resource.Resource, error)
	Create(ctx context.Context, r *resource.Resource) (*resource.Resource, error)
	Update(ctx context.Context, r *resource.Resource) (*resource.Resource, error)
	Delete(ctx context.Context, resourceId string) error
	List(ctx context.Context, params *ListResourceQueryParams) ([]*resource.Resource, int64, error)
	CreateMenu(ctx context.Context, r *resource.Resource) (*resource.Resource, error)
	UpdateMenu(ctx context.Context, r *resource.Resource) (*resource.Resource, error)
}

func NewResourceUseCase(
	db *sql.DB,
	transRepo trans.Repository,
	resourceRepo resource.Repository,
	permissionRepo permission.Repository,
	dialectWrapper persistence.DialectWrapper,
) ResourceUseCase {
	return &resourceUseCase{
		db:             db,
		transRepo:      transRepo,
		resourceRepo:   resourceRepo,
		permissionRepo: permissionRepo,
		dialectWrapper: dialectWrapper,
	}
}

type resourceUseCase struct {
	db             *sql.DB
	transRepo      trans.Repository
	resourceRepo   resource.Repository
	permissionRepo permission.Repository
	dialectWrapper persistence.DialectWrapper
}

type ListResourceQueryParams struct {
	Limit        uint32
	Offset       uint32
	ResourceIds  []string
	ResourceName string
	ResourceType resource.ResourceType
}

func (uc resourceUseCase) Get(ctx context.Context, resourceId string) (*resource.Resource, error) {
	return uc.resourceRepo.FindOne(ctx, resourceId)
}

func (uc resourceUseCase) Create(ctx context.Context, r *resource.Resource) (*resource.Resource, error) {
	if err := uc.resourceRepo.Create(ctx, r); err != nil {
		return nil, err
	}
	return uc.resourceRepo.FindOne(ctx, r.Id)
}

func (uc resourceUseCase) Update(ctx context.Context, r *resource.Resource) (*resource.Resource, error) {
	if err := uc.transRepo.ExecTransWithLock(ctx, func(ctx context.Context) error {
		_, err := uc.resourceRepo.FindOne(ctx, r.Id)
		if err != nil {
			return err
		}
		return uc.resourceRepo.Update(ctx, r)
	}); err != nil {
		return nil, err
	}
	return uc.resourceRepo.FindOne(ctx, r.Id)
}

func (uc resourceUseCase) Delete(ctx context.Context, resourceId string) error {
	return uc.transRepo.ExecTransWithLock(ctx, func(ctx context.Context) error {
		_, err := uc.resourceRepo.FindOne(ctx, resourceId)
		if err != nil {
			return err
		}
		return uc.resourceRepo.Delete(ctx, resourceId)
	})
}

func (uc resourceUseCase) List(ctx context.Context, params *ListResourceQueryParams) ([]*resource.Resource, int64, error) {
	b := uc.dialectWrapper.From("resources")
	if params.ResourceIds != nil {
		b = b.Where(goqu.Ex{"id": params.ResourceIds})
	}
	if params.ResourceName != "" {
		b = b.Where(goqu.C("name").Like(fmt.Sprintf("%%%s%%", params.ResourceName)))
	}
	if params.ResourceType == resource.ResourceTypeApi {
		b = b.Where(goqu.C("type").Eq(resource.ResourceTypeApi))
	}
	if params.ResourceType == resource.ResourceTypeMenu {
		b = b.Where(goqu.C("type").Eq(resource.ResourceTypeMenu))
		b = b.Where(goqu.C("identifier").Neq("menu.*"))
	}
	cnt, err := postgres.SelectCount(ctx, uc.db, b)
	if err != nil {
		return nil, 0, err
	}

	b = b.Order(goqu.C("display_order").Asc())

	// Apply pagination only when limit > 0
	if params.Limit > 0 {
		b = b.Limit(uint(params.Limit)).Offset(uint(params.Offset))
	}

	items, err := uc.list(ctx, b)
	if err != nil {
		return nil, 0, err
	}
	return items, cnt, nil
}

func (uc resourceUseCase) list(ctx context.Context, b *goqu.SelectDataset) ([]*resource.Resource, error) {
	b = b.Select("*")
	query, queryParams, err := b.Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}
	rows, err := uc.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			logger.Infof("error closing rows: %v", err)
		}
	}()
	var items []*resource.Resource
	for rows.Next() {
		var i sqlc.Resource
		if err := rows.Scan(
			&i.ID,
			&i.ParentID,
			&i.Name,
			&i.Identifier,
			&i.Type,
			&i.Path,
			&i.Component,
			&i.DisplayOrder,
			&i.Description,
			&i.Metadata,
			&i.Status,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		identifier := new(resource.Identifier)
		if err := identifier.Scan(i.Identifier); err != nil {
			return nil, err
		}
		var metadata map[string]string
		if i.Metadata.Valid {
			if err := json.Unmarshal(i.Metadata.RawMessage, &metadata); err != nil {
				return nil, err
			}
		}
		items = append(items, &resource.Resource{
			Id:           i.ID,
			ParentId:     i.ParentID,
			Name:         i.Name,
			Identifier:   *identifier,
			Type:         resource.ResourceType(i.Type),
			Path:         i.Path.String,
			Component:    i.Component.String,
			DisplayOrder: i.DisplayOrder.Int32,
			Description:  i.Description,
			Metadata:     metadata,
			Status:       resource.Status(i.Status),
			CreatedAt:    i.CreatedAt,
			UpdatedAt:    i.UpdatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (uc resourceUseCase) CreateMenu(ctx context.Context, r *resource.Resource) (*resource.Resource, error) {
	if err := r.ValidateForMenu(); err != nil {
		return nil, err
	}
	r.Type = resource.ResourceTypeMenu
	r.Identifier.Api = string(r.Type)
	r.Identifier.Category = r.Path

	if err := uc.transRepo.ExecTrans(ctx, func(ctx context.Context) error {
		if err := uc.resourceRepo.Create(ctx, r); err != nil {
			return err
		}
		p, err := permission.Factory("get", r.Id, "")
		if err != nil {
			return err
		}
		if err := uc.permissionRepo.Create(ctx, p); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return uc.resourceRepo.FindOne(ctx, r.Id)
}

func (uc resourceUseCase) UpdateMenu(ctx context.Context, r *resource.Resource) (*resource.Resource, error) {
	if err := r.ValidateForMenu(); err != nil {
		return nil, err
	}
	r.Type = resource.ResourceTypeMenu
	r.Identifier.Api = string(r.Type)
	r.Identifier.Category = r.Path

	if err := uc.transRepo.ExecTransWithLock(ctx, func(ctx context.Context) error {
		old, err := uc.resourceRepo.FindOne(ctx, r.Id)
		if err != nil {
			return err
		}
		if err := uc.resourceRepo.Update(ctx, r); err != nil {
			return err
		}

		if old.Identifier.String() != r.Identifier.String() {
			ps, err := uc.permissionRepo.FindByResourceId(ctx, old.Id)
			if err != nil {
				return err
			}
			for _, p := range ps {
				if err := uc.permissionRepo.Delete(ctx, p.Id); err != nil {
					return err
				}
			}
			p, err := permission.Factory("get", r.Id, "")
			if err != nil {
				return err
			}
			if err := uc.permissionRepo.Create(ctx, p); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return uc.resourceRepo.FindOne(ctx, r.Id)
}
