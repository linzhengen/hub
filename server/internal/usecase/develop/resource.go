package develop

import (
	"context"

	"github.com/linzhengen/hub/v1/server/internal/domain/system"
)

type ResourceUseCase interface {
	ImportResourcesAndPermissions(ctx context.Context) error
}

func NewResourceUseCase(
	resourceSvc system.ResourceService,
) ResourceUseCase {
	return &resourceUseCase{
		resourceSvc: resourceSvc,
	}
}

type resourceUseCase struct {
	resourceSvc system.ResourceService
}

func (a resourceUseCase) ImportResourcesAndPermissions(ctx context.Context) error {
	if err := a.resourceSvc.CreateResourcesAndPermissionsFromAPI(ctx); err != nil {
		return err
	}
	return a.resourceSvc.CreateResourcesAndPermissionsFromMenu(ctx)
}
