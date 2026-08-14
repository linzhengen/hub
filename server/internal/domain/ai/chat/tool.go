package chat

import "context"

// Tool is one hub operation the assistant may call.
//
// InputSchema is a JSON Schema object describing the arguments, which is what
// the LLM is given and what its generated arguments are validated against.
type Tool struct {
	Name        string
	Description string
	InputSchema map[string]any
}

// ToolBox is the set of operations the assistant can perform for a user.
//
// Both methods act as the user in the context: Tools lists only what that user
// is allowed to call, and Call refuses anything they are not. The two are
// deliberately separate checks - the list keeps the model from proposing work
// the user cannot do, and the check inside Call is what actually holds, since
// a permission can be revoked between the two.
type ToolBox interface {
	Tools(ctx context.Context) ([]Tool, error)
	// Call runs one tool and returns its result as JSON for the model to read.
	// A refusal or a bad argument is returned as an error, not a panic: the
	// model is expected to read it and try something else.
	Call(ctx context.Context, name string, args []byte) (string, error)
}
