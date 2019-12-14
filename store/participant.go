package store

import (
	"time"
)

type Participant struct {
	User   User
	Time   time.Time
	ChatId int64
}

func (p *Participant) Id() string {
	return p.User.Uid()
}

func (p *Participant) Name() string {
	return p.User.Name()
}

func (p *Participant) Link() string {
	return p.User.Link()
}

func (p *Participant) IsUnresolved() bool {
	return p.User.Type == UserUnresolved
}
