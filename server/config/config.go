package config

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	envconfig "github.com/sethvargo/go-envconfig"

	"github.com/linzhengen/hub/server/pkg/logger"
)

type EnvConfig struct {
	AppEnv   string `env:"APP_ENV,required"`
	Port     int    `env:"PORT,default=9090"`
	Database string `env:"DATABASE,default=postgres"`
	Auth
	Grpc
	Log
	PostgreSQL
	CORS
	Migration
	Seed
	RateLimit
	KeyCloak
	Anthropic
}

type Anthropic struct {
	// APIKey is required for the real client, but not when Mock is on: local
	// development would otherwise be blocked on an Anthropic account.
	APIKey string `env:"ANTHROPIC_API_KEY"`
	Model  string `env:"ANTHROPIC_MODEL,default=claude-sonnet-4-6"`
	// Mock swaps the Anthropic client for a scripted one that needs neither a
	// key nor network access. See internal/infrastructure/ai/mock.
	Mock bool `env:"ANTHROPIC_MOCK,default=false"`
	// SessionTokenBudget caps how many tokens one chat session may spend before
	// it stops answering. It bounds both an honest runaway - a conversation that
	// keeps re-reading a large directory - and a deliberate one, and it is per
	// session rather than per user so that hitting it costs the user one
	// conversation rather than the assistant.
	SessionTokenBudget int64 `env:"ANTHROPIC_SESSION_TOKEN_BUDGET,default=400000"`
	// MockDelay is the pause the mock leaves between streamed deltas, so the
	// chat UI can be seen filling in rather than appearing at once.
	MockDelay time.Duration `env:"ANTHROPIC_MOCK_DELAY,default=25ms"`
}

// Validate keeps the API key mandatory for the real client while letting the
// mock backend run without one.
func (a Anthropic) Validate() error {
	if !a.Mock && a.APIKey == "" {
		return errors.New("ANTHROPIC_API_KEY is required unless ANTHROPIC_MOCK=true")
	}
	return nil
}

type Grpc struct {
	Host               string `env:"GRPC_HOST,default=localhost"`
	Port               int    `env:"GRPC_PORT,default=9090"`
	MaxGRPCMessageSize int    `env:"GRPC_MAX_MESSAGE_SIZE,default=1073741824"`
}

func (g Grpc) Addr() string {
	return fmt.Sprintf("%s:%d", g.Host, g.Port)
}

type Log struct {
	Level  int    `env:"LOG_LEVEL,default=5"` //  1: fatal 2: error, 3: warn, 4: info, 5: debug, 6: trace
	Format string `env:"LOG_FORMAT,default=json"`
}

type PostgreSQL struct {
	User         string        `env:"POSTGRES_USER,required"`
	Pass         string        `env:"POSTGRES_PASS,required"`
	Port         int           `env:"POSTGRES_PORT,required"`
	Host         string        `env:"POSTGRES_HOST,required"`
	DBName       string        `env:"POSTGRES_DB_NAME,required"`
	SSLMode      string        `env:"POSTGRES_SSL_MODE,default=disable"`
	MaxLifetime  time.Duration `env:"POSTGRES_MAX_LIFE_TIME,default=7200s"`
	MaxOpenConns int           `env:"POSTGRES_MAX_OPEN_CONNS,default=10"`
	MaxIdleConns int           `env:"POSTGRES_MAX_IDLE_CONNS,default=10"`
}

func (p PostgreSQL) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Pass, p.DBName, p.SSLMode)
}

type CORS struct {
	AllowOrigins     []string `env:"CORS_ALLOW_ORIGINS,default=*"`
	AllowMethods     []string `env:"CORS_ALLOW_METHODS,default=GET,POST,PUT,DELETE,PATCH"`
	AllowHeaders     []string `env:"CORS_ALLOW_HEADERS,default=ACCEPT,Authorization,Content-Type,X-CSRF-Token"`
	AllowCredentials bool     `env:"CORS_ALLOW_CREDENTIALS,default=false"`
	MaxAge           int      `env:"CORS_MAX_AGE,default=7200"`
}

// Validate rejects a CORS configuration that allows any origin ("*") while
// also allowing credentials, since that combination has no safe effect for
// browsers and signals a misconfiguration rather than an intentional policy.
func (c CORS) Validate() error {
	if c.AllowCredentials && slices.Contains(c.AllowOrigins, "*") {
		return errors.New("CORS_ALLOW_CREDENTIALS=true requires an explicit CORS_ALLOW_ORIGINS allowlist, not \"*\"")
	}
	return nil
}

type Migration struct {
	Auto bool `env:"MIGRATION_AUTO,default=false"`
}

type Seed struct {
	Auto bool `env:"SEED_AUTO,default=false"`
}

type RateLimit struct {
	ApiRateLimit uint64 `env:"API_RATE_LIMIT,default=100"`
}

type Auth struct {
	DisableAuth bool `env:"DISABLE_AUTH,default=false"`
	// AuthzPolicyCacheTTL keeps a user's effective RBAC policies in memory for
	// this long, so a burst of requests runs the permission join once instead
	// of once per call.
	//
	// This does not decide how stale a decision can be. The cache is keyed on
	// the database's RBAC revision counter and drops everything the moment an
	// administrator edits the graph, so a grant or revocation lands within a
	// second whatever this is set to. The TTL only bounds how long an entry
	// nobody has invalidated is kept, which is a memory question.
	//
	// Set it to zero to disable caching entirely and re-read the policies on
	// every request.
	AuthzPolicyCacheTTL time.Duration `env:"AUTHZ_POLICY_CACHE_TTL,default=5m"`
}

type KeyCloak struct {
	KeycloakURL  string `env:"KEYCLOAK_URL,required"`
	Realm        string `env:"KEYCLOAK_REALM,default=hub"`
	ClientId     string `env:"KEYCLOAK_CLIENT_ID,required"`
	ClientSecret string `env:"KEYCLOAK_CLIENT_SECRET,required"`
	AdminUser    string `env:"KEYCLOAK_ADMIN_USER,default=admin"`
	AdminPass    string `env:"KEYCLOAK_ADMIN_PASS,default=admin"`
	AdminRealm   string `env:"KEYCLOAK_ADMIN_REALM,default=master"`
}

// Validate runs every embedded section's own check. It has to be spelled out:
// with more than one embedded Validate, a promoted call would be ambiguous, and
// a section whose check nobody calls is a section that silently does nothing.
func (c EnvConfig) Validate() error {
	if err := c.CORS.Validate(); err != nil {
		return err
	}
	return c.Anthropic.Validate()
}

func New(ctx context.Context) EnvConfig {
	var c EnvConfig
	if err := envconfig.Process(ctx, &c); err != nil {
		logger.Severe(err)
	}
	logger.Must(c.Validate())
	return c
}
