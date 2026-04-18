package modules

import (
	"database/sql"
	"log"

	"github.com/Nerzal/gocloak/v13"
	"go.uber.org/dig"

	"github.com/linzhengen/hub/v1/server/config"
	tokenInfra "github.com/linzhengen/hub/v1/server/internal/infrastructure/oidc/token"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence/postgres"
	transInfra "github.com/linzhengen/hub/v1/server/internal/infrastructure/trans"
	"github.com/linzhengen/hub/v1/server/internal/interface/grpc/register"
)

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// ProvideShared registers shared dependencies used across multiple features.
func ProvideShared(c *dig.Container, envCfg config.EnvConfig, db *sql.DB) {
	// config
	must(c.Provide(func() config.EnvConfig {
		return envCfg
	}))
	must(c.Provide(func() config.KeyCloak {
		return envCfg.KeyCloak
	}))
	must(c.Provide(func() string {
		return envCfg.Database
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
	// provide appropriate Querier
	must(c.Provide(func() persistence.Querier {
		q := persistence.GetQuerier(db)
		return persistence.NewQuerierAdapter(q)
	}))
	// provide appropriate dialect
	must(c.Provide(func() persistence.DialectWrapper {
		return postgres.NewDialect()
	}))
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
