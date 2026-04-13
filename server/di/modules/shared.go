package modules

import (
	"database/sql"
	"log"

	"github.com/Nerzal/gocloak/v13"
	"go.uber.org/dig"

	"github.com/linzhengen/hub/server/config"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence/mysql"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence/mysql/sqlc"
	transInfra "github.com/linzhengen/hub/server/internal/infrastructure/trans"
	"github.com/linzhengen/hub/server/internal/interface/grpc/register"
	tokenInfra "github.com/linzhengen/hub/server/internal/infrastructure/oidc/token"
)

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// ProvideShared registers shared dependencies used across multiple features.
func ProvideShared(c *dig.Container, envCfg config.EnvConfig, db *sql.DB) {
	// config
	must(c.Provide(func() config.MySQL {
		return envCfg.MySQL
	}))
	must(c.Provide(func() config.EnvConfig {
		return envCfg
	}))
	must(c.Provide(func() config.KeyCloak {
		return envCfg.KeyCloak
	}))
	// grpc server options
	must(c.Provide(func() *register.Opts {
		return &register.Opts{
			APIRateLimit:       envCfg.ApiRateLimit,
			MaxGRPCMessageSize: envCfg.MaxGRPCMessageSize,
			DisableAuth:        envCfg.DisableAuth,
		}
	}))
	// db
	must(c.Provide(func() *sql.DB {
		return db
	}))
	must(c.Provide(func() sqlc.DBTX {
		return db
	}))
	must(c.Provide(mysql.NewDialect))
	must(c.Provide(sqlc.New))
	// transactions
	must(c.Provide(transInfra.New))
	// token infra (used by auth)
	must(c.Provide(func() *tokenInfra.KeyCloak {
		return &tokenInfra.KeyCloak{
			ClientId:    envCfg.ClientId,
			Realm:       envCfg.Realm,
			Client:      gocloak.NewClient(envCfg.KeycloakURL),
			DisableAuth: envCfg.DisableAuth,
		}
	}))
	must(c.Provide(tokenInfra.New))
	// gRPC register
	must(c.Provide(register.New))
}
