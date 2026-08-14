// Package mock answers chat requests from a canned script instead of calling
// Anthropic, so `make dev` needs neither an API key nor network access and a
// developer working on the chat UI is not billed for every reload.
//
// It is selected by ANTHROPIC_MOCK=true; see config.Anthropic.
package mock

import (
	"context"
	"fmt"
	"strings"
	"time"

	chatDomain "github.com/linzhengen/hub/server/internal/domain/ai/chat"
)

// ErrorTrigger lets a developer exercise the streaming error path - the part of
// the UI that is hardest to reach against a healthy upstream - by sending a
// message that contains it.
const ErrorTrigger = "!error"

type client struct {
	model string
	delay time.Duration
}

// New returns a chat.Service that streams a scripted reply. delay is the pause
// between deltas; it makes the stream observable in the browser and can be set
// to zero in tests.
func New(model string, delay time.Duration) chatDomain.Service {
	return &client{model: model, delay: delay}
}

func (c *client) Send(ctx context.Context, messages []*chatDomain.Message) (<-chan chatDomain.Delta, error) {
	prompt := lastUserContent(messages)
	turn := countUserMessages(messages)

	ch := make(chan chatDomain.Delta, 64)
	go func() {
		defer close(ch)

		if strings.Contains(prompt, ErrorTrigger) {
			send(ctx, ch, chatDomain.Delta{
				Error: fmt.Errorf("mock llm: simulated upstream failure (triggered by %q)", ErrorTrigger),
			})
			return
		}

		for _, part := range split(c.reply(prompt, turn)) {
			if !c.pause(ctx) {
				return
			}
			if !send(ctx, ch, chatDomain.Delta{Text: part}) {
				return
			}
		}
		send(ctx, ch, chatDomain.Delta{Done: true})
	}()

	return ch, nil
}

// reply is deliberately made of Markdown - headings, a list, a fenced block -
// because the chat UI has to render all of it and a plain sentence would not
// exercise that.
func (c *client) reply(prompt string, turn int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "**%s (mock)** ・ turn %d\n\n", c.model, turn)

	if prompt == "" {
		b.WriteString("You sent no text, so there is nothing to echo back.\n\n")
	} else {
		b.WriteString("You said:\n\n")
		b.WriteString(quote(prompt))
		b.WriteString("\n\n")
	}

	b.WriteString("This answer comes from the mock LLM backend, not from Anthropic. ")
	b.WriteString("It streams in several deltas so the client behaves exactly as it does against the real API.\n\n")
	b.WriteString("- Send a message containing `" + ErrorTrigger + "` to make the stream fail.\n")
	b.WriteString("- Set `ANTHROPIC_MOCK_DELAY` to change how fast the deltas arrive.\n")
	b.WriteString("- Talk to the real API instead:\n\n")
	b.WriteString("```sh\nANTHROPIC_MOCK=false ANTHROPIC_API_KEY=sk-ant-... make dev\n```\n")

	return b.String()
}

// pause waits between deltas, reporting whether the stream should continue.
func (c *client) pause(ctx context.Context) bool {
	if c.delay <= 0 {
		return ctx.Err() == nil
	}
	timer := time.NewTimer(c.delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}

func send(ctx context.Context, ch chan<- chatDomain.Delta, d chatDomain.Delta) bool {
	select {
	case ch <- d:
		return true
	case <-ctx.Done():
		return false
	}
}

// split cuts the reply after every space, which keeps whitespace and newlines
// attached to the delta before them: concatenating the parts rebuilds the input.
func split(s string) []string {
	parts := strings.SplitAfter(s, " ")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func quote(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = "> " + l
	}
	return strings.Join(lines, "\n")
}

func lastUserContent(messages []*chatDomain.Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == chatDomain.RoleUser {
			return messages[i].Content
		}
	}
	return ""
}

func countUserMessages(messages []*chatDomain.Message) int {
	n := 0
	for _, m := range messages {
		if m.Role == chatDomain.RoleUser {
			n++
		}
	}
	return n
}

// compile-time guard: the package must keep implementing the domain interface.
var _ chatDomain.Service = (*client)(nil)
