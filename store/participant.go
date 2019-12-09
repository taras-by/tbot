package store

import "time"

type Participant struct {
	User   User
	Time   time.Time
	ChatId int64
}
