package audit

import (
	"context"

	auditDomain "github.com/linzhengen/hub/server/internal/domain/audit"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence"
)

type repositoryImpl struct {
	q persistence.Querier
}

func New(q persistence.Querier) auditDomain.Repository {
	return &repositoryImpl{q: q}
}

// Create writes the record through whichever querier the context carries, so a
// record written inside a transaction commits with the change it describes and
// is rolled back with it.
func (r repositoryImpl) Create(ctx context.Context, e *auditDomain.Entry) error {
	arguments := e.Arguments
	if len(arguments) == 0 {
		arguments = []byte(`{}`)
	}
	row, err := persistence.GetQ(ctx, r.q).CreateAuditLog(
		ctx,
		e.ActorUserId,
		string(e.Channel),
		e.SessionId,
		e.Resource,
		e.Action,
		e.TargetId,
		arguments,
		e.Succeeded,
		e.Error,
		e.Client,
		e.ApprovalId,
	)
	if err != nil {
		return err
	}
	e.Id = row.ID
	e.CreatedAt = row.CreatedAt
	return nil
}
