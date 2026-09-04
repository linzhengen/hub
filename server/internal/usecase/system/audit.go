package system

import (
	"context"
	"database/sql"
	"time"

	"github.com/doug-martin/goqu/v9"

	"github.com/linzhengen/hub/server/internal/domain/audit"
	"github.com/linzhengen/hub/server/internal/domain/auth"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence/postgres"
	"github.com/linzhengen/hub/server/internal/usecase/pagination"
	"github.com/linzhengen/hub/server/internal/usecase/scope"

	"github.com/linzhengen/hub/server/pkg/logger"
)

// AuditUseCase reads the audit log. It offers no write: records are written by
// the recording interceptor as changes are made, and an endpoint that could add
// or amend one would undo the point of having them.
type AuditUseCase interface {
	List(ctx context.Context, params *ListAuditLogQueryParams) ([]*audit.Entry, int64, error)
}

func NewAuditUseCase(
	db *sql.DB,
	dialectWrapper persistence.DialectWrapper,
	authSvc auth.Service,
) AuditUseCase {
	return &auditUseCase{db: db, dialectWrapper: dialectWrapper, authSvc: authSvc}
}

type auditUseCase struct {
	db             *sql.DB
	dialectWrapper persistence.DialectWrapper
	authSvc        auth.Service
}

type ListAuditLogQueryParams struct {
	Limit       uint32
	Offset      uint32
	ActorUserId string
	Resource    string
	Action      string
	TargetId    string
	Channel     audit.Channel
	Since       time.Time
	Until       time.Time
}

func (uc auditUseCase) List(
	ctx context.Context,
	params *ListAuditLogQueryParams,
) ([]*audit.Entry, int64, error) {
	// The log is the one place every change in the installation is written
	// down, so an unnarrowed read of it is a way around every other boundary.
	visible, err := scope.VisibleOrgs(ctx, uc.authSvc)
	if err != nil {
		return nil, 0, err
	}
	if visible.Empty() {
		return nil, 0, nil
	}

	b := uc.dialectWrapper.From("audit_logs")
	if !visible.All {
		b = b.Where(goqu.Ex{"org_id": visible.OrgIds})
	}
	if params.ActorUserId != "" {
		b = b.Where(goqu.Ex{"actor_user_id": params.ActorUserId})
	}
	if params.Resource != "" {
		b = b.Where(goqu.Ex{"resource": params.Resource})
	}
	if params.Action != "" {
		b = b.Where(goqu.Ex{"action": params.Action})
	}
	if params.TargetId != "" {
		b = b.Where(goqu.Ex{"target_id": params.TargetId})
	}
	if params.Channel != "" {
		b = b.Where(goqu.Ex{"channel": string(params.Channel)})
	}
	if !params.Since.IsZero() {
		b = b.Where(goqu.C("created_at").Gte(params.Since))
	}
	if !params.Until.IsZero() {
		b = b.Where(goqu.C("created_at").Lte(params.Until))
	}

	cnt, err := postgres.SelectCount(ctx, uc.db, b)
	if err != nil {
		return nil, 0, err
	}

	page := pagination.New(params.Limit, params.Offset)
	// Newest first: a log is read to find out what just happened.
	b = b.Order(goqu.C("created_at").Desc()).Limit(page.Limit()).Offset(page.Offset())

	items, err := uc.list(ctx, b)
	if err != nil {
		return nil, 0, err
	}
	return items, cnt, nil
}

func (uc auditUseCase) list(ctx context.Context, b *goqu.SelectDataset) ([]*audit.Entry, error) {
	query, queryParams, err := b.Select("*").Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}
	rows, err := uc.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Infof("error closing rows: %v", err)
		}
	}()

	var items []*audit.Entry
	for rows.Next() {
		var (
			entry      audit.Entry
			orgId      sql.NullString
			channel    string
			sessionId  sql.NullString
			approvalId sql.NullString
			arguments  []byte
		)
		if err := rows.Scan(
			&entry.Id,
			&entry.ActorUserId,
			&orgId,
			&channel,
			&sessionId,
			&entry.Resource,
			&entry.Action,
			&entry.TargetId,
			&arguments,
			&entry.Succeeded,
			&entry.Error,
			&entry.Client,
			&entry.CreatedAt,
			&approvalId,
		); err != nil {
			return nil, err
		}
		entry.Channel = audit.Channel(channel)
		entry.OrgId = orgId.String
		entry.SessionId = sessionId.String
		entry.ApprovalId = approvalId.String
		entry.Arguments = arguments
		items = append(items, &entry)
	}
	return items, rows.Err()
}
