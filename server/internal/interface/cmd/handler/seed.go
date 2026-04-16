package handler

import (
	"context"
	"database/sql"

	"github.com/spf13/cobra"

	"github.com/linzhengen/hub/v1/server/internal/usecase/develop"
)

func NewSeedHandler(db *sql.DB, seedUseCase develop.SeedUseCase) SeedHandler {
	return &seedHandler{db: db, seedUseCase: seedUseCase}
}

type SeedHandler interface {
	Seed(ctx context.Context) *cobra.Command
}

type seedHandler struct {
	db          *sql.DB
	seedUseCase develop.SeedUseCase
}

func (h *seedHandler) Seed(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "seed",
		Short: "Execute seed",
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.seedUseCase.Seed(ctx)
		},
	}
}
