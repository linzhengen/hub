package user

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/linzhengen/hub/v1/server/internal/domain/contextx"
)

func TestUserFinder_GetMeMenus(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)

	dialect := goqu.Dialect("postgres")

	finder := &userFinder{
		db:      db,
		dialect: dialect,
	}

	userId := "test-user"
	ctx = contextx.WithUserID(ctx, userId)

	metadata := map[string]interface{}{"key": "value"}
	metadataJson, _ := json.Marshal(metadata)

	// Mock for the first query (identifiers)
	identRows := sqlmock.NewRows([]string{"identifier"}).AddRow("menu.menu1")
	b := dialect.From(goqu.I("user_groups").As("ug")).
		Join(goqu.I("group_roles").As("gr"), goqu.On(goqu.I("ug.group_id").Eq(goqu.I("gr.group_id")))).
		Join(goqu.I("role_permissions").As("rp"), goqu.On(goqu.I("gr.role_id").Eq(goqu.I("rp.role_id")))).
		Join(goqu.I("permissions").As("p"), goqu.On(goqu.I("rp.permission_id").Eq(goqu.I("p.id")))).
		Join(goqu.I("resources").As("r"), goqu.On(goqu.I("p.resource_id").Eq(goqu.I("r.id")))).
		Where(goqu.I("ug.user_id").Eq(userId)).
		Where(goqu.I("r.type").Eq("menu"))

	identQuery, identArgs, _ := b.Select(goqu.I("r.identifier")).Prepared(true).ToSQL()
	driverIdentArgs := make([]driver.Value, len(identArgs))
	for i, arg := range identArgs {
		driverIdentArgs[i] = arg
	}
	mock.ExpectQuery(identQuery).WithArgs(driverIdentArgs...).WillReturnRows(identRows)

	// Mock for the second query (menus)
	menuRows := sqlmock.NewRows([]string{
		"id", "parent_id", "name", "identifier", "type", "path", "component", "display_order", "description", "metadata", "status", "created_at", "updated_at",
	}).AddRow(
		"menu1", "", "Menu 1", "menu.menu1", "menu", "/menu1", "Menu1", 1, "Menu 1", metadataJson, "active", time.Now(), time.Now(),
	)

	menuQueryBuilder := dialect.From("resources").
		Where(goqu.I("type").Eq("menu")).
		Where(goqu.Or(goqu.I("identifier").Eq("menu.menu1"))).
		Where(goqu.I("identifier").NotLike("%*")).
		Order(goqu.I("display_order").Asc())

	menuQuery, menuArgs, _ := menuQueryBuilder.Select(
		"id", "parent_id", "name", "identifier", "type", "path", "component",
		"display_order", "description", "metadata", "status", "created_at", "updated_at",
	).Prepared(true).ToSQL()

	driverMenuArgs := make([]driver.Value, len(menuArgs))
	for i, arg := range menuArgs {
		driverMenuArgs[i] = arg
	}
	mock.ExpectQuery(menuQuery).WithArgs(driverMenuArgs...).WillReturnRows(menuRows)

	menus, err := finder.GetMeMenus(ctx)

	assert.NoError(t, err)
	assert.Len(t, menus, 1)
	assert.Equal(t, "Menu 1", menus[0].Name)

	assert.NoError(t, mock.ExpectationsWereMet())
}
