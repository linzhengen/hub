package hubcli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigPathHonoursTheOverride(t *testing.T) {
	t.Setenv("HUB_CONFIG", "/tmp/hub.yaml")

	path, err := ConfigPath()
	require.NoError(t, err)
	assert.Equal(t, "/tmp/hub.yaml", path)
}

// The CLI is normally run from flags and environment variables alone, so a
// missing config file must not be an error.
func TestLoadConfigTreatsAMissingFileAsEmpty(t *testing.T) {
	cfg, err := LoadConfig(filepath.Join(t.TempDir(), "absent.yaml"))

	require.NoError(t, err)
	assert.Empty(t, cfg.Profiles)
}

func TestSaveConfigRoundTripsAndKeepsTheFilePrivate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.yaml")
	cfg := Config{}.WithProfile("default", Profile{
		Endpoint: "http://hub.test",
		Token:    "t0ken",
		OIDC:     OIDC{Issuer: "http://keycloak/realms/hub", ClientID: "hub"},
	})

	require.NoError(t, SaveConfig(path, cfg))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), "the file holds tokens and secrets")

	loaded, err := LoadConfig(path)
	require.NoError(t, err)
	assert.Equal(t, cfg, loaded)
	assert.Equal(t, "default", loaded.Current, "the first profile written becomes the current one")
}

func TestProfileFallsBackToTheCurrentThenToDefault(t *testing.T) {
	cfg := Config{
		Current: "staging",
		Profiles: map[string]Profile{
			"default": {Endpoint: "http://default"},
			"staging": {Endpoint: "http://staging"},
		},
	}

	assert.Equal(t, "http://staging", cfg.Profile("").Endpoint)
	assert.Equal(t, "http://default", cfg.Profile("default").Endpoint)
	assert.Equal(t, "http://default", Config{Profiles: cfg.Profiles}.Profile("").Endpoint)
	assert.Empty(t, cfg.Profile("absent").Endpoint)
}

func TestWithProfileDoesNotMutateTheReceiver(t *testing.T) {
	original := Config{Current: "default", Profiles: map[string]Profile{"default": {Endpoint: "http://one"}}}

	updated := original.WithProfile("default", Profile{Endpoint: "http://two"})

	assert.Equal(t, "http://one", original.Profiles["default"].Endpoint)
	assert.Equal(t, "http://two", updated.Profiles["default"].Endpoint)
}

func TestResolveLayersFileThenEnvironmentThenFlags(t *testing.T) {
	profile := Profile{
		Endpoint: "http://from-file",
		Token:    "file-token",
		OIDC:     OIDC{Issuer: "http://file-issuer", ClientID: "file-client"},
	}

	t.Run("file only", func(t *testing.T) {
		settings := Resolve(profile, Settings{})
		assert.Equal(t, "http://from-file", settings.Endpoint)
		assert.Equal(t, "file-token", settings.Token)
	})

	t.Run("environment beats the file", func(t *testing.T) {
		t.Setenv("HUB_ENDPOINT", "http://from-env")
		t.Setenv("HUB_TOKEN", "env-token")

		settings := Resolve(profile, Settings{})
		assert.Equal(t, "http://from-env", settings.Endpoint)
		assert.Equal(t, "env-token", settings.Token)
	})

	t.Run("flags beat the environment", func(t *testing.T) {
		t.Setenv("HUB_ENDPOINT", "http://from-env")
		t.Setenv("HUB_OIDC_CLIENT_ID", "env-client")

		settings := Resolve(profile, Settings{Endpoint: "http://from-flag"})
		assert.Equal(t, "http://from-flag", settings.Endpoint)
		assert.Equal(t, "env-client", settings.OIDC.ClientID, "an unset flag does not clear the environment")
	})

	t.Run("endpoint has a local default", func(t *testing.T) {
		assert.Equal(t, "http://localhost:9090", Resolve(Profile{}, Settings{}).Endpoint)
	})
}
