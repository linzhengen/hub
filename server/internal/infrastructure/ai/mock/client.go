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

// ToolTrigger makes the mock call a tool before answering, so the tool path -
// permission filtering, dispatch, redaction - can be exercised without an API
// key. The real model decides for itself when to call one; the mock cannot, so
// it is asked explicitly.
const ToolTrigger = "!tool"

type client struct {
	model string
	delay time.Duration
	tools chatDomain.ToolBox
}

// New returns a chat.Service that streams a scripted reply. delay is the pause
// between deltas; it makes the stream observable in the browser and can be set
// to zero in tests.
func New(model string, delay time.Duration, tools chatDomain.ToolBox) chatDomain.Service {
	return &client{model: model, delay: delay, tools: tools}
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

		for _, part := range split(c.reply(ctx, prompt, turn)) {
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
func (c *client) reply(ctx context.Context, prompt string, turn int) string {
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
	if strings.Contains(prompt, ToolTrigger) {
		b.WriteString(c.callFirstTool(ctx))
		b.WriteString("\n\n")
	}

	b.WriteString("- Send a message containing `" + ErrorTrigger + "` to make the stream fail.\n")
	b.WriteString("- Send a message containing `" + ToolTrigger + "` to make it call a tool.\n")
	b.WriteString("- Set `ANTHROPIC_MOCK_DELAY` to change how fast the deltas arrive.\n")
	b.WriteString("- Talk to the real API instead:\n\n")
	b.WriteString("```sh\nANTHROPIC_MOCK=false ANTHROPIC_API_KEY=sk-ant-... make dev\n```\n")

	return b.String()
}

// announceTool emits the same frame the real client emits before a lookup, so
// the client's tool display can be built and tested without an API key.
func (c *client) announceTool(ctx context.Context, ch chan<- chatDomain.Delta) {
	if c.tools == nil {
		return
	}
	tools, err := c.tools.Tools(ctx)
	if err != nil || len(tools) == 0 {
		return
	}
	send(ctx, ch, chatDomain.Delta{Tool: &chatDomain.ToolCall{
		Name:      tools[0].Name,
		Arguments: "{}",
	}})
}

// callFirstTool runs the first tool the caller is allowed to use and reports
// what came back, which is enough to see that listing, permission filtering,
// dispatch and redaction all work end to end.
func (c *client) callFirstTool(ctx context.Context) string {
	if c.tools == nil {
		return "No tool box is wired up."
	}
	tools, err := c.tools.Tools(ctx)
	if err != nil {
		return fmt.Sprintf("Listing tools failed: %v", err)
	}
	if len(tools) == 0 {
		return "You are not permitted to call any tool."
	}

	names := make([]string, len(tools))
	for i, tool := range tools {
		names[i] = tool.Name
	}

	result, err := c.tools.Call(ctx, tools[0].Name, nil)
	if err != nil {
		return fmt.Sprintf("Tools you may call: %s\n\nCalling `%s` failed: %v",
			strings.Join(names, ", "), tools[0].Name, err)
	}
	return fmt.Sprintf("Tools you may call: %s\n\n`%s` returned:\n\n```json\n%s\n```",
		strings.Join(names, ", "), tools[0].Name, result)
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
