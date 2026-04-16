package di

import (
	"database/sql"

	"go.uber.org/dig"

	"github.com/linzhengen/hub/v1/server/config"
	"github.com/linzhengen/hub/v1/server/di/modules"
)

func NewDI(envCfg config.EnvConfig, db *sql.DB) *dig.Container {
	c := dig.New()

	// Register shared dependencies (config, db, transactions, etc.)
	modules.ProvideShared(c, envCfg, db)

	// Register feature-specific dependencies
	modules.ProvideAuth(c)
	modules.ProvideUser(c)
	modules.ProvideSystem(c)
	modules.ProvideGateway(c)
	modules.ProvideCLI(c)

	return c
}
