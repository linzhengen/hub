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

type Delta struct {
	Text  string
	Done  bool
	Error error
}
