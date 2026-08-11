package claude

import (
	"context"
	"errors"
	"io"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	chatDomain "github.com/linzhengen/hub/server/internal/domain/ai/chat"
	"github.com/linzhengen/hub/server/pkg/logger"
)

type client struct {
	c     anthropic.Client
	model anthropic.Model
}

func New(apiKey, model string) chatDomain.Service {
	c := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &client{
		c:     c,
		model: anthropic.Model(model),
	}
}

func (cl *client) Send(ctx context.Context, messages []*chatDomain.Message) (<-chan chatDomain.Delta, error) {
	params := anthropic.MessageNewParams{
		Model:     cl.model,
		MaxTokens: 4096,
		Messages:  toAnthropicMessages(messages),
	}

	stream := cl.c.Messages.NewStreaming(ctx, params)

	ch := make(chan chatDomain.Delta, 64)
	go func() {
		defer close(ch)
		for stream.Next() {
			event := stream.Current()
			switch e := event.AsAny().(type) {
			case anthropic.ContentBlockDeltaEvent:
				if delta, ok := e.Delta.AsAny().(anthropic.TextDelta); ok && delta.Text != "" {
					select {
					case ch <- chatDomain.Delta{Text: delta.Text}:
					case <-ctx.Done():
						return
					}
				}
			}
		}
		if err := stream.Err(); err != nil && !errors.Is(err, io.EOF) {
			logger.Errorf("claude stream error: %v", err)
			select {
			case ch <- chatDomain.Delta{Error: err}:
			case <-ctx.Done():
			}
			return
		}
		select {
		case ch <- chatDomain.Delta{Done: true}:
		case <-ctx.Done():
		}
	}()

	return ch, nil
}

func toAnthropicMessages(messages []*chatDomain.Message) []anthropic.MessageParam {
	params := make([]anthropic.MessageParam, 0, len(messages))
	for _, m := range messages {
		switch m.Role {
		case chatDomain.RoleUser:
			params = append(params, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
		case chatDomain.RoleAssistant:
			params = append(params, anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content)))
		}
	}
	return params
}
