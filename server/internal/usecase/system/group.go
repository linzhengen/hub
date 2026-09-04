package system

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"

	"github.com/linzhengen/hub/server/internal/domain/system/group"
	"github.com/linzhengen/hub/server/internal/domain/system/group/grouprole"
	"github.com/linzhengen/hub/server/internal/domain/trans"
	"github.com/linzhengen/hub/server/internal/domain/user/usergroup"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence/postgres"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence/postgres/sqlc"
	"github.com/linzhengen/hub/server/internal/usecase/pagination"

	"github.com/linzhengen/hub/server/pkg/logger"
)

type GroupUseCase interface {
	Get(ctx context.Context, groupId string) (*group.Group, error)
	Create(ctx context.Context, g *group.Group) (*group.Group, error)
	Update(ctx context.Context, g *group.Group) (*group.Group, error)
	Delete(ctx context.Context, groupId string) error
	List(ctx context.Context, params *ListGroupQueryParams) ([]*group.Group, int64, error)
	AddRoles(ctx context.Context, groupID string, roleIDs []string, expiresAt *time.Time) (*group.Group, error)
	RemoveRoles(ctx context.Context, groupID string, roleIDs []string) (*group.Group, error)
	AddUsers(ctx context.Context, groupID string, userIDs []string, expiresAt *time.Time) (*group.Group, error)
	RemoveUsers(ctx context.Context, groupID string, userIDs []string) (*group.Group, error)
}

func NewGroupUseCase(
	db *sql.DB,
	transRepo trans.Repository,
	groupRepo group.Repository,
	groupRoleRepo grouprole.Repository,
	userGroupRepo usergroup.Repository,
	dialectWrapper persistence.DialectWrapper,
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
	dialectWrapper persistence.DialectWrapper
}

type ListGroupQueryParams struct {
	Limit     uint32
	Offset    uint32
	GroupIds  []string
	GroupName string
	Status    group.Status
	RoleIds   []string
	// OrgId narrows the listing to one organization. Empty is every
	// organization, which is what this query meant before there were any.
	OrgId string
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
		b = b.Where(postgres.In("id", params.GroupIds))
	}
	if params.GroupName != "" {
		b = b.Where(goqu.C("name").Like(fmt.Sprintf("%%%s%%", params.GroupName)))
	}
	if params.Status != "" {
		b = b.Where(goqu.Ex{"groups.status": params.Status})
	}
	if params.OrgId != "" {
		b = b.Where(goqu.Ex{"groups.org_id": params.OrgId})
	}
	if params.RoleIds != nil {
		// Create a subquery to check if the role belongs to any of the specified permissions
		subquery := uc.dialectWrapper.From("group_roles").
			Select(goqu.L("1")).
			Where(goqu.Ex{
				"group_roles.group_id": goqu.I("groups.id"),
			}).Where(postgres.In("group_roles.role_id", params.RoleIds))

		// Use EXISTS with the subquery
		b = b.Where(goqu.L("EXISTS ?", subquery))
	}
	cnt, err := postgres.SelectCount(ctx, uc.db, b)
	if err != nil {
		return nil, 0, err
	}

	page := pagination.New(params.Limit, params.Offset)
	b = b.Limit(page.Limit()).Offset(page.Offset())

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
			// Scanned positionally from SELECT *, so the order here has to
			// follow the column order in 000002.
			&i.OrgID,
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
			OrgId:       i.OrgID,
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
	groupRoleMap := make(map[string][]group.RoleGrant)

	// Build a query to get all group-role relationships for the group IDs
	grQuery := uc.dialectWrapper.From("group_roles").
		Select("group_id", "role_id", "expires_at").
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
		var expiresAt sql.NullTime
		if err := grRows.Scan(&groupId, &roleId, &expiresAt); err != nil {
			return nil, err
		}
		grant := group.RoleGrant{RoleId: roleId}
		if expiresAt.Valid {
			grant.ExpiresAt = &expiresAt.Time
		}
		groupRoleMap[groupId] = append(groupRoleMap[groupId], grant)
	}

	if err := grRows.Err(); err != nil {
		return nil, err
	}

	// Set the role grants for each group
	for _, item := range items {
		if grants, ok := groupRoleMap[item.Id]; ok {
			item.SetRoles(grants)
		}
	}

	return items, nil
}

// AddRoles grants each of roleIDs to the group, keeping the roles it already
// has. The set is applied in one transaction, so a failure part way through
// does not leave the group holding some of the roles but not the rest.
//
// expiresAt applies to every role in the call and is nil for a grant that does
// not end.
func (uc groupUseCase) AddRoles(
	ctx context.Context,
	groupID string,
	roleIDs []string,
	expiresAt *time.Time,
) (*group.Group, error) {
	return uc.changeRoles(ctx, groupID, roleIDs,
		func(ctx context.Context, groupID, roleID string) error {
			return uc.groupRoleRepo.AssignRole(ctx, groupID, roleID, expiresAt)
		})
}

// RemoveRoles revokes each of roleIDs from the group.
func (uc groupUseCase) RemoveRoles(ctx context.Context, groupID string, roleIDs []string) (*group.Group, error) {
	return uc.changeRoles(ctx, groupID, roleIDs, uc.groupRoleRepo.UnassignRole)
}

func (uc groupUseCase) changeRoles(
	ctx context.Context,
	groupID string,
	roleIDs []string,
	apply func(ctx context.Context, groupID, roleID string) error,
) (*group.Group, error) {
	if err := uc.transRepo.ExecTrans(ctx, func(ctx context.Context) error {
		for _, roleID := range roleIDs {
			if err := apply(ctx, groupID, roleID); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return uc.Get(ctx, groupID)
}

func (uc groupUseCase) AddUsers(
	ctx context.Context,
	groupID string,
	userIDs []string,
	expiresAt *time.Time,
) (*group.Group, error) {
	if err := uc.transRepo.ExecTrans(ctx, func(ctx context.Context) error {
		return uc.userGroupRepo.AddUsersToGroup(ctx, groupID, userIDs, expiresAt)
	}); err != nil {
		return nil, err
	}
	return uc.Get(ctx, groupID)
}

func (uc groupUseCase) RemoveUsers(ctx context.Context, groupID string, userIDs []string) (*group.Group, error) {
	if err := uc.transRepo.ExecTrans(ctx, func(ctx context.Context) error {
		return uc.userGroupRepo.RemoveUsersFromGroup(ctx, groupID, userIDs)
	}); err != nil {
		return nil, err
	}
	return uc.Get(ctx, groupID)
}
