package ai

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	domain "github.com/linzhengen/hub/server/internal/domain/ai/agent"
	usecase "github.com/linzhengen/hub/server/internal/usecase/ai"
	pbv1 "github.com/linzhengen/hub/server/pb/ai/agent/v1"
)

func NewAgentHandler(agentUseCase usecase.AgentUseCase) pbv1.AgentServiceServer {
	return &agentHandler{agentUseCase: agentUseCase}
}

type agentHandler struct {
	agentUseCase usecase.AgentUseCase
}

func (h agentHandler) CreateAgent(
	ctx context.Context,
	request *pbv1.CreateAgentRequest,
) (*pbv1.CreateAgentResponse, error) {
	created, credentials, err := h.agentUseCase.Create(ctx, &domain.Agent{
		OrgId:         request.OrgId,
		Name:          request.Name,
		Description:   request.Description,
		ParentAgentId: request.GetParentAgentId(),
	})
	if err != nil {
		return nil, err
	}
	return &pbv1.CreateAgentResponse{
		Agent:       agentDomainToPb(created),
		Credentials: credentialsToPb(credentials),
	}, nil
}

func (h agentHandler) GetAgent(
	ctx context.Context,
	request *pbv1.GetAgentRequest,
) (*pbv1.GetAgentResponse, error) {
	found, err := h.agentUseCase.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &pbv1.GetAgentResponse{Agent: agentDomainToPb(found)}, nil
}

func (h agentHandler) ListAgents(
	ctx context.Context,
	request *pbv1.ListAgentsRequest,
) (*pbv1.ListAgentsResponse, error) {
	items, total, err := h.agentUseCase.List(ctx, domain.ListParams{
		Limit:         request.Limit,
		Offset:        request.Offset,
		OrgId:         request.GetOrgId(),
		ParentAgentId: request.GetParentAgentId(),
	})
	if err != nil {
		return nil, err
	}
	pbItems := make([]*pbv1.Agent, 0, len(items))
	for _, item := range items {
		pbItems = append(pbItems, agentDomainToPb(item))
	}
	return &pbv1.ListAgentsResponse{Agents: pbItems, Total: total}, nil
}

func (h agentHandler) RotateAgentSecret(
	ctx context.Context,
	request *pbv1.RotateAgentSecretRequest,
) (*pbv1.RotateAgentSecretResponse, error) {
	credentials, err := h.agentUseCase.RotateSecret(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &pbv1.RotateAgentSecretResponse{Credentials: credentialsToPb(credentials)}, nil
}

func (h agentHandler) DeleteAgent(
	ctx context.Context,
	request *pbv1.DeleteAgentRequest,
) (*pbv1.DeleteAgentResponse, error) {
	if err := h.agentUseCase.Delete(ctx, request.Id); err != nil {
		return nil, err
	}
	return &pbv1.DeleteAgentResponse{}, nil
}

func agentDomainToPb(a *domain.Agent) *pbv1.Agent {
	return &pbv1.Agent{
		Id:              a.Id,
		OrgId:           a.OrgId,
		UserId:          a.UserId,
		Name:            a.Name,
		Description:     a.Description,
		ClientId:        a.ClientId,
		AuthMethod:      authMethodToPb(a.AuthMethod),
		ParentAgentId:   a.ParentAgentId,
		CreatedByUserId: a.CreatedByUserId,
		SecretRotatedAt: timestamppb.New(a.SecretRotatedAt),
		CreatedAt:       timestamppb.New(a.CreatedAt),
		UpdatedAt:       timestamppb.New(a.UpdatedAt),
	}
}

func authMethodToPb(m domain.AuthMethod) pbv1.AuthMethod {
	switch m {
	case domain.AuthMethodClientSecret:
		return pbv1.AuthMethod_AUTH_METHOD_CLIENT_SECRET
	case domain.AuthMethodPrivateKeyJWT:
		return pbv1.AuthMethod_AUTH_METHOD_PRIVATE_KEY_JWT
	default:
		return pbv1.AuthMethod_AUTH_METHOD_UNSPECIFIED
	}
}

// credentialsToPb carries the secret out to the caller. It is the only place it
// appears in a response, and there is no rpc that reads it back afterwards.
func credentialsToPb(c usecase.Credentials) *pbv1.Credentials {
	return &pbv1.Credentials{ClientId: c.ClientId, Secret: c.Secret}
}
