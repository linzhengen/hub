package handler

import (
	"context"
	"database/sql"

	"github.com/spf13/cobra"

	"github.com/linzhengen/hub/v1/server/internal/usecase/develop"
)

func NewMigrateHandler(db *sql.DB, migrateUseCase develop.MigrateUseCase) MigrateHandler {
	return &migrateHandler{db: db, migrateUseCase: migrateUseCase}
}

type MigrateHandler interface {
	Up(ctx context.Context) *cobra.Command
	Down(ctx context.Context) *cobra.Command
}

type migrateHandler struct {
	db             *sql.DB
	migrateUseCase develop.MigrateUseCase
}

func (h *migrateHandler) Up(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate-up",
		Short: "Migrate up",
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.migrateUseCase.Up(ctx)
		},
	}
}

func (h *migrateHandler) Down(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate-down",
		Short: "Migrate down",
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.migrateUseCase.Down(ctx)
		},
	}
}
