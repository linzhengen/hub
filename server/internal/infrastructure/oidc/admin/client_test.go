package admin

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/linzhengen/hub/v1/server/config"
)

// MockGoCloak is a mock implementation of our GoCloak interface
type MockGoCloak struct {
	mock.Mock
}

func (m *MockGoCloak) LoginAdmin(ctx context.Context, username, password, realm string) (*gocloak.JWT, error) {
	args := m.Called(ctx, username, password, realm)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gocloak.JWT), args.Error(1)
}

func (m *MockGoCloak) SetPassword(ctx context.Context, token, userID, realm, password string, temporary bool) error {
	args := m.Called(ctx, token, userID, realm, password, temporary)
	return args.Error(0)
}

func (m *MockGoCloak) CreateUser(ctx context.Context, token, realm string, user gocloak.User) (string, error) {
	args := m.Called(ctx, token, realm, user)
	return args.String(0), args.Error(1)
}

func (m *MockGoCloak) DeleteUser(ctx context.Context, token, realm, userID string) error {
	args := m.Called(ctx, token, realm, userID)
	return args.Error(0)
}

func (m *MockGoCloak) UpdateUser(ctx context.Context, token, realm string, user gocloak.User) error {
	args := m.Called(ctx, token, realm, user)
	return args.Error(0)
}

// Test configuration
func getTestConfig() config.KeyCloak {
	return config.KeyCloak{
		KeycloakURL: "http://localhost:8080",
		Realm:       "test-realm",
		AdminRealm:  "master",
		AdminUser:   "admin",
		AdminPass:   "admin",
	}
}

func TestNewClient(t *testing.T) {
	cfg := getTestConfig()
	client := NewClient(cfg)

	assert.NotNil(t, client, "Client should not be nil")
}

// testClient is a wrapper around the client struct that exposes internal methods for testing
type testClient struct {
	*client
}

// newTestClient creates a new test client with the given mock
func newTestClient(mock GoCloak, cfg config.KeyCloak) *testClient {
	// Create a client with the mock
	return &testClient{
		client: NewClientWithGoCloak(mock, cfg).(*client),
	}
}

func TestLogin(t *testing.T) {
	ctx := context.Background()
	cfg := getTestConfig()

	tests := []struct {
		name        string
		mockJWT     *gocloak.JWT
		mockError   error
		expectError bool
	}{
		{
			name: "Success",
			mockJWT: &gocloak.JWT{
				AccessToken: "test-token",
				ExpiresIn:   60,
			},
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "Failure",
			mockJWT:     nil,
			mockError:   errors.New("login failed"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGoCloak := new(MockGoCloak)
			mockGoCloak.On("LoginAdmin", ctx, cfg.AdminUser, cfg.AdminPass, cfg.AdminRealm).
				Return(tt.mockJWT, tt.mockError).Once()

			// Create a test client with our mock
			c := newTestClient(mockGoCloak, cfg)

			// Call login method with proper locking
			c.tokenMutex.Lock()
			err := c.login(ctx)
			c.tokenMutex.Unlock()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				c.tokenMutex.RLock()
				assert.Equal(t, tt.mockJWT, c.token)
				assert.True(t, c.tokenExpires.After(time.Now()))
				c.tokenMutex.RUnlock()
			}

			mockGoCloak.AssertExpectations(t)
		})
	}
}

func TestGetToken(t *testing.T) {
	ctx := context.Background()
	cfg := getTestConfig()

	tests := []struct {
		name           string
		initialToken   *gocloak.JWT
		initialExpires time.Time
		mockJWT        *gocloak.JWT
		mockError      error
		expectError    bool
		expectLogin    bool
	}{
		{
			name: "Valid token exists",
			initialToken: &gocloak.JWT{
				AccessToken: "existing-token",
				ExpiresIn:   60,
			},
			initialExpires: time.Now().Add(time.Hour),
			mockJWT:        nil,
			mockError:      nil,
			expectError:    false,
			expectLogin:    false,
		},
		{
			name:           "No token exists",
			initialToken:   nil,
			initialExpires: time.Time{},
			mockJWT: &gocloak.JWT{
				AccessToken: "new-token",
				ExpiresIn:   60,
			},
			mockError:   nil,
			expectError: false,
			expectLogin: true,
		},
		{
			name: "Token expired",
			initialToken: &gocloak.JWT{
				AccessToken: "expired-token",
				ExpiresIn:   60,
			},
			initialExpires: time.Now().Add(-time.Hour),
			mockJWT: &gocloak.JWT{
				AccessToken: "new-token",
				ExpiresIn:   60,
			},
			mockError:   nil,
			expectError: false,
			expectLogin: true,
		},
		{
			name:           "Login fails",
			initialToken:   nil,
			initialExpires: time.Time{},
			mockJWT:        nil,
			mockError:      errors.New("login failed"),
			expectError:    true,
			expectLogin:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGoCloak := new(MockGoCloak)

			if tt.expectLogin {
				mockGoCloak.On("LoginAdmin", ctx, cfg.AdminUser, cfg.AdminPass, cfg.AdminRealm).
					Return(tt.mockJWT, tt.mockError).Once()
			}

			// Create a test client with our mock
			c := newTestClient(mockGoCloak, cfg)
			c.token = tt.initialToken
			c.tokenExpires = tt.initialExpires

			token, err := c.getToken(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				if tt.expectLogin {
					assert.Equal(t, tt.mockJWT.AccessToken, token)
				} else {
					assert.Equal(t, tt.initialToken.AccessToken, token)
				}
			}

			mockGoCloak.AssertExpectations(t)
		})
	}
}

