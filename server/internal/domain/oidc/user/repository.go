package user

import "context"

// Repository defines the interface for Keycloak user operations.
type Repository interface {
	UpdatePassword(ctx context.Context, userID string, password string) error
	UpdateEmail(ctx context.Context, userID string, email string) error
	CreateUser(ctx context.Context, username, email, password string) (string, error)
	DeleteUser(ctx context.Context, userID string) error
}
