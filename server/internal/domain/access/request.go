// Package access holds the request-and-approval side of authorization: asking
// to be put in a group, and somebody agreeing.
//
// The graph itself lives in the system and user domains. This package only ever
// asks them to change it, and only once a person has said so.
package access

import (
	"time"

	"github.com/linzhengen/hub/server/pkg/uuid"
)

// Status is where a request has got to. A request is decided once; there is no
// route back to Pending.
type Status string

const (
	StatusPending   Status = "Pending"
	StatusApproved  Status = "Approved"
	StatusRejected  Status = "Rejected"
	StatusCancelled Status = "Cancelled"
)

// Origin is the surface a request came in by.
//
// It is recorded because OriginAIChat changes how a request should be read. The
// assistant cannot exceed anyone's permissions and cannot approve anything, but
// it composes requests out of text other people wrote, so an approver needs to
// know that a request came from a conversation rather than from a colleague
// typing it.
type Origin string

const (
	OriginConsole Origin = "Console"
	OriginCLI     Origin = "CLI"
	OriginAIChat  Origin = "AIChat"
)

// Request is one person asking for another to be put in a group.
//
// Requester and Subject are separate because the interesting requests are made
// on somebody else's behalf: a manager for a report, the assistant for whoever
// it is answering.
type Request struct {
	Id              string
	RequesterUserId string
	SubjectUserId   string
	GroupId         string
	Reason          string
	// RequestedUntil is the term asked for, nil for "permanently". On approval
	// it becomes the membership's expiry, so a request for a week grants a
	// week.
	RequestedUntil *time.Time
	Status         Status
	Origin         Origin
	// SessionId is the chat session the assistant raised this in. Empty for
	// every other origin.
	SessionId       string
	DecidedByUserId string
	DecidedAt       *time.Time
	DecisionComment string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Factory builds a pending request. Everything about the decision is left
// empty: a request that arrives already decided is not a request.
func Factory(
	requesterUserId string,
	subjectUserId string,
	groupId string,
	reason string,
	requestedUntil *time.Time,
	origin Origin,
	sessionId string,
) *Request {
	now := time.Now()
	return &Request{
		Id:              uuid.MustUUID().String(),
		RequesterUserId: requesterUserId,
		SubjectUserId:   subjectUserId,
		GroupId:         groupId,
		Reason:          reason,
		RequestedUntil:  requestedUntil,
		Status:          StatusPending,
		Origin:          origin,
		SessionId:       sessionId,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// Pending reports whether the request is still awaiting a decision.
func (r *Request) Pending() bool {
	return r.Status == StatusPending
}

// DecidableBy reports whether userId may be the one to decide this request.
//
// The rule is only that they are not the person who asked. Whether they hold
// the permission to grant the group at all is the authorization interceptor's
// question, asked before this one; this is the check that interceptor cannot
// make, because "may you grant this group" is true of the requester too when
// the requester is an administrator.
//
// Without it, an administrator could raise a request and approve it in the same
// breath, and the approval would attest to nothing.
func (r *Request) DecidableBy(userId string) bool {
	return userId != r.RequesterUserId
}
