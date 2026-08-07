package hubcli

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tokenServer stands in for Keycloak's token endpoint, recording the grant it
// was sent so a test can assert which one the CLI chose.
func tokenServer(t *testing.T, status int, body string) (*httptest.Server, *url.Values) {
	t.Helper()

	form := &url.Values{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/protocol/openid-connect/token", r.URL.Path)
		require.NoError(t, r.ParseForm())
		*form = r.PostForm

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server, form
}

func TestRequestTokenChoosesTheGrant(t *testing.T) {
	t.Run("a username selects the password grant", func(t *testing.T) {
		server, form := tokenServer(t, http.StatusOK, `{"access_token":"at","refresh_token":"rt","expires_in":300}`)

		token, err := requestToken(context.Background(),
			OIDC{Issuer: server.URL, ClientID: "hub", ClientSecret: "shh"}, "admin", "pw")
		require.NoError(t, err)

		assert.Equal(t, "password", form.Get("grant_type"))
		assert.Equal(t, "admin", form.Get("username"))
		assert.Equal(t, "pw", form.Get("password"))
		assert.Equal(t, "hub", form.Get("client_id"))
		assert.Equal(t, "shh", form.Get("client_secret"))
		assert.Equal(t, "at", token.AccessToken)
		assert.Equal(t, "rt", token.RefreshToken)
	})

	t.Run("no username selects client credentials", func(t *testing.T) {
		server, form := tokenServer(t, http.StatusOK, `{"access_token":"at"}`)

		_, err := requestToken(context.Background(), OIDC{Issuer: server.URL, ClientID: "hub"}, "", "")
		require.NoError(t, err)

		assert.Equal(t, "client_credentials", form.Get("grant_type"))
		assert.Empty(t, form.Get("client_secret"), "a public client sends no secret")
	})

	t.Run("refreshing uses the refresh grant", func(t *testing.T) {
		server, form := tokenServer(t, http.StatusOK, `{"access_token":"at2","expires_in":300}`)

		_, err := refreshToken(context.Background(), OIDC{Issuer: server.URL, ClientID: "hub"}, "rt")
		require.NoError(t, err)

		assert.Equal(t, "refresh_token", form.Get("grant_type"))
		assert.Equal(t, "rt", form.Get("refresh_token"))
	})
}

// expires_in is a duration, which is meaningless once written to a file, so it
// is resolved against the clock as the response arrives.
func TestRequestTokenResolvesTheExpiry(t *testing.T) {
	server, _ := tokenServer(t, http.StatusOK, `{"access_token":"at","expires_in":300}`)

	before := time.Now()
	token, err := requestToken(context.Background(), OIDC{Issuer: server.URL, ClientID: "hub"}, "", "")
	require.NoError(t, err)

	assert.WithinRange(t, token.ExpiresAt(), before.Add(300*time.Second), time.Now().Add(300*time.Second))
}

func TestRequestTokenRejectsUnusableResponses(t *testing.T) {
	t.Run("missing configuration", func(t *testing.T) {
		_, err := requestToken(context.Background(), OIDC{}, "", "")
		assert.ErrorContains(t, err, "no OIDC issuer configured")

		_, err = requestToken(context.Background(), OIDC{Issuer: "http://x"}, "", "")
		assert.ErrorContains(t, err, "no OIDC client configured")
	})

	t.Run("an error from the provider", func(t *testing.T) {
		server, _ := tokenServer(t, http.StatusUnauthorized, `{"error":"invalid_grant"}`)

		_, err := requestToken(context.Background(), OIDC{Issuer: server.URL, ClientID: "hub"}, "admin", "wrong")

		var apiErr *APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Contains(t, err.Error(), "invalid_grant")
	})

	t.Run("a response with no token", func(t *testing.T) {
		server, _ := tokenServer(t, http.StatusOK, `{"token_type":"Bearer"}`)

		_, err := requestToken(context.Background(), OIDC{Issuer: server.URL, ClientID: "hub"}, "", "")
		assert.ErrorContains(t, err, "no access_token")
	})
}

func TestSaveToken(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv("HUB_CONFIG", path)

	c := &CLI{settings: Settings{Endpoint: "http://hub.test", OIDC: OIDC{Issuer: "http://issuer", ClientID: "hub"}}}
	expiry := time.Now().Add(time.Hour).Truncate(time.Second)

	require.NoError(t, c.saveToken(Token{AccessToken: "at", RefreshToken: "rt", expiresAt: expiry}))

	saved, err := LoadConfig(path)
	require.NoError(t, err)
	profile := saved.Profile("default")
	assert.Equal(t, "at", profile.Token)
	assert.Equal(t, "rt", profile.RefreshToken)
	assert.WithinDuration(t, expiry, profile.ExpiresAt, time.Second)
	assert.Equal(t, "http://hub.test", profile.Endpoint, "the endpoint in use is remembered")
	assert.Equal(t, "hub", profile.OIDC.ClientID)

	t.Run("a refresh that does not rotate keeps the old refresh token", func(t *testing.T) {
		require.NoError(t, c.saveToken(Token{AccessToken: "at2"}))

		saved, err := LoadConfig(path)
		require.NoError(t, err)
		assert.Equal(t, "at2", saved.Profile("default").Token)
		assert.Equal(t, "rt", saved.Profile("default").RefreshToken)
	})
}

func TestEnsureFreshToken(t *testing.T) {
	newCLI := func(t *testing.T, settings Settings) *CLI {
		t.Helper()
		t.Setenv("HUB_CONFIG", filepath.Join(t.TempDir(), "config.yaml"))
		return &CLI{settings: settings, client: NewClient(settings)}
	}

	t.Run("renews a spent token and stores the new one", func(t *testing.T) {
		server, form := tokenServer(t, http.StatusOK, `{"access_token":"fresh","refresh_token":"rt2","expires_in":300}`)
		c := newCLI(t, Settings{
			Token:        "stale",
			RefreshToken: "rt",
			ExpiresAt:    time.Now().Add(-time.Minute),
			OIDC:         OIDC{Issuer: server.URL, ClientID: "hub"},
		})

		require.NoError(t, c.ensureFreshToken(context.Background()))

		assert.Equal(t, "refresh_token", form.Get("grant_type"))
		assert.Equal(t, "fresh", c.settings.Token)
		assert.Equal(t, "rt2", c.settings.RefreshToken)

		path, err := ConfigPath()
		require.NoError(t, err)
		saved, err := LoadConfig(path)
		require.NoError(t, err)
		assert.Equal(t, "fresh", saved.Profile("default").Token, "the next command must not have to refresh again")
	})

	t.Run("leaves a token that is still good", func(t *testing.T) {
		server, form := tokenServer(t, http.StatusOK, `{"access_token":"fresh"}`)
		c := newCLI(t, Settings{
			Token:        "good",
			RefreshToken: "rt",
			ExpiresAt:    time.Now().Add(time.Hour),
			OIDC:         OIDC{Issuer: server.URL, ClientID: "hub"},
		})

		require.NoError(t, c.ensureFreshToken(context.Background()))

		assert.Equal(t, "good", c.settings.Token)
		assert.Empty(t, form.Get("grant_type"), "no request should have been made")
	})

	// A token the caller passed in is theirs. Replacing it would ignore what
	// they asked for, and writing it to their config file would be worse.
	t.Run("never touches a token given on the command line", func(t *testing.T) {
		server, form := tokenServer(t, http.StatusOK, `{"access_token":"fresh"}`)
		c := newCLI(t, Settings{
			Token:        "explicit",
			RefreshToken: "rt",
			ExpiresAt:    time.Now().Add(-time.Hour),
			OIDC:         OIDC{Issuer: server.URL, ClientID: "hub"},
		})
		c.tokenOverridden = true

		require.NoError(t, c.ensureFreshToken(context.Background()))

		assert.Equal(t, "explicit", c.settings.Token)
		assert.Empty(t, form.Get("grant_type"))
	})

	t.Run("without a refresh token there is nothing to do", func(t *testing.T) {
		c := newCLI(t, Settings{Token: "only", ExpiresAt: time.Now().Add(-time.Hour)})

		assert.NoError(t, c.ensureFreshToken(context.Background()))
		assert.Equal(t, "only", c.settings.Token)
	})

	// The provider did not say when the token dies, so the only way to find out
	// is to use it.
	t.Run("an unknown expiry is not treated as expired", func(t *testing.T) {
		server, form := tokenServer(t, http.StatusOK, `{"access_token":"fresh"}`)
		c := newCLI(t, Settings{
			Token:        "unknown-expiry",
			RefreshToken: "rt",
			OIDC:         OIDC{Issuer: server.URL, ClientID: "hub"},
		})

		require.NoError(t, c.ensureFreshToken(context.Background()))

		assert.Equal(t, "unknown-expiry", c.settings.Token)
		assert.Empty(t, form.Get("grant_type"))
	})

	t.Run("a failed refresh says what to do about it", func(t *testing.T) {
		server, _ := tokenServer(t, http.StatusBadRequest, `{"error":"invalid_grant"}`)
		c := newCLI(t, Settings{
			RefreshToken: "expired",
			ExpiresAt:    time.Now().Add(-time.Hour),
			OIDC:         OIDC{Issuer: server.URL, ClientID: "hub"},
		})

		err := c.ensureFreshToken(context.Background())

		assert.ErrorContains(t, err, "hub auth login")
	})
}
