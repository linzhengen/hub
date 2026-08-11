package chat

import "time"

type Session struct {
	Id        string
	UserId    string
	Title     string
	CreatedAt time.Time
}
