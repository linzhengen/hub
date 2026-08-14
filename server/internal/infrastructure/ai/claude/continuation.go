package claude

import (
	"encoding/json"

	"github.com/anthropics/anthropic-sdk-go"
)

// continuation is what a paused conversation needs to be taken back up.
//
// It holds only the assistant turn that asked - the one turn not yet in the
// stored history - because everything before it is already persisted as chat
// messages and is passed back in. The provider's own parameter types are
// deliberately not serialised: they marshal for sending but do not round trip,
// and pinning a vendor's wire structs into stored bytes would make an SDK
// upgrade a data migration.
type continuation struct {
	// Text is what the assistant said before asking.
	Text string
	// Uses is every tool the assistant asked for in that turn. All of them need
	// a result before the conversation can go on, even the ones the user was
	// not asked about, so all of them are kept.
	Uses []toolUse
	// ProposedID is the tool_use the user was asked to approve.
	ProposedID string
}

type toolUse struct {
	ID    string
	Name  string
	Input json.RawMessage
	Write bool
}

func (c continuation) encode() ([]byte, error) {
	return json.Marshal(c)
}

func decodeContinuation(raw []byte) (continuation, error) {
	var c continuation
	err := json.Unmarshal(raw, &c)
	return c, err
}

// assistantTurn rebuilds the message the assistant sent, which has to go back
// into the conversation verbatim: a tool_result is only valid as an answer to
// the tool_use block that precedes it.
func (c continuation) assistantTurn() anthropic.MessageParam {
	blocks := make([]anthropic.ContentBlockParamUnion, 0, len(c.Uses)+1)
	if c.Text != "" {
		blocks = append(blocks, anthropic.NewTextBlock(c.Text))
	}
	for _, use := range c.Uses {
		blocks = append(blocks, anthropic.ContentBlockParamUnion{
			OfToolUse: &anthropic.ToolUseBlockParam{
				ID:    use.ID,
				Name:  use.Name,
				Input: use.Input,
			},
		})
	}
	return anthropic.NewAssistantMessage(blocks...)
}
