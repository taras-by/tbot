package store

import (
	"crypto/md5"
	"encoding/hex"
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
	id := u.Id
	if id == "" {
		s := md5.Sum([]byte(u.UserName))
		id = hex.EncodeToString(s[:])
	}
	return fmt.Sprintf("%s:%s", u.Type, id)
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
	if u.UserName == "" {
		return u.Name()
	}
	if u.Type == UserGuest {
		return u.UserName
	}
	return "@" + u.UserName
}
