package register

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/linzhengen/hub/v1/server/internal/interface/cmd/handler"
)

type Commands []*cobra.Command

func New(
	migrateHandler handler.MigrateHandler,
	seedHandler handler.SeedHandler,
	importResourceHandler handler.ImportResourceHandler,
) Commands {
	ctx := context.Background()
	return []*cobra.Command{
		migrateHandler.Up(ctx),
		migrateHandler.Down(ctx),
		seedHandler.Seed(ctx),
		importResourceHandler.Import(ctx),
	}
}
