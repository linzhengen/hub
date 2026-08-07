package hubcli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DefaultConfigDir is the directory holding the CLI's configuration, relative
// to the user's home directory.
const DefaultConfigDir = ".hub"

// Profile is one named set of connection settings.
type Profile struct {
	// Endpoint is the base URL of the hub API, e.g. "http://localhost:9090".
	Endpoint string `yaml:"endpoint,omitempty"`
	// Token is a bearer token. `hub auth login` writes one here.
	Token string `yaml:"token,omitempty"`
	// OIDC describes where `hub auth login` gets a token from.
	OIDC OIDC `yaml:"oidc,omitempty"`
}

// OIDC holds the OpenID Connect client settings used to obtain a token.
type OIDC struct {
	// Issuer is the realm URL, e.g. "http://localhost:8080/realms/hub".
	Issuer string `yaml:"issuer,omitempty"`
	// ClientID is the OIDC client the CLI authenticates as.
	ClientID string `yaml:"client_id,omitempty"`
	// ClientSecret is the client secret, required for confidential clients.
	ClientSecret string `yaml:"client_secret,omitempty"`
}

// Config is the on-disk configuration file.
type Config struct {
	// Current is the profile used when --profile is not given.
	Current string `yaml:"current,omitempty"`
	// Profiles holds every configured environment by name.
	Profiles map[string]Profile `yaml:"profiles,omitempty"`
}

// ConfigPath returns the location of the configuration file, honouring
// HUB_CONFIG so a caller can point the CLI at its own file.
func ConfigPath() (string, error) {
	if override := os.Getenv("HUB_CONFIG"); override != "" {
		return override, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot locate the home directory: %w", err)
	}
	return filepath.Join(home, DefaultConfigDir, "config.yaml"), nil
}

// LoadConfig reads the configuration file. A missing file is not an error: the
// CLI is fully usable from flags and environment variables alone, which is how
// it is normally run in CI and by agents.
func LoadConfig(path string) (Config, error) {
	raw, err := os.ReadFile(path) //nolint:gosec // the path is the user's own config
	if errors.Is(err, os.ErrNotExist) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return cfg, nil
}

// SaveConfig writes the configuration file, creating its directory. The file is
// owner-only because it stores bearer tokens and client secrets.
func SaveConfig(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Dir(path), err)
	}
	raw, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// Profile returns the named profile, or the current one when name is empty.
func (c Config) Profile(name string) Profile {
	if name == "" {
		name = c.Current
	}
	if name == "" {
		name = "default"
	}
	return c.Profiles[name]
}

// WithProfile returns a copy of c with the named profile replaced. It is a
// value method so callers cannot accidentally mutate a shared config.
func (c Config) WithProfile(name string, profile Profile) Config {
	profiles := make(map[string]Profile, len(c.Profiles)+1)
	for k, v := range c.Profiles {
		profiles[k] = v
	}
	profiles[name] = profile

	current := c.Current
	if current == "" {
		current = name
	}
	return Config{Current: current, Profiles: profiles}
}

// Settings is the connection configuration actually used for a command, after
// flags, environment variables and the config file have been combined.
type Settings struct {
	Endpoint string
	Token    string
	OIDC     OIDC
}

// Resolve layers the sources in order of increasing precedence: the config
// file, then the environment, then explicit flags. Anything an agent passes on
// the command line therefore wins over whatever happens to be on the machine.
func Resolve(profile Profile, flags Settings) Settings {
	// Settings mirrors Profile field for field, so the conversion is the whole
	// of the "config file" layer. A field added to one and not the other stops
	// compiling here rather than being silently dropped.
	s := Settings(profile)

	override(&s.Endpoint, os.Getenv("HUB_ENDPOINT"), flags.Endpoint)
	override(&s.Token, os.Getenv("HUB_TOKEN"), flags.Token)
	override(&s.OIDC.Issuer, os.Getenv("HUB_OIDC_ISSUER"), flags.OIDC.Issuer)
	override(&s.OIDC.ClientID, os.Getenv("HUB_OIDC_CLIENT_ID"), flags.OIDC.ClientID)
	override(&s.OIDC.ClientSecret, os.Getenv("HUB_OIDC_CLIENT_SECRET"), flags.OIDC.ClientSecret)

	if s.Endpoint == "" {
		s.Endpoint = "http://localhost:9090"
	}
	return s
}

func override(target *string, values ...string) {
	for _, v := range values {
		if v != "" {
			*target = v
		}
	}
}
