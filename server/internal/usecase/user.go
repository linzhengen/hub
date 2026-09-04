package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"

	"github.com/linzhengen/hub/server/internal/domain/auth"
	"github.com/linzhengen/hub/server/internal/domain/contextx"
	oidcUserDomain "github.com/linzhengen/hub/server/internal/domain/oidc/user"
	"github.com/linzhengen/hub/server/internal/domain/system/resource/menu"
	"github.com/linzhengen/hub/server/internal/domain/trans"
	"github.com/linzhengen/hub/server/internal/domain/user"
	"github.com/linzhengen/hub/server/internal/domain/user/usergroup"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence/postgres"
	"github.com/linzhengen/hub/server/internal/usecase/pagination"

	"github.com/linzhengen/hub/server/pkg/logger"
)

type UserUseCase interface {
	Me(ctx context.Context) (*user.User, error)
	Get(ctx context.Context, userId string) (*user.User, error)
	Update(ctx context.Context, u *user.User, password *string) (*user.User, error)
	Delete(ctx context.Context, userId string) error
	Create(ctx context.Context, username, email, password string, groupIds []string) (*user.User, error)
	List(ctx context.Context, params *ListUserQueryParams) ([]*user.User, int64, error)
	AddGroups(ctx context.Context, userId string, groupIds []string, expiresAt *time.Time) (*user.User, error)
	RemoveGroups(ctx context.Context, userId string, groupIds []string) (*user.User, error)
	GetMeMenus(ctx context.Context) ([]*menu.Menu, error)
	SendMeVerifyEmail(ctx context.Context) error
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
	authSvc auth.Service,
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
		authSvc:        authSvc,
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
	authSvc        auth.Service
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

// SendMeVerifyEmail asks the OIDC provider to resend the address-verification
// email to the caller. The target is always the authenticated user, so no user
// id is accepted from the request.
func (uc userUseCase) SendMeVerifyEmail(ctx context.Context) error {
	userId, ok := contextx.GetUserID(ctx)
	if !ok {
		err := fmt.Errorf("user not found in context")
		logger.Errorf("SendMeVerifyEmail: %v", err)
		return err
	}
	if err := uc.oidcUserRepo.SendVerifyEmail(ctx, userId); err != nil {
		logger.Errorf("SendMeVerifyEmail: failed to send verify email to user %s: %v", userId, err)
		return err
	}
	return nil
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
	memberships := make([]user.GroupMembership, 0, len(ug))
	for _, g := range ug {
		if g.UserId != userId {
			continue
		}
		memberships = append(memberships, user.GroupMembership{
			GroupId:   g.GroupId,
			ExpiresAt: g.ExpiresAt,
		})
	}
	u.SetGroups(memberships)
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
		if err := uc.userGroupRepo.Upsert(ctx, u.Id, u.GroupIds()); err != nil {
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
		if err := uc.userRepo.Delete(ctx, userId); err != nil {
			logger.Errorf("Delete: failed to delete user %s from DB: %v", userId, err)
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	if err := uc.oidcUserRepo.DeleteUser(ctx, userId); err != nil {
		logger.Errorf("Delete: failed to delete user %s from Keycloak: %v", userId, err)
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
		// The groups a user is created in are a set of ids with no term, so
		// they do not expire. A grant with a term arrives later, through
		// AddGroupsToUser or an approved request.
		Groups: user.PermanentMemberships(groupIds),
	}
	if _, err := uc.userSvc.CreateIfNotExists(ctx, u); err != nil {
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

// callerId is who is asking, empty when nobody is - which VisibleOrgs answers
// with an empty scope rather than an error, so a listing shows nothing instead
// of everything.
func callerId(ctx context.Context) string {
	id, _ := contextx.GetUserID(ctx)
	return id
}

func (uc userUseCase) List(ctx context.Context, params *ListUserQueryParams) ([]*user.User, int64, error) {
	// A user is visible where they hold a place: a directory of everybody on the
	// installation is not one tenant's to read.
	//
	// The consequence is that somebody who belongs to no group is visible only
	// to the platform, which is where a brand-new account starts. That is the
	// right way round - a tenant admin sees the people in their tenant - but it
	// is the reason a new user does not appear in another tenant's picker.
	scope, err := uc.authSvc.VisibleOrgs(ctx, callerId(ctx), contextx.GetOrgID(ctx))
	if err != nil {
		return nil, 0, err
	}
	if scope.Empty() {
		return nil, 0, nil
	}

	// Start with a base query for users
	b := uc.dialectWrapper.From("users")

	if !scope.All {
		visible := uc.dialectWrapper.From("user_groups").
			Join(goqu.I("groups"), goqu.On(goqu.I("groups.id").Eq(goqu.I("user_groups.group_id")))).
			Select(goqu.L("1")).
			Where(goqu.Ex{
				"user_groups.user_id": goqu.I("users.id"),
				"groups.org_id":       scope.OrgIds,
			})
		b = b.Where(goqu.L("EXISTS ?", visible))
	}

	// Apply filters
	if params.UserIds != nil {
		b = b.Where(postgres.In("users.id", params.UserIds))
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
			Where(postgres.In("user_groups.group_id", params.GroupIds)).
			Where(goqu.Ex{
				"user_groups.user_id": goqu.I("users.id"),
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

	page := pagination.New(params.Limit, params.Offset)
	b = b.Limit(page.Limit()).Offset(page.Offset())

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
	userGroupMap := make(map[string][]user.GroupMembership)

	// Build a query to get all user-group relationships for the user IDs
	ugQuery := uc.dialectWrapper.From("user_groups").
		Select("user_id", "group_id", "expires_at").
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
		var expiresAt sql.NullTime
		if err := ugRows.Scan(&userId, &groupId, &expiresAt); err != nil {
			logger.Errorf("List: failed to scan user-group row: %v", err)
			return nil, 0, err
		}
		membership := user.GroupMembership{GroupId: groupId}
		if expiresAt.Valid {
			membership.ExpiresAt = &expiresAt.Time
		}
		userGroupMap[userId] = append(userGroupMap[userId], membership)
	}

	if err := ugRows.Err(); err != nil {
		logger.Errorf("List: error after iterating user-group rows: %v", err)
		return nil, 0, err
	}

	// Set the memberships for each user
	for _, item := range items {
		if memberships, ok := userGroupMap[item.Id]; ok {
			item.SetGroups(memberships)
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

// AddGroups puts the user in each of groupIds, leaving the groups they are
// already in alone. The whole set is applied in one transaction so a partial
// failure does not leave the user in some of the groups but not the rest.
//
// expiresAt applies to every group in the call and is nil for a membership
// that does not end.
func (uc userUseCase) AddGroups(
	ctx context.Context,
	userId string,
	groupIds []string,
	expiresAt *time.Time,
) (*user.User, error) {
	return uc.changeGroups(ctx, "AddGroups", userId, groupIds,
		func(ctx context.Context, userId, groupId string) error {
			return uc.userGroupRepo.AssignGroup(ctx, userId, groupId, expiresAt)
		})
}

// RemoveGroups takes the user out of each of groupIds.
func (uc userUseCase) RemoveGroups(ctx context.Context, userId string, groupIds []string) (*user.User, error) {
	return uc.changeGroups(ctx, "RemoveGroups", userId, groupIds, uc.userGroupRepo.UnassignGroup)
}

func (uc userUseCase) changeGroups(
	ctx context.Context,
	operation string,
	userId string,
	groupIds []string,
	apply func(ctx context.Context, userId, groupId string) error,
) (*user.User, error) {
	if err := uc.transRepo.ExecTrans(ctx, func(ctx context.Context) error {
		for _, groupId := range groupIds {
			if err := apply(ctx, userId, groupId); err != nil {
				logger.Errorf("%s: failed for group %s and user %s: %v", operation, groupId, userId, err)
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	u, err := uc.Get(ctx, userId)
	if err != nil {
		logger.Errorf("%s: failed to get user %s afterwards: %v", operation, userId, err)
		return nil, err
	}
	return u, nil
}

func (uc userUseCase) GetMeMenus(ctx context.Context) ([]*menu.Menu, error) {
	return uc.userFinder.GetMeMenus(ctx)
}
