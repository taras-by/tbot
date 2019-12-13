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
		if p.User.FirstName != "" {
			if p.User.LastName != "" {
				return p.User.FirstName + " " + p.User.LastName
			}
			return p.User.FirstName
		}
	}
	return p.User.UserName
}

func (p *Participant) Link() string {
	if p.User.Type == UserTelegram {
		return "@" + p.User.UserName
	}
	return p.User.UserName
}
