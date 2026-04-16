package modules

import (
	"go.uber.org/dig"

	cmdHandler "github.com/linzhengen/hub/v1/server/internal/interface/cmd/handler"
	cmdRegister "github.com/linzhengen/hub/v1/server/internal/interface/cmd/register"
	"github.com/linzhengen/hub/v1/server/internal/usecase/develop"
)

// ProvideCLI registers CLI-related dependencies.
func ProvideCLI(c *dig.Container) {
	// usecase
	must(c.Provide(develop.NewMigrateUseCase))
	must(c.Provide(develop.NewSeedUseCase))
	// interface (CLI)
	must(c.Provide(cmdHandler.NewMigrateHandler))
	must(c.Provide(cmdHandler.NewSeedHandler))
	must(c.Provide(cmdHandler.NewImportResourceHandler))
	must(c.Provide(cmdRegister.New))
}
