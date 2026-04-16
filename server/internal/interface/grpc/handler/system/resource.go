package system

import (
	"context"
	"time"

	"github.com/linzhengen/hub/v1/server/internal/domain/system/resource"
	"github.com/linzhengen/hub/v1/server/internal/usecase/system"
	pbv1 "github.com/linzhengen/hub/v1/server/pb/system/resource/v1"
	"github.com/linzhengen/hub/v1/server/pkg/uuid"
)

func NewResourceHandler(
	resourceUseCase system.ResourceUseCase,
) pbv1.ResourceServiceServer {
	return &resourceHandler{
		resourceUseCase: resourceUseCase,
	}
}

type resourceHandler struct {
	resourceUseCase system.ResourceUseCase
}

func (h resourceHandler) CreateResource(ctx context.Context, request *pbv1.CreateResourceRequest) (*pbv1.CreateResourceResponse, error) {
	var identifier resource.Identifier
	if request.Identifier != nil {
		identifier = resource.Identifier{
			Api:      request.Identifier.Api,
			Category: request.Identifier.Category,
		}
	}
	f := resource.Factory(
		request.Name,
		request.ParentId,
		identifier,
		resource.ResourceType(request.Type),
		request.Path,
		request.GetComponent(),
		request.DisplayOrder,
		request.GetDescription(),
		request.Metadata,
	)
	if request.Status != pbv1.Status_STATUS_UNSPECIFIED {
		status, err := toResourceDomainStatus(request.Status)
		if err != nil {
			return nil, err
		}
		f.Status = status
	}
	r, err := h.resourceUseCase.Create(ctx, f)
	if err != nil {
		return nil, err
	}
	return &pbv1.CreateResourceResponse{Resource: resourceDomainToPb(r)}, nil
}

func (h resourceHandler) DeleteResource(ctx context.Context, request *pbv1.DeleteResourceRequest) (*pbv1.DeleteResourceResponse, error) {
	if err := h.resourceUseCase.Delete(ctx, request.Id); err != nil {
		return nil, err
	}
	return &pbv1.DeleteResourceResponse{}, nil
}

func (h resourceHandler) GetResource(ctx context.Context, request *pbv1.GetResourceRequest) (*pbv1.GetResourceResponse, error) {
	r, err := h.resourceUseCase.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &pbv1.GetResourceResponse{Resource: resourceDomainToPb(r)}, nil
}

func (h resourceHandler) ListResource(ctx context.Context, request *pbv1.ListResourceRequest) (*pbv1.ListResourceResponse, error) {
	items, total, err := h.resourceUseCase.List(ctx, &system.ListResourceQueryParams{
		Limit:        request.Limit,
		Offset:       request.Offset,
		ResourceIds:  request.ResourceIds,
		ResourceName: request.ResourceName,
		ResourceType: resource.ResourceType(request.ResourceType),
	})
	if err != nil {
		return nil, err
	}
	var pbItems []*pbv1.Resource
	for _, v := range items {
		pbItems = append(pbItems, resourceDomainToPb(v))
	}
	return &pbv1.ListResourceResponse{
		Resources: pbItems,
		Total:     total,
	}, nil
}

func (h resourceHandler) UpdateResource(ctx context.Context, request *pbv1.UpdateResourceRequest) (*pbv1.UpdateResourceResponse, error) {
	status, err := toResourceDomainStatus(request.Status)
	if err != nil {
		return nil, err
	}
	var identifier resource.Identifier
	if request.Identifier != nil {
		identifier = resource.Identifier{
			Api:      request.Identifier.Api,
			Category: request.Identifier.Category,
		}
	}
	p, err := h.resourceUseCase.Update(ctx, &resource.Resource{
		Id:           request.Id,
		ParentId:     request.ParentId,
		Name:         request.Name,
		Identifier:   identifier,
		Type:         resource.ResourceType(request.Type),
		Path:         request.Path,
		Component:    request.GetComponent(),
		DisplayOrder: request.DisplayOrder,
		Description:  request.GetDescription(),
		Metadata:     request.Metadata,
		Status:       status,
		UpdatedAt:    time.Now(),
	})
	if err != nil {
		return nil, err
	}
	return &pbv1.UpdateResourceResponse{Resource: resourceDomainToPb(p)}, nil
}

func (h resourceHandler) ListMenuResource(ctx context.Context, request *pbv1.ListMenuResourceRequest) (*pbv1.ListMenuResourceResponse, error) {
	items, total, err := h.resourceUseCase.List(ctx, &system.ListResourceQueryParams{
		Limit:        request.Limit,
		Offset:       request.Offset,
		ResourceIds:  request.ResourceIds,
		ResourceName: request.ResourceName,
		ResourceType: resource.ResourceTypeMenu,
	})
	if err != nil {
		return nil, err
	}
	var pbItems []*pbv1.Resource
	for _, v := range items {
		pbItems = append(pbItems, resourceDomainToPb(v))
	}
	return &pbv1.ListMenuResourceResponse{
		Resources: pbItems,
		Total:     total,
	}, nil
}

func (h resourceHandler) CreateMenuResource(ctx context.Context, request *pbv1.CreateMenuResourceRequest) (*pbv1.CreateMenuResourceResponse, error) {
	status, err := toResourceDomainStatus(request.Status)
	if err != nil {
		return nil, err
	}
	f := &resource.Resource{
		Id:           uuid.MustUUID().String(),
		ParentId:     request.ParentId,
		Name:         request.Name,
		Status:       status,
		Path:         request.Path,
		Component:    request.GetComponent(),
		DisplayOrder: request.DisplayOrder,
		Description:  request.GetDescription(),
		Metadata:     request.GetMetadata(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	r, err := h.resourceUseCase.CreateMenu(ctx, f)
	if err != nil {
		return nil, err
	}
	return &pbv1.CreateMenuResourceResponse{Resource: resourceDomainToPb(r)}, nil
}

func (h resourceHandler) UpdateMenuResource(ctx context.Context, request *pbv1.UpdateMenuResourceRequest) (*pbv1.UpdateMenuResourceResponse, error) {
	status, err := toResourceDomainStatus(request.Status)
	if err != nil {
		return nil, err
	}
	f := &resource.Resource{
		Id:           request.Id,
		ParentId:     request.ParentId,
		Name:         request.Name,
		Status:       status,
		Path:         request.Path,
		Component:    request.GetComponent(),
		DisplayOrder: request.DisplayOrder,
		Description:  request.GetDescription(),
		Metadata:     request.GetMetadata(),
		UpdatedAt:    time.Now(),
	}
	r, err := h.resourceUseCase.UpdateMenu(ctx, f)
	if err != nil {
		return nil, err
	}
	return &pbv1.UpdateMenuResourceResponse{Resource: resourceDomainToPb(r)}, nil
}
