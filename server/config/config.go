package config

import (
	"context"
	"fmt"
	"time"

	envconfig "github.com/sethvargo/go-envconfig"

	"github.com/linzhengen/hub/v1/server/pkg/logger"
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
	AllowCredentials bool     `env:"CORS_ALLOW_CREDENTIALS,default=true"`
	MaxAge           int      `env:"CORS_MAX_AGE,default=7200"`
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

func New(ctx context.Context) EnvConfig {
	var c EnvConfig
	if err := envconfig.Process(ctx, &c); err != nil {
		logger.Severe(err)
	}
	return c
}
