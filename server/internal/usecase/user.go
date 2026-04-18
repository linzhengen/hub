package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/doug-martin/goqu/v9"

	"github.com/linzhengen/hub/v1/server/internal/domain/contextx"
	oidcUserDomain "github.com/linzhengen/hub/v1/server/internal/domain/oidc/user"
	"github.com/linzhengen/hub/v1/server/internal/domain/system/resource/menu"
	"github.com/linzhengen/hub/v1/server/internal/domain/trans"
	"github.com/linzhengen/hub/v1/server/internal/domain/user"
	"github.com/linzhengen/hub/v1/server/internal/domain/user/usergroup"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence"

	"github.com/linzhengen/hub/v1/server/pkg/logger"
)

type UserUseCase interface {
	Me(ctx context.Context) (*user.User, error)
	Get(ctx context.Context, userId string) (*user.User, error)
	Update(ctx context.Context, u *user.User, password *string) (*user.User, error)
	Delete(ctx context.Context, userId string) error
	Create(ctx context.Context, username, email, password string, groupIds []string) (*user.User, error)
	List(ctx context.Context, params *ListUserQueryParams) ([]*user.User, int64, error)
	AssignGroup(ctx context.Context, userId, groupId string) (*user.User, error)
	UnassignGroup(ctx context.Context, userId, groupId string) (*user.User, error)
	GetMeMenus(ctx context.Context) ([]*menu.Menu, error)
}

func NewUserUseCase(
	db *sql.DB,
	dialectWrapper persistence.DialectWrapper,
	transRepo trans.Repository,
	userRepo user.Repository,
	userSvc user.Service,
	userGroupRepo usergroup.Repository,
	oidcUserRepo oidcUserDomain.Repository,
	userFinder UserFinder,
) UserUseCase {
	return &userUseCase{
		db:             db,
		dialectWrapper: dialectWrapper,
		transRepo:      transRepo,
		userRepo:       userRepo,
		userSvc:        userSvc,
		userGroupRepo:  userGroupRepo,
		oidcUserRepo:   oidcUserRepo,
		userFinder:     userFinder,
	}
}

type ListUserQueryParams struct {
	Limit      uint32
	Offset     uint32
	UserIds    []string
	UserEmails []string
	UserName   string
	Status     user.Status
	GroupIds   []string
}

type userUseCase struct {
	db             *sql.DB
	dialectWrapper persistence.DialectWrapper
	transRepo      trans.Repository
	userRepo       user.Repository
	userSvc        user.Service
	userGroupRepo  usergroup.Repository
	oidcUserRepo   oidcUserDomain.Repository
	userFinder     UserFinder
}

func (uc userUseCase) Me(ctx context.Context) (*user.User, error) {
	userId, ok := contextx.GetUserID(ctx)
	if !ok {
		err := fmt.Errorf("user not found in context")
		logger.Errorf("Me: %v", err)
		return nil, err
	}
	u, err := uc.Get(ctx, userId)
	if err != nil {
		logger.Errorf("Me: failed to get user %s: %v", userId, err)
		return nil, err
	}
	return u, nil
}

func (uc userUseCase) Get(ctx context.Context, userId string) (*user.User, error) {
	ug, err := uc.userGroupRepo.FindByUserId(ctx, userId)
	if err != nil {
		logger.Errorf("Get: failed to find user group for user %s: %v", userId, err)
		return nil, err
	}
	u, err := uc.userRepo.FindOne(ctx, userId)
	if err != nil {
		logger.Errorf("Get: failed to find user %s: %v", userId, err)
		return nil, err
	}
	u.SetGroupIds(ug.GroupIds(userId))
	return u, nil
}

