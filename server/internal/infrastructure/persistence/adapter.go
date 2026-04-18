package persistence

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/sqlc-dev/pqtype"

	postgressqlc "github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence/postgres/sqlc"
)

// PostgreSQLQuerier implements Querier for PostgreSQL
type PostgreSQLQuerier struct {
	q *postgressqlc.Queries
}

// NewPostgreSQLQuerier creates a new PostgreSQLQuerier
func NewPostgreSQLQuerier(q *postgressqlc.Queries) *PostgreSQLQuerier {
	return &PostgreSQLQuerier{q: q}
}

func (p *PostgreSQLQuerier) WithTx(tx *sql.Tx) Querier {
	return &PostgreSQLQuerier{q: p.q.WithTx(tx)}
}

func (p *PostgreSQLQuerier) SelectUserById(ctx context.Context, id string) (*UserModel, error) {
	user, err := p.q.SelectUserById(ctx, id)
	if err != nil {
		return nil, err
	}
	return &UserModel{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (p *PostgreSQLQuerier) SelectUserForUpdate(ctx context.Context, id string) (*UserModel, error) {
	user, err := p.q.SelectUserForUpdate(ctx, id)
	if err != nil {
		return nil, err
	}
	return &UserModel{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (p *PostgreSQLQuerier) CreateUser(ctx context.Context, id, username, email, status string) error {
	return p.q.CreateUser(ctx, postgressqlc.CreateUserParams{
		ID:       id,
		Username: username,
		Email:    email,
		Status:   status,
	})
}

func (p *PostgreSQLQuerier) UpdateUser(ctx context.Context, id, username, email, status string) error {
	return p.q.UpdateUser(ctx, postgressqlc.UpdateUserParams{
		ID:       id,
		Username: username,
		Email:    email,
		Status:   status,
	})
}

func (p *PostgreSQLQuerier) DeleteUser(ctx context.Context, id string) error {
	return p.q.DeleteUser(ctx, id)
}

func (p *PostgreSQLQuerier) SelectGroupById(ctx context.Context, id string) (*GroupModel, error) {
	g, err := p.q.SelectGroupById(ctx, id)
	if err != nil {
		return nil, err
	}
	return &GroupModel{
		ID:          g.ID,
		Name:        g.Name,
		Description: g.Description,
		Status:      g.Status,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}, nil
}

func (p *PostgreSQLQuerier) SelectGroupForUpdate(ctx context.Context, id string) (*GroupModel, error) {
	g, err := p.q.SelectGroupForUpdate(ctx, id)
	if err != nil {
		return nil, err
	}
	return &GroupModel{
		ID:          g.ID,
		Name:        g.Name,
		Description: g.Description,
		Status:      g.Status,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}, nil
}

func (p *PostgreSQLQuerier) CreateGroup(ctx context.Context, id, name, status, description string) error {
	return p.q.CreateGroup(ctx, postgressqlc.CreateGroupParams{
		ID:          id,
		Name:        name,
		Status:      status,
		Description: description,
	})
}

func (p *PostgreSQLQuerier) UpdateGroup(ctx context.Context, id, name, status, description string) error {
	return p.q.UpdateGroup(ctx, postgressqlc.UpdateGroupParams{
		ID:          id,
		Name:        name,
		Status:      status,
		Description: description,
	})
}

func (p *PostgreSQLQuerier) DeleteGroup(ctx context.Context, id string) error {
	return p.q.DeleteGroup(ctx, id)
}

func (p *PostgreSQLQuerier) SelectPermissionById(ctx context.Context, id string) (*PermissionModel, error) {
	perm, err := p.q.SelectPermissionById(ctx, id)
	if err != nil {
		return nil, err
	}
	return &PermissionModel{
		ID:          perm.ID,
		Verb:        perm.Verb,
		ResourceID:  perm.ResourceID,
		Description: perm.Description,
		CreatedAt:   perm.CreatedAt,
		UpdatedAt:   perm.UpdatedAt,
	}, nil
}

func (p *PostgreSQLQuerier) SelectPermissionForUpdate(ctx context.Context, id string) (*PermissionModel, error) {
	perm, err := p.q.SelectPermissionForUpdate(ctx, id)
	if err != nil {
		return nil, err
	}
	return &PermissionModel{
		ID:          perm.ID,
		Verb:        perm.Verb,
		ResourceID:  perm.ResourceID,
		Description: perm.Description,
		CreatedAt:   perm.CreatedAt,
		UpdatedAt:   perm.UpdatedAt,
	}, nil
}

func (p *PostgreSQLQuerier) SelectPermissionByResourceId(ctx context.Context, resourceID string) ([]*PermissionModel, error) {
	ps, err := p.q.SelectPermissionByResourceId(ctx, resourceID)
	if err != nil {
		return nil, err
	}
	var res []*PermissionModel
	for _, p := range ps {
		res = append(res, &PermissionModel{
			ID:          p.ID,
			Verb:        p.Verb,
			ResourceID:  p.ResourceID,
			Description: p.Description,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		})
	}
	return res, nil
}

func (p *PostgreSQLQuerier) CreatePermission(ctx context.Context, id, verb, resourceID, description string) error {
	return p.q.CreatePermission(ctx, postgressqlc.CreatePermissionParams{
		ID:          id,
		Verb:        verb,
		ResourceID:  resourceID,
		Description: description,
	})
}

func (p *PostgreSQLQuerier) UpdatePermission(ctx context.Context, id, verb, resourceID, description string) error {
	return p.q.UpdatePermission(ctx, postgressqlc.UpdatePermissionParams{
		ID:          id,
		Verb:        verb,
		ResourceID:  resourceID,
		Description: description,
	})
}

func (p *PostgreSQLQuerier) DeletePermission(ctx context.Context, id string) error {
	return p.q.DeletePermissions(ctx, id)
}

func (p *PostgreSQLQuerier) SelectUserAuthorizedPolicies(ctx context.Context, userID string) ([]*UserAuthorizedPolicyModel, error) {
	rows, err := p.q.SelectUserAuthorizedPolices(ctx, userID)
	if err != nil {
		return nil, err
	}
	res := make([]*UserAuthorizedPolicyModel, len(rows))
	for i, row := range rows {
		res[i] = &UserAuthorizedPolicyModel{
			ID:         row.ID,
			Identifier: row.Identifier,
			Verb:       row.Verb,
		}
	}
	return res, nil
}

func (p *PostgreSQLQuerier) SelectRoleById(ctx context.Context, id string) (*RoleModel, error) {
	r, err := p.q.SelectRoleById(ctx, id)
	if err != nil {
		return nil, err
	}
	return &RoleModel{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}, nil
}

func (p *PostgreSQLQuerier) SelectRoleForUpdate(ctx context.Context, id string) (*RoleModel, error) {
	r, err := p.q.SelectRoleForUpdate(ctx, id)
	if err != nil {
		return nil, err
	}
	return &RoleModel{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}, nil
}

func (p *PostgreSQLQuerier) CreateRole(ctx context.Context, id, name, description string) error {
	return p.q.CreateRole(ctx, postgressqlc.CreateRoleParams{
		ID:          id,
		Name:        name,
		Description: description,
	})
}

func (p *PostgreSQLQuerier) UpdateRole(ctx context.Context, id, name, description string) error {
	return p.q.UpdateRole(ctx, postgressqlc.UpdateRoleParams{
		ID:          id,
		Name:        name,
		Description: description,
	})
}

func (p *PostgreSQLQuerier) DeleteRole(ctx context.Context, id string) error {
	return p.q.DeleteRole(ctx, id)
}

func (p *PostgreSQLQuerier) SelectResourceById(ctx context.Context, id string) (*ResourceModel, error) {
	r, err := p.q.SelectResourceById(ctx, id)
	if err != nil {
		return nil, err
	}
	var metadata map[string]string
	if r.Metadata.Valid {
		_ = json.Unmarshal(r.Metadata.RawMessage, &metadata)
	}
	return &ResourceModel{
		ID:           r.ID,
		ParentID:     r.ParentID,
		Name:         r.Name,
		Identifier:   r.Identifier,
		Type:         r.Type,
		Path:         r.Path.String,
		Component:    r.Component.String,
		DisplayOrder: r.DisplayOrder.Int32,
		Status:       r.Status,
		Description:  r.Description,
		Metadata:     metadata,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}, nil
}

func (p *PostgreSQLQuerier) SelectResourceByIdentifier(ctx context.Context, identifier string) (*ResourceModel, error) {
	r, err := p.q.SelectResourceByIdentifier(ctx, identifier)
	if err != nil {
		return nil, err
	}
	var metadata map[string]string
	if r.Metadata.Valid {
		_ = json.Unmarshal(r.Metadata.RawMessage, &metadata)
	}
	return &ResourceModel{
		ID:           r.ID,
		ParentID:     r.ParentID,
		Name:         r.Name,
		Identifier:   r.Identifier,
		Type:         r.Type,
		Path:         r.Path.String,
		Component:    r.Component.String,
		DisplayOrder: r.DisplayOrder.Int32,
		Status:       r.Status,
		Description:  r.Description,
		Metadata:     metadata,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}, nil
}

func (p *PostgreSQLQuerier) SelectResourceForUpdate(ctx context.Context, id string) (*ResourceModel, error) {
	r, err := p.q.SelectResourceForUpdate(ctx, id)
	if err != nil {
		return nil, err
	}
	var metadata map[string]string
	if r.Metadata.Valid {
		_ = json.Unmarshal(r.Metadata.RawMessage, &metadata)
	}
	return &ResourceModel{
		ID:           r.ID,
		ParentID:     r.ParentID,
		Name:         r.Name,
		Identifier:   r.Identifier,
		Type:         r.Type,
		Path:         r.Path.String,
		Component:    r.Component.String,
		DisplayOrder: r.DisplayOrder.Int32,
		Status:       r.Status,
		Description:  r.Description,
		Metadata:     metadata,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}, nil
}

func (p *PostgreSQLQuerier) CreateResource(ctx context.Context, id, parentID, name, identifier, resourceType, path, component, status string, displayOrder int32, description string, metadata map[string]string) error {
	m, _ := json.Marshal(metadata)
	return p.q.CreateResource(ctx, postgressqlc.CreateResourceParams{
		ID:           id,
		ParentID:     parentID,
		Name:         name,
		Identifier:   identifier,
		Type:         resourceType,
		Path:         sql.NullString{String: path, Valid: path != ""},
		Component:    sql.NullString{String: component, Valid: component != ""},
		Status:       status,
		DisplayOrder: sql.NullInt32{Int32: displayOrder, Valid: true},
		Description:  description,
		Metadata:     pqtype.NullRawMessage{RawMessage: m, Valid: m != nil},
	})
}

func (p *PostgreSQLQuerier) UpdateResource(ctx context.Context, id, parentID, name, identifier, resourceType, path, component, status string, displayOrder int32, description string, metadata map[string]string) error {
	m, _ := json.Marshal(metadata)
	return p.q.UpdateResource(ctx, postgressqlc.UpdateResourceParams{
		ID:           id,
		ParentID:     parentID,
		Name:         name,
		Identifier:   identifier,
		Type:         resourceType,
		Path:         sql.NullString{String: path, Valid: path != ""},
		Component:    sql.NullString{String: component, Valid: component != ""},
		Status:       status,
		DisplayOrder: sql.NullInt32{Int32: displayOrder, Valid: true},
		Description:  description,
		Metadata:     pqtype.NullRawMessage{RawMessage: m, Valid: m != nil},
	})
}

func (p *PostgreSQLQuerier) DeleteResource(ctx context.Context, id string) error {
	return p.q.DeleteResource(ctx, id)
}

func (p *PostgreSQLQuerier) AddPermissionToRole(ctx context.Context, roleID, permissionID string) error {
	return p.q.AddPermissionToRole(ctx, postgressqlc.AddPermissionToRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
}

func (p *PostgreSQLQuerier) RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) error {
	return p.q.RemovePermissionFromRole(ctx, postgressqlc.RemovePermissionFromRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
}

func (p *PostgreSQLQuerier) IsPermissionInRole(ctx context.Context, roleID, permissionID string) (bool, error) {
	return p.q.IsPermissionInRole(ctx, postgressqlc.IsPermissionInRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
}

func (p *PostgreSQLQuerier) DeleteRoleAllPermission(ctx context.Context, roleID string) error {
	return p.q.DeleteRoleAllPermission(ctx, roleID)
}

func (p *PostgreSQLQuerier) SelectRolePermissionByRoleId(ctx context.Context, roleID string) ([]*RolePermissionModel, error) {
	rows, err := p.q.SelectRolePermissionByRoleId(ctx, roleID)
	if err != nil {
		return nil, err
	}
	res := make([]*RolePermissionModel, len(rows))
	for i, row := range rows {
		res[i] = &RolePermissionModel{
			RoleID:       row.RoleID,
			PermissionID: row.PermissionID,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
		}
	}
	return res, nil
}

func (p *PostgreSQLQuerier) CreateUserGroup(ctx context.Context, userID, groupID string) error {
	return p.q.CreateUserGroup(ctx, postgressqlc.CreateUserGroupParams{
		UserID:  userID,
		GroupID: groupID,
	})
}

func (p *PostgreSQLQuerier) DeleteUserGroup(ctx context.Context, userID, groupID string) error {
	return p.q.DeleteUserGroup(ctx, postgressqlc.DeleteUserGroupParams{
		UserID:  userID,
		GroupID: groupID,
	})
}

func (p *PostgreSQLQuerier) DeleteUserAllGroup(ctx context.Context, userID string) error {
	return p.q.DeleteUserAllGroup(ctx, userID)
}

func (p *PostgreSQLQuerier) IsUserInGroup(ctx context.Context, userID, groupID string) (bool, error) {
	return p.q.IsUserInGroup(ctx, postgressqlc.IsUserInGroupParams{
		UserID:  userID,
		GroupID: groupID,
	})
}

func (p *PostgreSQLQuerier) SelectUserGroupByUserId(ctx context.Context, userID string) ([]*UserGroupModel, error) {
	rows, err := p.q.SelectUserGroupByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}
	res := make([]*UserGroupModel, len(rows))
	for i, row := range rows {
		res[i] = &UserGroupModel{
			UserID:    row.UserID,
			GroupID:   row.GroupID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		}
	}
	return res, nil
}

func (p *PostgreSQLQuerier) CreateGroupRole(ctx context.Context, groupID, roleID string) error {
	return p.q.CreateGroupRole(ctx, postgressqlc.CreateGroupRoleParams{
		GroupID: groupID,
		RoleID:  roleID,
	})
}

func (p *PostgreSQLQuerier) DeleteGroupRole(ctx context.Context, groupID, roleID string) error {
	return p.q.DeleteGroupRole(ctx, postgressqlc.DeleteGroupRoleParams{
		GroupID: groupID,
		RoleID:  roleID,
	})
}

func (p *PostgreSQLQuerier) DeleteGroupAllRole(ctx context.Context, groupID string) error {
	return p.q.DeleteGroupAllRole(ctx, groupID)
}

func (p *PostgreSQLQuerier) SelectGroupRoleByGroupId(ctx context.Context, groupID string) ([]*GroupRoleModel, error) {
	rows, err := p.q.SelectGroupRoleByGroupId(ctx, groupID)
	if err != nil {
		return nil, err
	}
	res := make([]*GroupRoleModel, len(rows))
	for i, row := range rows {
		res[i] = &GroupRoleModel{
			GroupID:   row.GroupID,
			RoleID:    row.RoleID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		}
	}
	return res, nil
}

// NewQuerierAdapter creates the appropriate Querier adapter
func NewQuerierAdapter(queries interface{}) Querier {
	if q, ok := queries.(*postgressqlc.Queries); ok {
		return NewPostgreSQLQuerier(q)
	}
	return nil
}
