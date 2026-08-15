package access

import (
	"context"
	"database/sql"
	"errors"
	"time"

	accessDomain "github.com/linzhengen/hub/server/internal/domain/access"
	"github.com/linzhengen/hub/server/internal/infrastructure/persistence"
)

func New(q persistence.Querier) accessDomain.Repository {
	return &repositoryImpl{q: q}
}

type repositoryImpl struct {
	q persistence.Querier
}

func (r repositoryImpl) Create(ctx context.Context, req *accessDomain.Request) error {
	return persistence.GetQ(ctx, r.q).CreateAccessRequest(
		ctx,
		req.Id,
		req.RequesterUserId,
		req.SubjectUserId,
		req.GroupId,
		req.Reason,
		nullTime(req.RequestedUntil),
		string(req.Status),
		string(req.Origin),
		req.SessionId,
	)
}

func (r repositoryImpl) FindOne(ctx context.Context, id string) (*accessDomain.Request, error) {
	row, err := persistence.GetQ(ctx, r.q).SelectAccessRequest(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert(row), nil
}

func (r repositoryImpl) List(
	ctx context.Context,
	params accessDomain.ListParams,
) ([]*accessDomain.Request, int64, error) {
	q := persistence.GetQ(ctx, r.q)

	total, err := q.CountAccessRequests(
		ctx,
		params.RequesterUserId,
		params.SubjectUserId,
		params.GroupId,
		string(params.Status),
	)
	if err != nil {
		return nil, 0, err
	}

	rows, err := q.ListAccessRequests(
		ctx,
		params.RequesterUserId,
		params.SubjectUserId,
		params.GroupId,
		string(params.Status),
		int32(params.Offset), //nolint:gosec // bounded by the request's limit rule
		int32(params.Limit),  //nolint:gosec // bounded by the request's limit rule
	)
	if err != nil {
		return nil, 0, err
	}

	requests := make([]*accessDomain.Request, 0, len(rows))
	for _, row := range rows {
		requests = append(requests, convert(row))
	}
	return requests, total, nil
}

func (r repositoryImpl) Decide(
	ctx context.Context,
	id string,
	d accessDomain.Decision,
) (*accessDomain.Request, bool, error) {
	row, err := persistence.GetQ(ctx, r.q).DecideAccessRequest(
		ctx,
		string(d.Status),
		d.DecidedByUserId,
		sql.NullTime{Time: d.DecidedAt, Valid: true},
		d.Comment,
		id,
	)
	// No row matched, so the request was not pending. That is the claim being
	// lost rather than a failure, and the use case turns it into the message a
	// caller should see.
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return convert(row), true, nil
}

func convert(m *persistence.AccessRequestModel) *accessDomain.Request {
	r := &accessDomain.Request{
		Id:              m.ID,
		RequesterUserId: m.RequesterUserID,
		SubjectUserId:   m.SubjectUserID,
		GroupId:         m.GroupID,
		Reason:          m.Reason,
		RequestedUntil:  fromNullTime(m.RequestedUntil),
		Status:          accessDomain.Status(m.Status),
		Origin:          accessDomain.Origin(m.Origin),
		SessionId:       m.SessionID,
		DecidedByUserId: m.DecidedByUserID,
		DecidedAt:       fromNullTime(m.DecidedAt),
		DecisionComment: m.DecisionComment,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
	return r
}

func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func fromNullTime(v sql.NullTime) *time.Time {
	if !v.Valid {
		return nil
	}
	t := v.Time
	return &t
}
