// Package audit records who changed the authorization graph, how they reached
// the server and what they asked for.
//
// The subject of a record is always a person. The assistant is a channel a
// person acts through, never an actor in its own right: "the AI did it" answers
// nothing, and a log that accepts that answer is not an audit log.
package audit

import (
	"context"
	"time"
)

// Channel is the route a change arrived by.
type Channel string

const (
	// ChannelAPI is a call the user made themselves - the web console, the hub
	// CLI, or a direct request. They are one channel because the server cannot
	// tell them apart in a way worth trusting; Client carries the client's own
	// claim about which it was, for reading, not for deciding.
	ChannelAPI Channel = "api"
	// ChannelAIChat is a call the assistant made on the user's behalf while
	// answering them. SessionId says in which conversation.
	ChannelAIChat Channel = "ai_chat"
)

// Entry is one recorded attempt.
type Entry struct {
	Id string
	// ActorUserId is the person responsible, including when Channel is
	// ChannelAIChat.
	ActorUserId string
	Channel     Channel
	// SessionId is the chat session the assistant was answering in. Empty for
	// every other channel.
	SessionId string
	// Resource and Action are the RBAC vocabulary of the operation, so a record
	// reads the same way a permission does.
	Resource string
	Action   string
	// TargetId is the id of the entity that was changed, when the operation
	// names one.
	TargetId string
	// Arguments is the request as it arrived, as JSON, with secrets removed.
	Arguments []byte
	// Succeeded is false for a refused or failed attempt, which is recorded
	// just as a successful one is.
	Succeeded bool
	// Error is the failure reported to the caller, empty when Succeeded.
	Error string
	// Client is what the caller said it was - a user agent string. Unverified.
	Client    string
	CreatedAt time.Time
}

// Repository stores audit records. It offers no update and no delete: a record
// that can be rewritten proves nothing. Reading is a query concern and lives in
// the use case, as it does for the other listings.
type Repository interface {
	Create(ctx context.Context, e *Entry) error
}

type sourceCtx struct{}

// Source is what the recording interceptor cannot work out from the request
// alone: which channel the call came in by.
type Source struct {
	Channel   Channel
	SessionId string
}

// NewContext marks ctx as belonging to a channel. Callers that dispatch on a
// user's behalf - the assistant's tool executor - set it; a request that
// arrives over the wire is ChannelAPI and needs nothing.
func NewContext(ctx context.Context, s Source) context.Context {
	return context.WithValue(ctx, sourceCtx{}, s)
}

// FromContext returns the channel ctx was marked with, defaulting to
// ChannelAPI. The default is the safe one: an unmarked call really is a call
// the user made directly.
func FromContext(ctx context.Context) Source {
	if s, ok := ctx.Value(sourceCtx{}).(Source); ok && s.Channel != "" {
		return s
	}
	return Source{Channel: ChannelAPI}
}
