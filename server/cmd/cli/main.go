package main

import (
	"database/sql"

	"github.com/spf13/cobra"

	"github.com/linzhengen/hub/v1/server/config"
	"github.com/linzhengen/hub/v1/server/di"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence"
	"github.com/linzhengen/hub/v1/server/internal/interface/cmd/register"
	"github.com/linzhengen/hub/v1/server/pkg/logger"
)

var rootCmd = &cobra.Command{
	Use:   "cli",
	Short: "cli for hub commands",
}

func main() {
	envCfg := config.New(rootCmd.Context())
	db := initDB(envCfg)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Severef("failed to close db, err: %v", err)
		}
	}(db)
	c := di.NewDI(envCfg, db)
	var commands register.Commands
	if err := c.Invoke(func(cs register.Commands) {
		commands = cs
	}); err != nil {
		logger.Severef("invoke commands err: %v", err)
	}
	rootCmd.AddCommand(commands...)
	if err := rootCmd.Execute(); err != nil {
		logger.Severef("execute commands err: %v", err)
	}
}

func initDB(envCfg config.EnvConfig) *sql.DB {
	db, err := persistence.NewConnection(envCfg)
	if err != nil {
		logger.Severef("failed connect db, err: %v", err)
	}
	return db
}
