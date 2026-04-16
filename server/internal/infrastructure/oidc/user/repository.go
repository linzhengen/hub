package user

import (
	"context"

	"github.com/linzhengen/hub/v1/server/internal/domain/oidc/user"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/oidc/admin"
)

type repository struct {
	adminClient admin.Client
}

// NewRepository creates a new user repository for Keycloak operations.
func NewRepository(adminClient admin.Client) user.Repository {
	return &repository{
		adminClient: adminClient,
	}
}

// UpdatePassword updates the user's password in Keycloak.
func (r *repository) UpdatePassword(ctx context.Context, userID string, password string) error {
	return r.adminClient.SetUserPassword(ctx, userID, password)
}

// UpdateEmail updates the user's email in Keycloak.
func (r *repository) UpdateEmail(ctx context.Context, userID string, email string) error {
	return r.adminClient.UpdateEmail(ctx, userID, email)
}

// CreateUser creates a new user in Keycloak and returns the user ID.
func (r *repository) CreateUser(ctx context.Context, username, email, password string) (string, error) {
	return r.adminClient.CreateUser(ctx, username, email, password)
}

// DeleteUser deletes a user from Keycloak.
func (r *repository) DeleteUser(ctx context.Context, userID string) error {
	return r.adminClient.DeleteUser(ctx, userID)
}
