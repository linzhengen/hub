package chat

import "time"

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type Message struct {
	Id        string
	SessionId string
	Role      Role
	Content   string
	CreatedAt time.Time
}

// ToolCall reports a lookup the assistant made while answering.
type ToolCall struct {
	Name string
	// Arguments is the JSON object the assistant supplied.
	Arguments string
}

// Delta is one frame of a streamed answer: a piece of text, a lookup the
// assistant made on the way, a change it wants to make, the end of the answer,
// or a failure.
type Delta struct {
	Text  string
	Done  bool
	Error error
	// Tool is set on the frames that report a lookup. Such a frame carries no
	// Text, so accumulating the answer is unaffected by it.
	Tool *ToolCall
	// Proposal is set on the frame that stops the answer to ask permission for
	// a change. It is the last frame of that stream: nothing has been changed,
	// and nothing will be until the user decides.
	Proposal *ToolProposal
	// Tokens reports what a round of the conversation cost. A backend emits it
	// as each round completes, so a conversation that stops early - on a
	// proposal, or on a failure - has still accounted for what it spent.
	Tokens int64
}