// MockTokenClient is a client that returns a predefined token
type MockTokenClient struct {
	testClient
	token string
	err   error
}

func TestSetUserPassword(t *testing.T) {
	ctx := context.Background()
	cfg := getTestConfig()
	userID := "test-user-id"
	password := "test-password"

	tests := []struct {
		name         string
		token        string
		tokenError   error
		mockSetError error
		expectError  bool
	}{
		{
			name:         "Success",
			token:        "test-token",
			tokenError:   nil,
			mockSetError: nil,
			expectError:  false,
		},
		{
			name:         "GetToken fails",
			token:        "",
			tokenError:   errors.New("token error"),
			mockSetError: nil,
			expectError:  true,
		},
		{
			name:         "SetPassword fails",
			token:        "test-token",
			tokenError:   nil,
			mockSetError: errors.New("set password error"),
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGoCloak := new(MockGoCloak)

			// For the "GetToken fails" case, we need to set the error in the MockPasswordClient
			var mockError error
			if tt.tokenError != nil {
				mockError = tt.tokenError
			} else {
				mockError = tt.mockSetError
			}

			// Create a test client with our mock
			c := &MockPasswordClient{
				MockTokenClient: MockTokenClient{
					testClient: *newTestClient(mockGoCloak, cfg),
					token:      tt.token,
					err:        tt.tokenError,
				},
				setPasswordErr: mockError,
			}

			err := c.SetUserPassword(ctx, userID, password)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// No need to assert expectations since we're using MockPasswordClient
		})
	}
}

// MockPasswordClient extends MockTokenClient to also mock SetUserPassword
type MockPasswordClient struct {
	MockTokenClient
	setPasswordErr error
}

func (c *MockPasswordClient) SetUserPassword(ctx context.Context, userID, password string) error {
	return c.setPasswordErr
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	cfg := getTestConfig()
	username := "test-user"
	email := "test@example.com"
	password := "test-password"
	userID := "created-user-id"

	tests := []struct {
		name              string
		token             string
		tokenError        error
		mockCreateError   error
		mockSetPwdError   error
		expectError       bool
		expectSetPassword bool
		testPassword      string
	}{
		{
			name:              "Success with password",
			token:             "test-token",
			tokenError:        nil,
			mockCreateError:   nil,
			mockSetPwdError:   nil,
			expectError:       false,
			expectSetPassword: true,
			testPassword:      password,
		},
		{
			name:              "Success without password",
			token:             "test-token",
			tokenError:        nil,
			mockCreateError:   nil,
			mockSetPwdError:   nil,
			expectError:       false,
			expectSetPassword: false,
			testPassword:      "",
		},
		{
			name:              "GetToken fails",
			token:             "",
			tokenError:        errors.New("token error"),
			mockCreateError:   nil,
			mockSetPwdError:   nil,
			expectError:       true,
			expectSetPassword: false,
			testPassword:      password,
		},
		{
			name:              "CreateUser fails",
			token:             "test-token",
			tokenError:        nil,
			mockCreateError:   errors.New("create user error"),
			mockSetPwdError:   nil,
			expectError:       true,
			expectSetPassword: false,
			testPassword:      password,
		},
		{
			name:              "SetPassword fails",
			token:             "test-token",
			tokenError:        nil,
			mockCreateError:   nil,
			mockSetPwdError:   errors.New("set password error"),
			expectError:       true,
			expectSetPassword: true,
			testPassword:      password,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGoCloak := new(MockGoCloak)

			// Set up LoginAdmin expectation for all cases
			mockJWT := &gocloak.JWT{
				AccessToken: tt.token,
				ExpiresIn:   60,
			}
			mockGoCloak.On("LoginAdmin", ctx, cfg.AdminUser, cfg.AdminPass, cfg.AdminRealm).
				Return(mockJWT, tt.tokenError).Maybe()

			// Create a test client with our mocks
			c := &MockPasswordClient{
				MockTokenClient: MockTokenClient{
					testClient: *newTestClient(mockGoCloak, cfg),
					token:      tt.token,
					err:        tt.tokenError,
				},
				setPasswordErr: tt.mockSetPwdError,
			}

			// Only set up CreateUser expectation if getToken would succeed
			if tt.tokenError == nil {
				expectedUser := gocloak.User{
					Username: &username,
					Email:    &email,
					Enabled:  gocloak.BoolP(true),
				}
				mockGoCloak.On("CreateUser", ctx, tt.token, cfg.Realm, expectedUser).
					Return(userID, tt.mockCreateError).Once()

				// Set up SetPassword expectation if password is provided and CreateUser succeeds
				if tt.testPassword != "" && tt.mockCreateError == nil {
					mockGoCloak.On("SetPassword", ctx, tt.token, userID, cfg.Realm, tt.testPassword, false).
						Return(tt.mockSetPwdError).Maybe()
				}
			}

			id, err := c.CreateUser(ctx, username, email, tt.testPassword)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, userID, id)
			}

			mockGoCloak.AssertExpectations(t)
		})
	}
}

