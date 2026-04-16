package system

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/doug-martin/goqu/v9"

	"github.com/linzhengen/hub/v1/server/internal/domain/system/permission"
	"github.com/linzhengen/hub/v1/server/internal/domain/trans"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence/mysql"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence/mysql/sqlc"

	"github.com/linzhengen/hub/v1/server/pkg/logger"
)

type PermissionUseCase interface {
	Get(ctx context.Context, permissionId string) (*permission.Permission, error)
	Create(ctx context.Context, p *permission.Permission) (*permission.Permission, error)
	Update(ctx context.Context, p *permission.Permission) (*permission.Permission, error)
	Delete(ctx context.Context, permissionId string) error
	List(ctx context.Context, params *ListPermissionQueryParams) ([]*permission.Permission, int64, error)
}

func NewPermissionUseCase(
	db *sql.DB,
	transRepo trans.Repository,
	permissionRepo permission.Repository,
	dialectWrapper mysql.DialectWrapper,
) PermissionUseCase {
	return &permissionUseCase{
		db:             db,
		transRepo:      transRepo,
		permissionRepo: permissionRepo,
		dialectWrapper: dialectWrapper,
	}
}

type permissionUseCase struct {
	db             *sql.DB
	transRepo      trans.Repository
	permissionRepo permission.Repository
	dialectWrapper mysql.DialectWrapper
}

type ListPermissionQueryParams struct {
	Limit          uint32
	Offset         uint32
	PermissionIds  []string
	PermissionName string
}

func (uc permissionUseCase) Get(ctx context.Context, permissionId string) (*permission.Permission, error) {
	return uc.permissionRepo.FindOne(ctx, permissionId)
}

func (uc permissionUseCase) Create(ctx context.Context, p *permission.Permission) (*permission.Permission, error) {
	if err := uc.permissionRepo.Create(ctx, p); err != nil {
		return nil, err
	}
	return uc.permissionRepo.FindOne(ctx, p.Id)
}

func (uc permissionUseCase) Update(ctx context.Context, p *permission.Permission) (*permission.Permission, error) {
	if err := uc.transRepo.ExecTransWithLock(ctx, func(ctx context.Context) error {
		_, err := uc.permissionRepo.FindOne(ctx, p.Id)
		if err != nil {
			return err
		}
		return uc.permissionRepo.Update(ctx, p)
	}); err != nil {
		return nil, err
	}
	return uc.permissionRepo.FindOne(ctx, p.Id)
}

func (uc permissionUseCase) Delete(ctx context.Context, permissionId string) error {
	return uc.transRepo.ExecTransWithLock(ctx, func(ctx context.Context) error {
		_, err := uc.permissionRepo.FindOne(ctx, permissionId)
		if err != nil {
			return err
		}
		return uc.permissionRepo.Delete(ctx, permissionId)
	})
}

func (uc permissionUseCase) List(ctx context.Context, params *ListPermissionQueryParams) ([]*permission.Permission, int64, error) {
	b := uc.dialectWrapper.From("permissions")
	if params.PermissionIds != nil {
		b = b.Where(goqu.Ex{"id": params.PermissionIds})
	}
	if params.PermissionName != "" {
		b = b.Where(goqu.C("name").Like(fmt.Sprintf("%%%s%%", params.PermissionName)))
	}
	cnt, err := mysql.SelectCount(ctx, uc.db, b)
	if err != nil {
		return nil, 0, err
	}

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

func (uc permissionUseCase) list(ctx context.Context, b *goqu.SelectDataset) ([]*permission.Permission, error) {
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
	var items []*permission.Permission
	for rows.Next() {
		var i sqlc.Permission
		if err := rows.Scan(
			&i.ID,
			&i.Verb,
			&i.ResourceID,
			&i.Description,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}

		items = append(items, &permission.Permission{
			Id:          i.ID,
			Verb:        permission.Verb(i.Verb),
			ResourceId:  i.ResourceID,
			Description: i.Description,
			CreatedAt:   i.CreatedAt,
			UpdatedAt:   i.UpdatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
