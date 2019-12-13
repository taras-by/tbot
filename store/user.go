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
	UserTelegram UserType = "telegram"
	UserGuest    UserType = "guest"
)

func (u User) Uid() string {
	if u.Id != "" {
		return u.Id
	}
	return fmt.Sprintf("%x", md5.Sum([]byte(u.UserName)))
}
