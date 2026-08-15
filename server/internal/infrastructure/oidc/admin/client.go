package admin

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Nerzal/gocloak/v13"

	"github.com/linzhengen/hub/server/config"

	"github.com/linzhengen/hub/server/pkg/logger"
)

// Client is a thread-safe client for performing Keycloak admin operations.
// It automatically manages the admin access token.
type Client interface {
	SetUserPassword(ctx context.Context, userID, password string) error
	UpdateEmail(ctx context.Context, userID, email string) error
	CreateUser(ctx context.Context, username, email, password string) (string, error)
	DeleteUser(ctx context.Context, userID string) error
	SendVerifyEmail(ctx context.Context, userID string) error

	// CreateServiceAccountClient registers a confidential client whose only
	// purpose is to authenticate a machine, and returns the handles hub needs
	// to keep track of it: Keycloak's internal id for the client, the id of the
	// user the client acts as, and the secret.
	//
	// The secret is returned once, here, and never read back. Keycloak can
	// re-issue one but cannot show the current one again in every
	// configuration, so treating it as write-once is the behaviour that holds
	// everywhere rather than the one that happens to work locally.
	CreateServiceAccountClient(ctx context.Context, clientID, description string) (ServiceAccountClient, error)
	// RotateClientSecret issues a new secret and invalidates the old one.
	RotateClientSecret(ctx context.Context, keycloakID string) (string, error)
	DeleteClient(ctx context.Context, keycloakID string) error
}

// ServiceAccountClient is what registering a machine's client yields.
type ServiceAccountClient struct {
	// KeycloakID is Keycloak's internal handle for the client, which every
	// later admin call needs.
	KeycloakID string
	// UserID is the client's own service account user. hub stores it in
	// `users`, which is what lets a machine hold groups and roles and appear in
	// the audit log the way a person does.
	UserID string
	// Secret is shown to the creator once and never again.
	Secret string
}

