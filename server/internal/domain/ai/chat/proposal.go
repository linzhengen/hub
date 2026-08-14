package chat

import "time"

// ProposalStatus is where a proposed write has got to.
type ProposalStatus string

const (
	// ProposalPending is waiting on the user.
	ProposalPending ProposalStatus = "pending"
	// ProposalApproved was approved and has been run.
	ProposalApproved ProposalStatus = "approved"
	// ProposalRejected was declined; nothing ran.
	ProposalRejected ProposalStatus = "rejected"
)

// ToolProposal is a change the assistant wants to make and has not made.
//
// A read runs the moment the model asks for it. A write does not: the model
// only ever gets as far as proposing one, the stream stops, and the change
// happens when - and only when - the user says so.
//
// The reason is not that the model might exceed the user's permissions; it
// cannot, since every call goes through the same authorization the user's own
// requests do. It is that tool results carry text other people wrote - group
// descriptions, resource metadata, user names - and that text is in a position
// to talk the model into proposing something nobody asked for. Putting a person
// between the proposal and the change is what breaks that path.
type ToolProposal struct {
	Id        string
	SessionId string
	// ToolName and Arguments are what the user is shown. They are the real
	// operation and the real arguments, not a summary of them: "add 3 users to
	// a group" is not enough to approve, "add_users_to_group(id=..., ...)" is.
	ToolName  string
	Arguments string
	// Continuation is state the Service produced and only the Service can read:
	// what it needs to take the conversation back up from where it stopped. It
	// is opaque here on purpose, so the shape of a provider's message format
	// does not become part of the domain.
	Continuation []byte
	Status       ProposalStatus
	CreatedAt    time.Time
	DecidedAt    *time.Time
}

// Pending reports whether the proposal is still awaiting a decision. A decided
// proposal is never acted on again, which is what stops one approval from being
// replayed into several changes.
func (p *ToolProposal) Pending() bool {
	return p.Status == ProposalPending
}
