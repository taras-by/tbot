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
	if p.User.Type == UserTelegram {
		return p.User.Name
	}
	return p.User.Name
}

func (p *Participant) Link() string {
	if p.User.Type == UserTelegram {
		return "@" + p.User.Name
	}
	return p.User.Name
}