// GoCloak defines the interface for Keycloak operations
// This makes it easier to test without mocks by allowing alternative implementations
type GoCloak interface {
	LoginAdmin(ctx context.Context, username, password, realm string) (*gocloak.JWT, error)
	SetPassword(ctx context.Context, token, userID, realm, password string, temporary bool) error
	CreateUser(ctx context.Context, token, realm string, user gocloak.User) (string, error)
	DeleteUser(ctx context.Context, token, realm, userID string) error
	UpdateUser(ctx context.Context, token, realm string, user gocloak.User) error
	SendVerifyEmail(ctx context.Context, token, userID, realm string) error
	CreateClient(ctx context.Context, token, realm string, newClient gocloak.Client) (string, error)
	DeleteClient(ctx context.Context, token, realm, idOfClient string) error
	GetClientSecret(ctx context.Context, token, realm, idOfClient string) (*gocloak.CredentialRepresentation, error)
	RegenerateClientSecret(ctx context.Context, token, realm, idOfClient string) (*gocloak.CredentialRepresentation, error)
	GetClientServiceAccount(ctx context.Context, token, realm, idOfClient string) (*gocloak.User, error)
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

func (g *RealGoCloak) CreateClient(ctx context.Context, token, realm string, newClient gocloak.Client) (string, error) {
	return g.client.CreateClient(ctx, token, realm, newClient)
}

func (g *RealGoCloak) DeleteClient(ctx context.Context, token, realm, idOfClient string) error {
	return g.client.DeleteClient(ctx, token, realm, idOfClient)
}

func (g *RealGoCloak) GetClientSecret(ctx context.Context, token, realm, idOfClient string) (*gocloak.CredentialRepresentation, error) {
	return g.client.GetClientSecret(ctx, token, realm, idOfClient)
}

func (g *RealGoCloak) RegenerateClientSecret(ctx context.Context, token, realm, idOfClient string) (*gocloak.CredentialRepresentation, error) {
	return g.client.RegenerateClientSecret(ctx, token, realm, idOfClient)
}

func (g *RealGoCloak) GetClientServiceAccount(ctx context.Context, token, realm, idOfClient string) (*gocloak.User, error) {
	return g.client.GetClientServiceAccount(ctx, token, realm, idOfClient)
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

// SendVerifyEmail implements the GoCloak interface
func (g *RealGoCloak) SendVerifyEmail(ctx context.Context, token, userID, realm string) error {
	return g.client.SendVerifyEmail(ctx, token, userID, realm)
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

// SendVerifyEmail asks Keycloak to send the address-verification email to the
// given user. Keycloak owns the mail template and delivery; the realm's SMTP
// settings must be configured for this to have any effect.
func (c *client) SendVerifyEmail(ctx context.Context, userID string) error {
	token, err := c.getToken(ctx)
	if err != nil {
		return err
	}

	return c.gocloak.SendVerifyEmail(ctx, token, userID, c.cfg.Realm)
}

// CreateServiceAccountClient registers a confidential client for a machine.
//
// Only the client credentials grant is enabled on it. A machine has no browser
// to redirect and no user to consent, so the flows that assume one are turned
// off rather than left at their defaults: a client that could also do the
// authorization code flow is a client that could be pointed at a login page.
func (c *client) CreateServiceAccountClient(
	ctx context.Context,
	clientID, description string,
) (ServiceAccountClient, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return ServiceAccountClient{}, err
	}

	keycloakID, err := c.gocloak.CreateClient(ctx, token, c.cfg.Realm, gocloak.Client{
		ClientID:                  &clientID,
		Description:               &description,
		Enabled:                   gocloak.BoolP(true),
		PublicClient:              gocloak.BoolP(false),
		ServiceAccountsEnabled:    gocloak.BoolP(true),
		StandardFlowEnabled:       gocloak.BoolP(false),
		ImplicitFlowEnabled:       gocloak.BoolP(false),
		DirectAccessGrantsEnabled: gocloak.BoolP(false),
	})
	if err != nil {
		return ServiceAccountClient{}, fmt.Errorf("failed to create keycloak client: %w", err)
	}

	serviceAccount, err := c.gocloak.GetClientServiceAccount(ctx, token, c.cfg.Realm, keycloakID)
	if err != nil {
		return ServiceAccountClient{}, fmt.Errorf("failed to read the client's service account user: %w", err)
	}
	if serviceAccount == nil || serviceAccount.ID == nil {
		return ServiceAccountClient{}, fmt.Errorf("keycloak returned no service account user for client %s", clientID)
	}

	secret, err := c.gocloak.GetClientSecret(ctx, token, c.cfg.Realm, keycloakID)
	if err != nil {
		return ServiceAccountClient{}, fmt.Errorf("failed to read the client secret: %w", err)
	}
	if secret == nil || secret.Value == nil {
		return ServiceAccountClient{}, fmt.Errorf("keycloak returned no secret for client %s", clientID)
	}

	return ServiceAccountClient{
		KeycloakID: keycloakID,
		UserID:     *serviceAccount.ID,
		Secret:     *secret.Value,
	}, nil
}

// RotateClientSecret issues a new secret, which invalidates the old one.
func (c *client) RotateClientSecret(ctx context.Context, keycloakID string) (string, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return "", err
	}

	secret, err := c.gocloak.RegenerateClientSecret(ctx, token, c.cfg.Realm, keycloakID)
	if err != nil {
		return "", fmt.Errorf("failed to regenerate the client secret: %w", err)
	}
	if secret == nil || secret.Value == nil {
		return "", fmt.Errorf("keycloak returned no secret for client %s", keycloakID)
	}
	return *secret.Value, nil
}

// DeleteClient removes the client, which is what actually stops the machine
// authenticating. Deleting only hub's record would leave working credentials
// behind.
func (c *client) DeleteClient(ctx context.Context, keycloakID string) error {
	token, err := c.getToken(ctx)
	if err != nil {
		return err
	}
	return c.gocloak.DeleteClient(ctx, token, c.cfg.Realm, keycloakID)
}
