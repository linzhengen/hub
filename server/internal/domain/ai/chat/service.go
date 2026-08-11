package chat

import "context"

// Service is the LLM backend interface. Implementations wrap a specific
// provider (e.g. Anthropic Claude) and stream text deltas on the returned
// channel. The channel is closed when the response is complete or on error.
type Service interface {
	Send(ctx context.Context, messages []*Message) (<-chan Delta, error)
}
