package system

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/doug-martin/goqu/v9"

	"github.com/linzhengen/hub/server/internal/domain/system/group"
	"github.com/linzhengen/hub/server/internal/domain/system/group/grouprole"
	"github.com/linzhengen/hub/server/internal/domain/trans"
	"github.com/linzhengen/hub/server/internal/domain/user/usergroup"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence/mysql"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence/mysql/sqlc"

	"github.com/linzhengen/hub/server/pkg/logger"
)

type GroupUseCase interface {
	Get(ctx context.Context, groupId string) (*group.Group, error)
	Create(ctx context.Context, g *group.Group) (*group.Group, error)
	Update(ctx context.Context, g *group.Group) (*group.Group, error)
	Delete(ctx context.Context, groupId string) error
	List(ctx context.Context, params *ListGroupQueryParams) ([]*group.Group, int64, error)
	AssignRole(ctx context.Context, groupId, roleId string) (*group.Group, error)
	AddUsersToGroup(ctx context.Context, groupID string, userIDs []string) (*group.Group, error)
	RemoveUsersFromGroup(ctx context.Context, groupID string, userIDs []string) (*group.Group, error)
	AssignRolesToGroup(ctx context.Context, groupID string, roleIDs []string) (*group.Group, error)
}

func NewGroupUseCase(
	db *sql.DB,
	transRepo trans.Repository,
	groupRepo group.Repository,
	groupRoleRepo grouprole.Repository,
	userGroupRepo usergroup.Repository,
	dialectWrapper mysql.DialectWrapper,
) GroupUseCase {
	return &groupUseCase{
		db:             db,
		transRepo:      transRepo,
		groupRepo:      groupRepo,
		groupRoleRepo:  groupRoleRepo,
		userGroupRepo:  userGroupRepo,
		dialectWrapper: dialectWrapper,
	}
}

type groupUseCase struct {
	db             *sql.DB
	transRepo      trans.Repository
	groupRepo      group.Repository
	groupRoleRepo  grouprole.Repository
	userGroupRepo  usergroup.Repository
	dialectWrapper mysql.DialectWrapper
}

type ListGroupQueryParams struct {
	Limit     uint32
	Offset    uint32
	GroupIds  []string
	GroupName string
	Status    group.Status
	RoleIds   []string
}

func (uc groupUseCase) Create(ctx context.Context, g *group.Group) (*group.Group, error) {
	if err := uc.groupRepo.Create(ctx, g); err != nil {
		return nil, err
	}
	return uc.groupRepo.FindOne(ctx, g.Id)
}

func (uc groupUseCase) Get(ctx context.Context, groupId string) (*group.Group, error) {
	return uc.groupRepo.FindOne(ctx, groupId)
}

func (uc groupUseCase) Update(ctx context.Context, g *group.Group) (*group.Group, error) {
	if err := uc.transRepo.ExecTransWithLock(ctx, func(ctx context.Context) error {
		_, err := uc.groupRepo.FindOne(ctx, g.Id)
		if err != nil {
			return err
		}
		return uc.groupRepo.Update(ctx, g)
	}); err != nil {
		return nil, err
	}
	return uc.groupRepo.FindOne(ctx, g.Id)
}

func (uc groupUseCase) Delete(ctx context.Context, groupId string) error {
	return uc.transRepo.ExecTransWithLock(ctx, func(ctx context.Context) error {
		_, err := uc.groupRepo.FindOne(ctx, groupId)
		if err != nil {
			return err
		}
		return uc.groupRepo.Delete(ctx, groupId)
	})
}

