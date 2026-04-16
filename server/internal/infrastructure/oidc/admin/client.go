package admin

import (
	"context"
	"sync"
	"time"

	"github.com/Nerzal/gocloak/v13"

	"github.com/linzhengen/hub/v1/server/config"

	"github.com/linzhengen/hub/v1/server/pkg/logger"
)

// Client is a thread-safe client for performing Keycloak admin operations.
// It automatically manages the admin access token.
type Client interface {
	SetUserPassword(ctx context.Context, userID, password string) error
	UpdateEmail(ctx context.Context, userID, email string) error
	CreateUser(ctx context.Context, username, email, password string) (string, error)
	DeleteUser(ctx context.Context, userID string) error
}

// GoCloak defines the interface for Keycloak operations
// This makes it easier to test without mocks by allowing alternative implementations
type GoCloak interface {
	LoginAdmin(ctx context.Context, username, password, realm string) (*gocloak.JWT, error)
	SetPassword(ctx context.Context, token, userID, realm, password string, temporary bool) error
	CreateUser(ctx context.Context, token, realm string, user gocloak.User) (string, error)
	DeleteUser(ctx context.Context, token, realm, userID string) error
	UpdateUser(ctx context.Context, token, realm string, user gocloak.User) error
}

// RealGoCloak wraps the actual gocloak.GoCloak client
type RealGoCloak struct {
	client *gocloak.GoCloak
}

// NewRealGoCloak creates a new wrapper around the real gocloak client
func NewRealGoCloak(url string) GoCloak {
	return &RealGoCloak{
		client: gocloak.NewClient(url),
	}
}

// LoginAdmin implements the GoCloak interface
func (g *RealGoCloak) LoginAdmin(ctx context.Context, username, password, realm string) (*gocloak.JWT, error) {
	return g.client.LoginAdmin(ctx, username, password, realm)
}

// SetPassword implements the GoCloak interface
func (g *RealGoCloak) SetPassword(ctx context.Context, token, userID, realm, password string, temporary bool) error {
	return g.client.SetPassword(ctx, token, userID, realm, password, temporary)
}

// CreateUser implements the GoCloak interface
func (g *RealGoCloak) CreateUser(ctx context.Context, token, realm string, user gocloak.User) (string, error) {
	return g.client.CreateUser(ctx, token, realm, user)
}

// DeleteUser implements the GoCloak interface
func (g *RealGoCloak) DeleteUser(ctx context.Context, token, realm, userID string) error {
	return g.client.DeleteUser(ctx, token, realm, userID)
}

// UpdateUser implements the GoCloak interface
func (g *RealGoCloak) UpdateUser(ctx context.Context, token, realm string, user gocloak.User) error {
	return g.client.UpdateUser(ctx, token, realm, user)
}

type client struct {
	gocloak      GoCloak
	cfg          config.KeyCloak
	token        *gocloak.JWT
	tokenMutex   sync.RWMutex
	tokenExpires time.Time
}

// NewClient creates a new Keycloak admin client.
func NewClient(cfg config.KeyCloak) Client {
	return &client{
		gocloak: NewRealGoCloak(cfg.KeycloakURL),
		cfg:     cfg,
	}
}

// NewClientWithGoCloak creates a new client with a custom GoCloak implementation
// This is useful for testing without mocks
func NewClientWithGoCloak(gocloak GoCloak, cfg config.KeyCloak) Client {
	return &client{
		gocloak: gocloak,
		cfg:     cfg,
	}
}

// login performs login for the admin user and stores the token.
// Note: This method does not handle locking, the caller must handle it.
func (c *client) login(ctx context.Context) error {
	logger.Infof("Logging in as Keycloak admin")
	token, err := c.gocloak.LoginAdmin(ctx, c.cfg.AdminUser, c.cfg.AdminPass, c.cfg.AdminRealm)
	if err != nil {
		return err
	}
	logger.Infof("Successfully logged in as Keycloak admin")

	c.token = token
	// Refresh the token a bit before it actually expires.
	c.tokenExpires = time.Now().Add(time.Duration(token.ExpiresIn-10) * time.Second)
	return nil
}

// getToken returns a valid admin access token, refreshing it if necessary.
func (c *client) getToken(ctx context.Context) (string, error) {
	// First try with a read lock
	c.tokenMutex.RLock()
	if c.token != nil && !time.Now().After(c.tokenExpires) {
		// Token is valid, use it
		token := c.token.AccessToken
		c.tokenMutex.RUnlock()
		return token, nil
	}
	c.tokenMutex.RUnlock() // Release read lock before acquiring write lock

	// Need to refresh token - acquire write lock
	c.tokenMutex.Lock()

	// Double check after acquiring write lock
	if c.token != nil && !time.Now().After(c.tokenExpires) {
		// Another goroutine refreshed the token while we were waiting
		token := c.token.AccessToken
		c.tokenMutex.Unlock()
		return token, nil
	}

	// Need to refresh token
	err := c.login(ctx)
	if err != nil {
		c.tokenMutex.Unlock()
		return "", err
	}

	// Get the token before unlocking
	token := c.token.AccessToken
	c.tokenMutex.Unlock()

	return token, nil
}

// SetUserPassword sets a new password for the given user ID in Keycloak.
func (c *client) SetUserPassword(ctx context.Context, userID, password string) error {
	token, err := c.getToken(ctx)
	if err != nil {
		return err
	}

	return c.gocloak.SetPassword(ctx, token, userID, c.cfg.Realm, password, false)
}

// CreateUser creates a new user in Keycloak and returns the user ID.
func (c *client) CreateUser(ctx context.Context, username, email, password string) (string, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return "", err
	}

	user := gocloak.User{
		Username: &username,
		Email:    &email,
		Enabled:  gocloak.BoolP(true),
	}

	userID, err := c.gocloak.CreateUser(ctx, token, c.cfg.Realm, user)
	if err != nil {
		return "", err
	}

	if password != "" {
		if err := c.SetUserPassword(ctx, userID, password); err != nil {
			return "", err
		}
	}

	return userID, nil
}

// DeleteUser deletes a user from Keycloak.
func (c *client) DeleteUser(ctx context.Context, userID string) error {
	token, err := c.getToken(ctx)
	if err != nil {
		return err
	}

	return c.gocloak.DeleteUser(ctx, token, c.cfg.Realm, userID)
}

// UpdateEmail updates the user's email in Keycloak.
func (c *client) UpdateEmail(ctx context.Context, userID, email string) error {
	token, err := c.getToken(ctx)
	if err != nil {
		return err
	}

	// Create a user object with the updated email
	user := gocloak.User{
		ID:    &userID,
		Email: &email,
	}

	// Update the user in Keycloak
	return c.gocloak.UpdateUser(ctx, token, c.cfg.Realm, user)
}
