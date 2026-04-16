package handler

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/linzhengen/hub/v1/server/internal/usecase/develop"
)

func NewImportResourceHandler(apiResourceUseCase develop.ResourceUseCase) ImportResourceHandler {
	return &resourceHandler{apiResourceUseCase: apiResourceUseCase}
}

type ImportResourceHandler interface {
	Import(ctx context.Context) *cobra.Command
}

type resourceHandler struct {
	apiResourceUseCase develop.ResourceUseCase
}

func (h *resourceHandler) Import(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resource-import",
		Short: "Import Apis and Menus to permissions and resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := h.apiResourceUseCase.ImportResourcesAndPermissions(ctx)
			if err != nil {
				return err
			}

			fmt.Printf("import successfully")
			return nil
		},
	}
	return cmd
}
