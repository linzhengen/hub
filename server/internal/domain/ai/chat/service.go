package chat

import "context"

// Service is the LLM backend interface. Implementations wrap a specific
// provider (e.g. Anthropic Claude) and stream text deltas on the returned
// channel. The channel is closed when the response is complete or on error.
//
// A conversation can stop in two ways: finished, or waiting on the user. The
// second happens when the model asks to change something - the implementation
// emits a Delta carrying a ToolProposal and stops, having changed nothing.
// Resume picks it back up once the user has decided.
type Service interface {
	Send(ctx context.Context, messages []*Message) (<-chan Delta, error)
	// Resume continues a conversation that stopped for approval.
	//
	// messages is the stored history, continuation the state the proposal
	// carried, and approved the user's answer. On false the model is told the
	// user declined and answers around it; nothing is run either way unless
	// approved is true.
	Resume(ctx context.Context, messages []*Message, continuation []byte, approved bool) (<-chan Delta, error)
}