func (uc userUseCase) Update(ctx context.Context, u *user.User, password *string) (*user.User, error) {
	// Get the original user to compare email
	originalUser, err := uc.userRepo.FindOne(ctx, u.Id)
	if err != nil {
		logger.Errorf("Update: failed to find original user %s for update: %v", u.Id, err)
		return nil, err
	}

	if err := uc.transRepo.ExecTransWithLock(ctx, func(ctx context.Context) error {
		if err := uc.userRepo.Update(ctx, u); err != nil {
			logger.Errorf("Update: failed to update user %s in DB: %v", u.Id, err)
			return err
		}
		if err := uc.userGroupRepo.Upsert(ctx, u.Id, u.GroupIds); err != nil {
			logger.Errorf("Update: failed to upsert user groups for user %s: %v", u.Id, err)
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// Update email in Keycloak if it's changed
	if originalUser.Email != u.Email {
		if err := uc.oidcUserRepo.UpdateEmail(ctx, u.Id, u.Email); err != nil {
			logger.Errorf("Update: failed to update email for user %s in Keycloak: %v", u.Id, err)
			return nil, fmt.Errorf("failed to update email for user %s: %w", u.Id, err)
		}
	}

	// Update password in Keycloak if it's provided
	if password != nil && *password != "" {
		if err := uc.oidcUserRepo.UpdatePassword(ctx, u.Id, *password); err != nil {
			logger.Errorf("Update: failed to update password for user %s in Keycloak: %v", u.Id, err)
			return nil, fmt.Errorf("failed to update password for user %s: %w", u.Id, err)
		}
	}

	updatedUser, err := uc.userRepo.FindOne(ctx, u.Id)
	if err != nil {
		logger.Errorf("Update: failed to find updated user %s: %v", u.Id, err)
		return nil, err
	}
	return updatedUser, nil
}

func (uc userUseCase) Delete(ctx context.Context, userId string) error {
	if err := uc.transRepo.ExecTransWithLock(ctx, func(ctx context.Context) error {
		u, err := uc.userRepo.FindOne(ctx, userId)
		if err != nil {
			logger.Errorf("Delete: failed to find user %s for deletion: %v", userId, err)
			return err
		}
		u.Status = user.InActive
		if err := uc.userRepo.Update(ctx, u); err != nil {
			logger.Errorf("Delete: failed to update user %s status to inactive: %v", userId, err)
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (uc userUseCase) Create(ctx context.Context, username, email, password string, groupIds []string) (*user.User, error) {
	// Create user in Keycloak first
	keycloakUserId, err := uc.oidcUserRepo.CreateUser(ctx, username, email, password)
	if err != nil {
		logger.Errorf("Create: failed to create user %s in Keycloak: %v", username, err)
		return nil, fmt.Errorf("failed to create user in Keycloak: %w", err)
	}

	// Then create user in DB
	u := &user.User{
		Id:       keycloakUserId, // Use the ID from Keycloak
		Username: username,
		Email:    email,
		Status:   user.Active,
		GroupIds: groupIds,
	}
	if err := uc.userSvc.CreateIfNotExists(ctx, u); err != nil {
		logger.Errorf("Create: failed to create user %s in DB (Keycloak ID: %s): %v", username, keycloakUserId, err)
		// Compensation logic: delete the Keycloak user if DB creation fails
		deleteErr := uc.oidcUserRepo.DeleteUser(ctx, keycloakUserId)
		if deleteErr != nil {
			// If deletion also fails, log it and return both errors
			logger.Errorf("Create: failed to delete Keycloak user %s after DB creation failed: %v", keycloakUserId, deleteErr)
			return nil, fmt.Errorf("user created in Keycloak but failed to create in DB: %w (additionally, failed to delete Keycloak user: %v)", err, deleteErr)
		}
		logger.Errorf("Create: failed to create user in DB (Keycloak user %s was deleted): %v", keycloakUserId, err)
		return nil, fmt.Errorf("failed to create user in DB (Keycloak user was deleted): %w", err)
	}

	return u, nil
}

func (uc userUseCase) List(ctx context.Context, params *ListUserQueryParams) ([]*user.User, int64, error) {
	// Start with a base query for users
	b := uc.dialectWrapper.From("users")

	// Apply filters
	if params.UserIds != nil {
		b = b.Where(goqu.Ex{"users.id": params.UserIds})
	}
	if params.UserEmails != nil {
		b = b.Where(goqu.Ex{"users.email": params.UserEmails})
	}
	if params.UserName != "" {
		b = b.Where(goqu.C("username").Table("users").Like(fmt.Sprintf("%%%s%%", params.UserName)))
	}
	if params.Status != "" {
		b = b.Where(goqu.Ex{"users.status": params.Status})
	}

	// If GroupIds is provided, use EXISTS to filter users who belong to any of the specified groups
	if params.GroupIds != nil {
		// Create a subquery to check if the user belongs to any of the specified groups
		subquery := uc.dialectWrapper.From("user_groups").
			Select(goqu.L("1")).
			Where(goqu.Ex{
				"user_groups.user_id":  goqu.I("users.id"),
				"user_groups.group_id": params.GroupIds,
			})

		// Use EXISTS with the subquery
		b = b.Where(goqu.L("EXISTS ?", subquery))
	}

	// Count total users matching the criteria
	// Since we're using EXISTS instead of JOIN, we don't need DISTINCT anymore
	countQuery := b.Select(goqu.COUNT("users.id"))
	query, queryParams, err := countQuery.Prepared(true).ToSQL()
	if err != nil {
		logger.Errorf("List: failed to build count SQL query: %v", err)
		return nil, 0, err
	}

	var cnt int64
	err = uc.db.QueryRowContext(ctx, query, queryParams...).Scan(&cnt)
	if err != nil {
		logger.Errorf("List: failed to count users: %v", err)
		return nil, 0, err
	}

	// Apply pagination only when limit > 0
	if params.Limit > 0 {
		b = b.Limit(uint(params.Limit)).Offset(uint(params.Offset))
	}

	// Get users
	items, err := uc.list(ctx, b)
	if err != nil {
		logger.Errorf("List: failed to list users: %v", err)
		return nil, 0, err
	}

	// If no users found, return empty result
	if len(items) == 0 {
		return items, cnt, nil
	}

	// Collect all user IDs
	var userIds []string
	for _, item := range items {
		userIds = append(userIds, item.Id)
	}

	// Fetch all user-group relationships in a single query
	userGroupMap := make(map[string][]string)

	// Build a query to get all user-group relationships for the user IDs
	ugQuery := uc.dialectWrapper.From("user_groups").
		Select("user_id", "group_id").
		Where(goqu.Ex{"user_id": userIds})

	ugSQL, ugParams, err := ugQuery.Prepared(true).ToSQL()
	if err != nil {
		logger.Errorf("List: failed to build user-group SQL query: %v", err)
		return nil, 0, err
	}

	// Execute the query
	ugRows, err := uc.db.QueryContext(ctx, ugSQL, ugParams...)
	if err != nil {
		logger.Errorf("List: failed to execute user-group SQL query: %v", err)
		return nil, 0, err
	}
	defer func() {
		err := ugRows.Close()
		if err != nil {
			logger.Errorf("List: error closing user-group rows: %v", err)
		}
	}()

	// Process the results
	for ugRows.Next() {
		var userId, groupId string
		if err := ugRows.Scan(&userId, &groupId); err != nil {
			logger.Errorf("List: failed to scan user-group row: %v", err)
			return nil, 0, err
		}
		userGroupMap[userId] = append(userGroupMap[userId], groupId)
	}

	if err := ugRows.Err(); err != nil {
		logger.Errorf("List: error after iterating user-group rows: %v", err)
		return nil, 0, err
	}

	// Set group IDs for each user
	for _, item := range items {
		if groupIds, ok := userGroupMap[item.Id]; ok {
			item.SetGroupIds(groupIds)
		}
	}

	return items, cnt, nil
}

func (uc userUseCase) list(ctx context.Context, b *goqu.SelectDataset) ([]*user.User, error) {
	query, queryParams, err := b.Select(
		"users.id",
		"users.username",
		"users.email",
		"users.status",
		"users.created_at",
		"users.updated_at",
	).Prepared(true).ToSQL()
	if err != nil {
		logger.Errorf("list: failed to build SQL query: %v", err)
		return nil, err
	}
	rows, err := uc.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		logger.Errorf("list: failed to execute SQL query: %v", err)
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			logger.Errorf("list: error closing rows: %v", err) // Changed to Errorf
		}
	}()
	var items []*user.User
	for rows.Next() {
		var i user.User
		if err := rows.Scan(
			&i.Id,
			&i.Username,
			&i.Email,
			&i.Status,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			logger.Errorf("list: failed to scan row: %v", err)
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		logger.Errorf("list: error after iterating rows: %v", err)
		return nil, err
	}
	return items, nil
}

func (uc userUseCase) AssignGroup(ctx context.Context, userId, groupId string) (*user.User, error) {
	if err := uc.userGroupRepo.AssignGroup(ctx, userId, groupId); err != nil {
		logger.Errorf("AssignGroup: failed to assign group %s to user %s: %v", groupId, userId, err)
		return nil, err
	}
	u, err := uc.Get(ctx, userId)
	if err != nil {
		logger.Errorf("AssignGroup: failed to get user %s after assigning group %s: %v", userId, groupId, err)
		return nil, err
	}
	return u, nil
}

func (uc userUseCase) UnassignGroup(ctx context.Context, userId, groupId string) (*user.User, error) {
	if err := uc.userGroupRepo.UnassignGroup(ctx, userId, groupId); err != nil {
		logger.Errorf("UnassignGroup: failed to unassign group %s from user %s: %v", groupId, userId, err)
		return nil, err
	}
	u, err := uc.Get(ctx, userId)
	if err != nil {
		logger.Errorf("UnassignGroup: failed to get user %s after unassigning group %s: %v", userId, groupId, err)
		return nil, err
	}
	return u, nil
}

func (uc userUseCase) GetMeMenus(ctx context.Context) ([]*menu.Menu, error) {
	return uc.userFinder.GetMeMenus(ctx)
}
