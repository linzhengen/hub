package system

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/linzhengen/hub/server/internal/domain/system/serviceaccount"
	"github.com/linzhengen/hub/server/internal/usecase/system"
	pbv1 "github.com/linzhengen/hub/server/pb/system/serviceaccount/v1"
)

func NewServiceAccountHandler(
	serviceAccountUseCase system.ServiceAccountUseCase,
) pbv1.ServiceAccountServiceServer {
	return &serviceAccountHandler{serviceAccountUseCase: serviceAccountUseCase}
}

type serviceAccountHandler struct {
	serviceAccountUseCase system.ServiceAccountUseCase
}

func (h serviceAccountHandler) CreateServiceAccount(
	ctx context.Context,
	request *pbv1.CreateServiceAccountRequest,
) (*pbv1.CreateServiceAccountResponse, error) {
	account, credentials, err := h.serviceAccountUseCase.Create(ctx, request.Name, request.Description)
	if err != nil {
		return nil, err
	}
	return &pbv1.CreateServiceAccountResponse{
		ServiceAccount: serviceAccountDomainToPb(account),
		Credentials:    credentialsToPb(credentials),
	}, nil
}

func (h serviceAccountHandler) GetServiceAccount(
	ctx context.Context,
	request *pbv1.GetServiceAccountRequest,
) (*pbv1.GetServiceAccountResponse, error) {
	account, err := h.serviceAccountUseCase.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &pbv1.GetServiceAccountResponse{ServiceAccount: serviceAccountDomainToPb(account)}, nil
}

func (h serviceAccountHandler) ListServiceAccounts(
	ctx context.Context,
	request *pbv1.ListServiceAccountsRequest,
) (*pbv1.ListServiceAccountsResponse, error) {
	items, total, err := h.serviceAccountUseCase.List(ctx, serviceaccount.ListParams{
		Limit:  request.Limit,
		Offset: request.Offset,
	})
	if err != nil {
		return nil, err
	}
	pbItems := make([]*pbv1.ServiceAccount, 0, len(items))
	for _, item := range items {
		pbItems = append(pbItems, serviceAccountDomainToPb(item))
	}
	return &pbv1.ListServiceAccountsResponse{ServiceAccounts: pbItems, Total: total}, nil
}

func (h serviceAccountHandler) RotateServiceAccountSecret(
	ctx context.Context,
	request *pbv1.RotateServiceAccountSecretRequest,
) (*pbv1.RotateServiceAccountSecretResponse, error) {
	credentials, err := h.serviceAccountUseCase.RotateSecret(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &pbv1.RotateServiceAccountSecretResponse{Credentials: credentialsToPb(credentials)}, nil
}

func (h serviceAccountHandler) DeleteServiceAccount(
	ctx context.Context,
	request *pbv1.DeleteServiceAccountRequest,
) (*pbv1.DeleteServiceAccountResponse, error) {
	if err := h.serviceAccountUseCase.Delete(ctx, request.Id); err != nil {
		return nil, err
	}
	return &pbv1.DeleteServiceAccountResponse{}, nil
}

func serviceAccountDomainToPb(s *serviceaccount.ServiceAccount) *pbv1.ServiceAccount {
	return &pbv1.ServiceAccount{
		Id:              s.Id,
		UserId:          s.UserId,
		Name:            s.Name,
		Description:     s.Description,
		ClientId:        s.ClientId,
		CreatedByUserId: s.CreatedByUserId,
		CreatedAt:       timestamppb.New(s.CreatedAt),
		UpdatedAt:       timestamppb.New(s.UpdatedAt),
	}
}

// credentialsToPb carries the secret out to the caller. It is the only place it
// appears in a response, and there is no rpc that reads it back afterwards.
func credentialsToPb(c system.Credentials) *pbv1.Credentials {
	return &pbv1.Credentials{ClientId: c.ClientId, Secret: c.Secret}
}
