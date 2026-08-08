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
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence/postgres"
	"github.com/linzhengen/hub/server/internal/usecase"
	"github.com/linzhengen/hub/server/pkg/logger"
)

type userFinder struct {
	db      *sql.DB
	dialect postgres.DialectWrapper
}

func NewFinder(db *sql.DB, dialect postgres.DialectWrapper) usecase.UserFinder {
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
			// "menu.*" → trim "*" → "menu." → trim trailing "." → "menu"
			// LIKE 'menu%' matches "menu", "menu.foo", etc.
			prefix := strings.TrimRight(strings.TrimSuffix(identifier, "*"), ".")
			if prefix == "" {
				// bare "*" means all menus
				exps = append(exps, goqu.L("TRUE"))
			} else {
				exps = append(exps, goqu.Or(
					goqu.I("identifier").Eq(prefix),
					goqu.I("identifier").Like(prefix+".%"),
				))
			}
		} else {
			exps = append(exps, goqu.I("identifier").Eq(identifier))
		}
	}

	if len(exps) == 0 {
		return []*menu.Menu{}, nil
	}

	// Step 1: fetch the menus the user has direct permission for.
	directQueryBuilder := f.dialect.From("resources").
		Where(goqu.I("type").Eq("menu")).
		Where(goqu.I("status").Eq("Active")).
		Where(goqu.Or(exps...)).
		Where(goqu.I("identifier").NotLike("%*")).
		Order(goqu.I("display_order").Asc())

	directQuery, directQueryParams, err := directQueryBuilder.Select(
		"id", "parent_id", "name", "identifier", "type", "path", "component",
		"display_order", "description", "metadata", "status", "created_at", "updated_at",
	).Prepared(true).ToSQL()
	if err != nil {
		logger.Errorf("GetMeMenus: failed to build menu SQL query: %v", err)
		return nil, err
	}

	menuRows, err := f.db.QueryContext(ctx, directQuery, directQueryParams...)
	if err != nil {
		logger.Errorf("GetMeMenus: failed to execute menu SQL query: %v", err)
		return nil, err
	}
	defer func() {
		if err := menuRows.Close(); err != nil {
			logger.Errorf("GetMeMenus: error closing menu rows: %v", err)
		}
	}()

	directMenus, err := f.scanMenus(menuRows)
	if err != nil {
		return nil, err
	}

	// Step 2: collect parent IDs that are not already in the result set,
	// so that a child-only permission still shows its ancestor menus.
	directIDs := make(map[string]struct{}, len(directMenus))
	for _, m := range directMenus {
		directIDs[m.Id] = struct{}{}
	}

	var missingParentIDs []interface{}
	for _, m := range directMenus {
		pid := strings.TrimSpace(m.ParentId)
		if pid != "" {
			if _, found := directIDs[pid]; !found {
				missingParentIDs = append(missingParentIDs, pid)
				directIDs[pid] = struct{}{} // avoid duplicates
			}
		}
	}

	if len(missingParentIDs) == 0 {
		return directMenus, nil
	}

	// Fetch missing ancestors (one level is enough for a two-level menu tree;
	// extend to a loop if deeper nesting is needed).
	parentQueryBuilder := f.dialect.From("resources").
		Where(goqu.I("type").Eq("menu")).
		Where(goqu.I("status").Eq("Active")).
		Where(goqu.I("id").In(missingParentIDs...)).
		Order(goqu.I("display_order").Asc())

	parentQuery, parentQueryParams, err := parentQueryBuilder.Select(
		"id", "parent_id", "name", "identifier", "type", "path", "component",
		"display_order", "description", "metadata", "status", "created_at", "updated_at",
	).Prepared(true).ToSQL()
	if err != nil {
		logger.Errorf("GetMeMenus: failed to build parent menu SQL query: %v", err)
		return nil, err
	}

	parentRows, err := f.db.QueryContext(ctx, parentQuery, parentQueryParams...)
	if err != nil {
		logger.Errorf("GetMeMenus: failed to execute parent menu SQL query: %v", err)
		return nil, err
	}
	defer func() {
		if err := parentRows.Close(); err != nil {
			logger.Errorf("GetMeMenus: error closing parent menu rows: %v", err)
		}
	}()

	parentMenus, err := f.scanMenus(parentRows)
	if err != nil {
		return nil, err
	}

	return append(parentMenus, directMenus...), nil
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
