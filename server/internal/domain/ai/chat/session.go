package chat

import "time"

type Session struct {
	Id        string
	UserId    string
	Title     string
	CreatedAt time.Time
	// TokensUsed is what this conversation has spent so far.
	TokensUsed int64
}
