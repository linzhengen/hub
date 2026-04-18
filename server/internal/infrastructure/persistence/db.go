package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/doug-martin/goqu/v9"

	"github.com/linzhengen/hub/v1/server/config"
	"github.com/linzhengen/hub/v1/server/internal/domain/contextx"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence/postgres"
	"github.com/linzhengen/hub/v1/server/pkg/logger"
)

// DialectWrapper is a type alias for goqu.DialectWrapper
type DialectWrapper = goqu.DialectWrapper

// NewConnection creates a new database connection based on the configuration
func NewConnection(cfg config.EnvConfig) (*sql.DB, error) {
	logger.Info("Using PostgreSQL database")
	return postgres.NewConn(cfg.PostgreSQL)
}

// GetQuerier returns the appropriate Querier based on the database type
func GetQuerier(db *sql.DB) interface{} {
	return postgres.NewQuerier(db)
}

// UserModel represents a user in the database
type UserModel struct {
	ID        string
	Username  string
	Email     string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// GroupModel represents a group in the database
type GroupModel struct {
	ID          string
	Name        string
	Description string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PermissionModel represents a permission in the database
type PermissionModel struct {
	ID          string
	Verb        string
	ResourceID  string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Querier is a common interface for database operations
type Querier interface {
	WithTx(tx *sql.Tx) Querier

	// User operations
	SelectUserById(ctx context.Context, id string) (*UserModel, error)
	SelectUserForUpdate(ctx context.Context, id string) (*UserModel, error)
	CreateUser(ctx context.Context, id, username, email, status string) error
	UpdateUser(ctx context.Context, id, username, email, status string) error
	DeleteUser(ctx context.Context, id string) error

	// Group operations
	SelectGroupById(ctx context.Context, id string) (*GroupModel, error)
	SelectGroupForUpdate(ctx context.Context, id string) (*GroupModel, error)
	CreateGroup(ctx context.Context, id, name, status, description string) error
	UpdateGroup(ctx context.Context, id, name, status, description string) error
	DeleteGroup(ctx context.Context, id string) error

	// Permission operations
	SelectPermissionById(ctx context.Context, id string) (*PermissionModel, error)
	SelectPermissionForUpdate(ctx context.Context, id string) (*PermissionModel, error)
	SelectPermissionByResourceId(ctx context.Context, resourceID string) ([]*PermissionModel, error)
	CreatePermission(ctx context.Context, id, verb, resourceID, description string) error
	UpdatePermission(ctx context.Context, id, verb, resourceID, description string) error
	DeletePermission(ctx context.Context, id string) error

	// Add other operations as needed...
	SelectUserAuthorizedPolicies(ctx context.Context, userID string) ([]*UserAuthorizedPolicyModel, error)

	// Role operations
	SelectRoleById(ctx context.Context, id string) (*RoleModel, error)
	SelectRoleForUpdate(ctx context.Context, id string) (*RoleModel, error)
	CreateRole(ctx context.Context, id, name, description string) error
	UpdateRole(ctx context.Context, id, name, description string) error
	DeleteRole(ctx context.Context, id string) error

	// Resource operations
	SelectResourceById(ctx context.Context, id string) (*ResourceModel, error)
	SelectResourceByIdentifier(ctx context.Context, identifier string) (*ResourceModel, error)
	SelectResourceForUpdate(ctx context.Context, id string) (*ResourceModel, error)
	CreateResource(ctx context.Context, id, parentID, name, identifier, resourceType, path, component, status string, displayOrder int32, description string, metadata map[string]string) error
	UpdateResource(ctx context.Context, id, parentID, name, identifier, resourceType, path, component, status string, displayOrder int32, description string, metadata map[string]string) error
	DeleteResource(ctx context.Context, id string) error

	// RolePermission operations
	AddPermissionToRole(ctx context.Context, roleID, permissionID string) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) error
	DeleteRoleAllPermission(ctx context.Context, roleID string) error
	IsPermissionInRole(ctx context.Context, roleID, permissionID string) (bool, error)
	SelectRolePermissionByRoleId(ctx context.Context, roleID string) ([]*RolePermissionModel, error)

	// UserGroup operations
	CreateUserGroup(ctx context.Context, userID, groupID string) error
	DeleteUserGroup(ctx context.Context, userID, groupID string) error
	DeleteUserAllGroup(ctx context.Context, userID string) error
	IsUserInGroup(ctx context.Context, userID, groupID string) (bool, error)
	SelectUserGroupByUserId(ctx context.Context, userID string) ([]*UserGroupModel, error)

	// GroupRole operations
	CreateGroupRole(ctx context.Context, groupID, roleID string) error
	DeleteGroupRole(ctx context.Context, groupID, roleID string) error
	DeleteGroupAllRole(ctx context.Context, groupID string) error
	SelectGroupRoleByGroupId(ctx context.Context, groupID string) ([]*GroupRoleModel, error)
}

// RoleModel represents a role in the database
type RoleModel struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ResourceModel represents a resource in the database
type ResourceModel struct {
	ID           string
	ParentID     string
	Name         string
	Identifier   string
	Type         string
	Path         string
	Component    string
	DisplayOrder int32
	Status       string
	Description  string
	Metadata     map[string]string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserGroupModel represents a user-group relationship
type UserGroupModel struct {
	UserID    string
	GroupID   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// GroupRoleModel represents a group-role relationship
type GroupRoleModel struct {
	GroupID   string
	RoleID    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// RolePermissionModel represents a role-permission relationship
type RolePermissionModel struct {
	RoleID       string
	PermissionID string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserAuthorizedPolicyModel represents an authorized policy for a user
type UserAuthorizedPolicyModel struct {
	ID         string
	Identifier string
	Verb       string
}

// GetQ returns the appropriate Querier from context or the default one
func GetQ(ctx context.Context, q Querier) Querier {
	if trTX, ok := contextx.FromTrans(ctx); ok {
		if tq, ok := trTX.(Querier); ok {
			return tq
		}
	}
	return q
}
