package agent

import (
	"context"

	domain "github.com/linzhengen/hub/server/internal/domain/ai/agent"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence"
)

func New(q persistence.Querier) domain.Repository {
	return &repositoryImpl{q: q}
}

type repositoryImpl struct {
	q persistence.Querier
}

func (r repositoryImpl) Create(ctx context.Context, a *domain.Agent) error {
	return persistence.GetQ(ctx, r.q).CreateAgent(
		ctx,
		a.Id,
		a.OrgId,
		a.UserId,
		a.Name,
		a.Description,
		a.ClientId,
		a.KeycloakId,
		string(a.AuthMethod),
		a.ParentAgentId,
		a.CreatedByUserId,
	)
}

func (r repositoryImpl) FindOne(ctx context.Context, id string) (*domain.Agent, error) {
	row, err := persistence.GetQ(ctx, r.q).SelectAgent(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert(row), nil
}

func (r repositoryImpl) List(
	ctx context.Context,
	params domain.ListParams,
) ([]*domain.Agent, int64, error) {
	q := persistence.GetQ(ctx, r.q)

	total, err := q.CountAgents(ctx, params.OrgId, params.ParentAgentId, params.OrgIds)
	if err != nil {
		return nil, 0, err
	}

	rows, err := q.ListAgents(
		ctx,
		params.OrgId,
		params.ParentAgentId,
		params.OrgIds,
		int32(params.Offset), //nolint:gosec // bounded by the pagination rule
		int32(params.Limit),  //nolint:gosec // bounded by the pagination rule
	)
	if err != nil {
		return nil, 0, err
	}

	agents := make([]*domain.Agent, 0, len(rows))
	for _, row := range rows {
		agents = append(agents, convert(row))
	}
	return agents, total, nil
}

func (r repositoryImpl) CountChildren(ctx context.Context, id string) (int64, error) {
	return persistence.GetQ(ctx, r.q).CountAgentChildren(ctx, id)
}

func (r repositoryImpl) RecordSecretRotation(ctx context.Context, id string) error {
	return persistence.GetQ(ctx, r.q).RecordAgentSecretRotation(ctx, id)
}

func (r repositoryImpl) Delete(ctx context.Context, id string) error {
	return persistence.GetQ(ctx, r.q).DeleteAgent(ctx, id)
}

func convert(m *persistence.AgentModel) *domain.Agent {
	return &domain.Agent{
		Id:              m.ID,
		OrgId:           m.OrgID,
		UserId:          m.UserID,
		Name:            m.Name,
		Description:     m.Description,
		ClientId:        m.ClientID,
		KeycloakId:      m.KeycloakID,
		AuthMethod:      domain.AuthMethod(m.AuthMethod),
		ParentAgentId:   m.ParentAgentID,
		CreatedByUserId: m.CreatedByUserID,
		SecretRotatedAt: m.SecretRotatedAt,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}