func (uc groupUseCase) List(ctx context.Context, params *ListGroupQueryParams) ([]*group.Group, int64, error) {
	b := uc.dialectWrapper.From("groups")
	if params.GroupIds != nil {
		b = b.Where(goqu.Ex{"id": params.GroupIds})
	}
	if params.GroupName != "" {
		b = b.Where(goqu.C("name").Like(fmt.Sprintf("%%%s%%", params.GroupName)))
	}
	if params.Status != "" {
		b = b.Where(goqu.Ex{"groups.status": params.Status})
	}
	if params.RoleIds != nil {
		// Create a subquery to check if the role belongs to any of the specified permissions
		subquery := uc.dialectWrapper.From("group_roles").
			Select(goqu.L("1")).
			Where(goqu.Ex{
				"group_roles.group_id": goqu.I("groups.id"),
				"group_roles.role_id":  params.RoleIds,
			})

		// Use EXISTS with the subquery
		b = b.Where(goqu.L("EXISTS ?", subquery))
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

func (uc groupUseCase) list(ctx context.Context, b *goqu.SelectDataset) ([]*group.Group, error) {
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
	var items []*group.Group
	for rows.Next() {
		var i sqlc.Group
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Status,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}

		items = append(items, &group.Group{
			Id:          i.ID,
			Name:        i.Name,
			Description: i.Description,
			Status:      group.Status(i.Status),
			CreatedAt:   i.CreatedAt,
			UpdatedAt:   i.UpdatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return items, nil
	}

	// Collect all group IDs
	var groupIds []string
	for _, item := range items {
		groupIds = append(groupIds, item.Id)
	}

	// Fetch all group-role relationships in a single query
	groupRoleMap := make(map[string][]string)

	// Build a query to get all group-role relationships for the group IDs
	grQuery := uc.dialectWrapper.From("group_roles").
		Select("group_id", "role_id").
		Where(goqu.Ex{"group_id": groupIds})

	grSQL, grParams, err := grQuery.Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}

	// Execute the query
	grRows, err := uc.db.QueryContext(ctx, grSQL, grParams...)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := grRows.Close()
		if err != nil {
			logger.Errorf("list: error closing group-role rows: %v", err)
		}
	}()

	// Process the results
	for grRows.Next() {
		var groupId, roleId string
		if err := grRows.Scan(&groupId, &roleId); err != nil {
			return nil, err
		}
		// Add the role ID to the map
		groupRoleMap[groupId] = append(groupRoleMap[groupId], roleId)
	}

	if err := grRows.Err(); err != nil {
		return nil, err
	}

	// Set role IDs for each group
	for _, item := range items {
		if roleIds, ok := groupRoleMap[item.Id]; ok {
			item.SetRoleIds(roleIds)
		}
	}

	return items, nil
}

func (uc groupUseCase) AssignRole(ctx context.Context, groupId, roleId string) (*group.Group, error) {
	if err := uc.groupRoleRepo.AssignRole(ctx, groupId, roleId); err != nil {
		return nil, err
	}
	return uc.Get(ctx, groupId)
}

func (uc groupUseCase) AddUsersToGroup(ctx context.Context, groupID string, userIDs []string) (*group.Group, error) {
	if err := uc.transRepo.ExecTrans(ctx, func(ctx context.Context) error {
		return uc.userGroupRepo.AddUsersToGroup(ctx, groupID, userIDs)
	}); err != nil {
		return nil, err
	}
	return uc.Get(ctx, groupID)
}

func (uc groupUseCase) RemoveUsersFromGroup(ctx context.Context, groupID string, userIDs []string) (*group.Group, error) {
	if err := uc.transRepo.ExecTrans(ctx, func(ctx context.Context) error {
		return uc.userGroupRepo.RemoveUsersFromGroup(ctx, groupID, userIDs)
	}); err != nil {
		return nil, err
	}
	return uc.Get(ctx, groupID)
}

func (uc groupUseCase) AssignRolesToGroup(ctx context.Context, groupID string, roleIDs []string) (*group.Group, error) {
	if err := uc.transRepo.ExecTrans(ctx, func(ctx context.Context) error {
		return uc.groupRoleRepo.Upsert(ctx, groupID, roleIDs)
	}); err != nil {
		return nil, err
	}
	return uc.Get(ctx, groupID)
}
