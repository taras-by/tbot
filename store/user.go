package store

import (
	"crypto/md5"
	"fmt"
)

type User struct {
	Id        string
	UserName  string
	FirstName string
	LastName  string
	Type      UserType
}

type UserType string

const (
	UserTelegram   UserType = "telegram"
	UserUnresolved UserType = "unresolved"
	UserGuest      UserType = "guest"
)

func (u User) Uid() string {
	if u.Id != "" {
		return u.Id
	}
	return fmt.Sprintf("%s:%x", u.Type, md5.Sum([]byte(u.UserName)))
}

func (u *User) Name() string {
	if u.Type == UserTelegram {
		if u.FirstName != "" {
			if u.LastName != "" {
				return u.FirstName + " " + u.LastName
			}
			return u.FirstName
		}
	}
	return u.UserName
}

func (u *User) Link() string {
	if u.Type == UserGuest {
		return u.UserName
	}
	return "@" + u.UserName
}