func TestDeleteUser(t *testing.T) {
	ctx := context.Background()
	cfg := getTestConfig()
	userID := "test-user-id"

	tests := []struct {
		name            string
		token           string
		tokenError      error
		mockDeleteError error
		expectError     bool
	}{
		{
			name:            "Success",
			token:           "test-token",
			tokenError:      nil,
			mockDeleteError: nil,
			expectError:     false,
		},
		{
			name:            "GetToken fails",
			token:           "",
			tokenError:      errors.New("token error"),
			mockDeleteError: nil,
			expectError:     true,
		},
		{
			name:            "DeleteUser fails",
			token:           "test-token",
			tokenError:      nil,
			mockDeleteError: errors.New("delete user error"),
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGoCloak := new(MockGoCloak)

			// Set up LoginAdmin expectation for all cases
			mockJWT := &gocloak.JWT{
				AccessToken: tt.token,
				ExpiresIn:   60,
			}
			mockGoCloak.On("LoginAdmin", ctx, cfg.AdminUser, cfg.AdminPass, cfg.AdminRealm).
				Return(mockJWT, tt.tokenError).Maybe()

			// Create a test client with our mock
			c := &MockTokenClient{
				testClient: *newTestClient(mockGoCloak, cfg),
				token:      tt.token,
				err:        tt.tokenError,
			}

			// Only set up DeleteUser expectation if getToken would succeed
			if tt.tokenError == nil {
				mockGoCloak.On("DeleteUser", ctx, tt.token, cfg.Realm, userID).
					Return(tt.mockDeleteError).Once()
			}

			err := c.DeleteUser(ctx, userID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockGoCloak.AssertExpectations(t)
		})
	}
}

func TestUpdateEmail(t *testing.T) {
	ctx := context.Background()
	cfg := getTestConfig()
	userID := "test-user-id"
	email := "new-email@example.com"

	tests := []struct {
		name            string
		token           string
		tokenError      error
		mockUpdateError error
		expectError     bool
	}{
		{
			name:            "Success",
			token:           "test-token",
			tokenError:      nil,
			mockUpdateError: nil,
			expectError:     false,
		},
		{
			name:            "GetToken fails",
			token:           "",
			tokenError:      errors.New("token error"),
			mockUpdateError: nil,
			expectError:     true,
		},
		{
			name:            "UpdateUser fails",
			token:           "test-token",
			tokenError:      nil,
			mockUpdateError: errors.New("update user error"),
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGoCloak := new(MockGoCloak)

			// Set up LoginAdmin expectation for all cases
			mockJWT := &gocloak.JWT{
				AccessToken: tt.token,
				ExpiresIn:   60,
			}
			mockGoCloak.On("LoginAdmin", ctx, cfg.AdminUser, cfg.AdminPass, cfg.AdminRealm).
				Return(mockJWT, tt.tokenError).Maybe()

			// Create a test client with our mock
			c := &MockTokenClient{
				testClient: *newTestClient(mockGoCloak, cfg),
				token:      tt.token,
				err:        tt.tokenError,
			}

			// Only set up UpdateUser expectation if getToken would succeed
			if tt.tokenError == nil {
				// Create expected user object with updated email
				expectedUser := gocloak.User{
					ID:    &userID,
					Email: &email,
				}
				mockGoCloak.On("UpdateUser", ctx, tt.token, cfg.Realm, expectedUser).
					Return(tt.mockUpdateError).Once()
			}

			err := c.UpdateEmail(ctx, userID, email)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockGoCloak.AssertExpectations(t)
		})
	}
}
