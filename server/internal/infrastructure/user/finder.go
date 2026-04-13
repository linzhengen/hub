package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"

	"github.com/linzhengen/hub/server/internal/domain/contextx"
	"github.com/linzhengen/hub/server/internal/domain/system/resource/menu"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence/mysql"
	"github.com/linzhengen/hub/server/internal/usecase"

	"github.com/linzhengen/hub/server/pkg/logger"
)

type userFinder struct {
	db      *sql.DB
	dialect mysql.DialectWrapper
}

func NewFinder(db *sql.DB, dialect mysql.DialectWrapper) usecase.UserFinder {
	return &userFinder{db: db, dialect: dialect}
}

func (f *userFinder) GetMeMenus(ctx context.Context) ([]*menu.Menu, error) {
	userId, ok := contextx.GetUserID(ctx)
	if !ok {
		err := fmt.Errorf("user not found in context")
		logger.Errorf("GetMeMenus: %v", err)
		return nil, err
	}

	b := f.dialect.From(goqu.I("user_groups").As("ug")).
		Join(goqu.I("group_roles").As("gr"), goqu.On(goqu.I("ug.group_id").Eq(goqu.I("gr.group_id")))).
		Join(goqu.I("role_permissions").As("rp"), goqu.On(goqu.I("gr.role_id").Eq(goqu.I("rp.role_id")))).
		Join(goqu.I("permissions").As("p"), goqu.On(goqu.I("rp.permission_id").Eq(goqu.I("p.id")))).
		Join(goqu.I("resources").As("r"), goqu.On(goqu.I("p.resource_id").Eq(goqu.I("r.id")))).
		Where(goqu.I("ug.user_id").Eq(userId)).
		Where(goqu.I("r.type").Eq("menu"))

	query, queryParams, err := b.Select(goqu.I("r.identifier")).Prepared(true).ToSQL()
	if err != nil {
		logger.Errorf("GetMeMenus: failed to build SQL query: %v", err)
		return nil, err
	}

	rows, err := f.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		logger.Errorf("GetMeMenus: failed to execute SQL query: %v", err)
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			logger.Errorf("GetMeMenus: error closing rows: %v", err)
		}
	}()

	var identifiers []string
	for rows.Next() {
		var identifier string
		if err := rows.Scan(&identifier); err != nil {
			logger.Errorf("GetMeMenus: failed to scan identifier: %v", err)
			return nil, err
		}
		identifiers = append(identifiers, identifier)
	}

	if err := rows.Err(); err != nil {
		logger.Errorf("GetMeMenus: error after iterating rows: %v", err)
		return nil, err
	}

	var exps []goqu.Expression
	for _, identifier := range identifiers {
		if strings.HasSuffix(identifier, "*") {
			prefix := strings.TrimSuffix(identifier, "*")
			exps = append(exps, goqu.I("identifier").Like(prefix+"%"))
		} else {
			exps = append(exps, goqu.I("identifier").Eq(identifier))
		}
	}

	if len(exps) == 0 {
		return []*menu.Menu{}, nil
	}

	menuQueryBuilder := f.dialect.From("resources").
		Where(goqu.I("type").Eq("menu")).
		Where(goqu.Or(exps...)).
		Where(goqu.I("identifier").NotLike("%*")).
		Order(goqu.I("display_order").Asc())

	menuQuery, menuQueryParams, err := menuQueryBuilder.Select(
		"id", "parent_id", "name", "identifier", "type", "path", "component",
		"display_order", "description", "metadata", "status", "created_at", "updated_at",
	).Prepared(true).ToSQL()
	if err != nil {
		logger.Errorf("GetMeMenus: failed to build menu SQL query: %v", err)
		return nil, err
	}

	menuRows, err := f.db.QueryContext(ctx, menuQuery, menuQueryParams...)
	if err != nil {
		logger.Errorf("GetMeMenus: failed to execute menu SQL query: %v", err)
		return nil, err
	}
	defer func() {
		err := menuRows.Close()
		if err != nil {
			logger.Errorf("GetMeMenus: error closing menu rows: %v", err)
		}
	}()

	return f.scanMenus(menuRows)
}

func (f *userFinder) scanMenus(rows *sql.Rows) ([]*menu.Menu, error) {
	var items []*menu.Menu
	for rows.Next() {
		var i menu.Menu
		var metadata []byte
		var parentId, path, component, description sql.NullString
		if err := rows.Scan(
			&i.Id,
			&parentId,
			&i.Name,
			&i.Identifier,
			&i.Type,
			&path,
			&component,
			&i.DisplayOrder,
			&description,
			&metadata,
			&i.Status,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			logger.Errorf("scanMenus: failed to scan row: %v", err)
			return nil, err
		}
		i.ParentId = parentId.String
		i.Path = path.String
		i.Component = component.String
		i.Description = description.String
		if err := json.Unmarshal(metadata, &i.Metadata); err != nil {
			logger.Errorf("scanMenus: failed to unmarshal metadata: %v", err)
			return nil, err
		}
		items = append(items, &i)
	}

	if err := rows.Err(); err != nil {
		logger.Errorf("scanMenus: error after iterating rows: %v", err)
		return nil, err
	}

	return items, nil
}
