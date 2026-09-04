package access

import (
	"context"
	"time"
)

// ListParams narrows a listing. An empty field is no filter.
type ListParams struct {
	Limit           uint32
	Offset          uint32
	RequesterUserId string
	SubjectUserId   string
	GroupId         string
	Status          Status
	// OrgIds narrows to the requests whose group belongs to one of these
	// organizations. nil is every organization, which is what the caller who
	// may see every organization asks for; an empty slice would be nothing, and
	// callers short-circuit rather than send one.
	OrgIds []string
}

// Decision is what a decider settled a request with.
type Decision struct {
	Status          Status
	DecidedByUserId string
	DecidedAt       time.Time
	Comment         string
}

type Repository interface {
	Create(ctx context.Context, r *Request) error
	FindOne(ctx context.Context, id string) (*Request, error)
	List(ctx context.Context, params ListParams) ([]*Request, int64, error)
	// Decide settles a request, and only if it is still pending.
	//
	// The condition is in the write rather than in a read before it, so two
	// decisions racing cannot both take effect: the second matches no row and
	// gets ok=false. One approval is one grant, however many times the button
	// is pressed.
	Decide(ctx context.Context, id string, d Decision) (r *Request, ok bool, err error)
}
