package system

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/linzhengen/hub/server/internal/domain/system/organization"
	"github.com/linzhengen/hub/server/internal/usecase/system"
	pbv1 "github.com/linzhengen/hub/server/pb/system/organization/v1"
)

func NewOrganizationHandler(
	organizationUseCase system.OrganizationUseCase,
) pbv1.OrganizationServiceServer {
	return &organizationHandler{organizationUseCase: organizationUseCase}
}

type organizationHandler struct {
	organizationUseCase system.OrganizationUseCase
}

func (h organizationHandler) CreateOrganization(
	ctx context.Context,
	request *pbv1.CreateOrganizationRequest,
) (*pbv1.CreateOrganizationResponse, error) {
	o, err := h.organizationUseCase.Create(ctx, organization.Factory(
		request.Name,
		request.Slug,
		toOrganizationDomainKind(request.Kind),
		request.Description,
	))
	if err != nil {
		return nil, err
	}
	return &pbv1.CreateOrganizationResponse{Organization: organizationDomainToPb(o)}, nil
}

func (h organizationHandler) GetOrganization(
	ctx context.Context,
	request *pbv1.GetOrganizationRequest,
) (*pbv1.GetOrganizationResponse, error) {
	o, err := h.organizationUseCase.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &pbv1.GetOrganizationResponse{Organization: organizationDomainToPb(o)}, nil
}

func (h organizationHandler) ListOrganization(
	ctx context.Context,
	request *pbv1.ListOrganizationRequest,
) (*pbv1.ListOrganizationResponse, error) {
	items, total, err := h.organizationUseCase.List(ctx, organization.ListParams{
		Name:   request.Name,
		Slug:   request.Slug,
		Kind:   toOrganizationDomainKind(request.Kind),
		Limit:  request.Limit,
		Offset: request.Offset,
	})
	if err != nil {
		return nil, err
	}
	return &pbv1.ListOrganizationResponse{
		Organizations: organizationsDomainToPb(items),
		Total:         total,
	}, nil
}

func (h organizationHandler) UpdateOrganization(
	ctx context.Context,
	request *pbv1.UpdateOrganizationRequest,
) (*pbv1.UpdateOrganizationResponse, error) {
	o, err := h.organizationUseCase.Update(ctx, &organization.Organization{
		Id:          request.Id,
		Name:        request.Name,
		Slug:        request.Slug,
		Description: request.Description,
		Status:      toOrganizationDomainStatus(request.Status),
	})
	if err != nil {
		return nil, err
	}
	return &pbv1.UpdateOrganizationResponse{Organization: organizationDomainToPb(o)}, nil
}

func (h organizationHandler) DeleteOrganization(
	ctx context.Context,
	request *pbv1.DeleteOrganizationRequest,
) (*pbv1.DeleteOrganizationResponse, error) {
	if err := h.organizationUseCase.Delete(ctx, request.Id); err != nil {
		return nil, err
	}
	return &pbv1.DeleteOrganizationResponse{}, nil
}

func (h organizationHandler) ListMyOrganizations(
	ctx context.Context,
	_ *pbv1.ListMyOrganizationsRequest,
) (*pbv1.ListMyOrganizationsResponse, error) {
	items, err := h.organizationUseCase.ListMine(ctx)
	if err != nil {
		return nil, err
	}
	return &pbv1.ListMyOrganizationsResponse{Organizations: organizationsDomainToPb(items)}, nil
}

func organizationsDomainToPb(items []*organization.Organization) []*pbv1.Organization {
	out := make([]*pbv1.Organization, 0, len(items))
	for _, item := range items {
		out = append(out, organizationDomainToPb(item))
	}
	return out
}

func organizationDomainToPb(m *organization.Organization) *pbv1.Organization {
	return &pbv1.Organization{
		Id:          m.Id,
		Name:        m.Name,
		Slug:        m.Slug,
		Kind:        toOrganizationPbKind(m.Kind),
		Description: m.Description,
		Status:      toOrganizationPbStatus(m.Status),
		CreatedAt:   timestamppb.New(m.CreatedAt),
		UpdatedAt:   timestamppb.New(m.UpdatedAt),
	}
}

func toOrganizationPbKind(k organization.Kind) pbv1.OrganizationKind {
	switch k {
	case organization.KindPlatform:
		return pbv1.OrganizationKind_ORGANIZATION_KIND_PLATFORM
	case organization.KindBusiness:
		return pbv1.OrganizationKind_ORGANIZATION_KIND_BUSINESS
	case organization.KindPersonal:
		return pbv1.OrganizationKind_ORGANIZATION_KIND_PERSONAL
	default:
		return pbv1.OrganizationKind_ORGANIZATION_KIND_UNSPECIFIED
	}
}

// toOrganizationDomainKind maps back. UNSPECIFIED becomes the empty kind, which
// on a list request means "every kind" and on a create is refused by the use
// case rather than silently defaulted.
func toOrganizationDomainKind(k pbv1.OrganizationKind) organization.Kind {
	switch k {
	case pbv1.OrganizationKind_ORGANIZATION_KIND_PLATFORM:
		return organization.KindPlatform
	case pbv1.OrganizationKind_ORGANIZATION_KIND_BUSINESS:
		return organization.KindBusiness
	case pbv1.OrganizationKind_ORGANIZATION_KIND_PERSONAL:
		return organization.KindPersonal
	default:
		return ""
	}
}

func toOrganizationPbStatus(s organization.Status) pbv1.OrganizationStatus {
	switch s {
	case organization.Active:
		return pbv1.OrganizationStatus_ORGANIZATION_STATUS_ACTIVE
	case organization.InActive:
		return pbv1.OrganizationStatus_ORGANIZATION_STATUS_INACTIVE
	default:
		return pbv1.OrganizationStatus_ORGANIZATION_STATUS_UNSPECIFIED
	}
}

func toOrganizationDomainStatus(s pbv1.OrganizationStatus) organization.Status {
	switch s {
	case pbv1.OrganizationStatus_ORGANIZATION_STATUS_INACTIVE:
		return organization.InActive
	default:
		return organization.Active
	}
}
