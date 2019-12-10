package store

import "strings"

func Escape(s string) string {
	s = strings.ReplaceAll(s, "@", "-")
	s = strings.ReplaceAll(s, "*", "-")
	s = strings.ReplaceAll(s, "`", "-")
	s = strings.ReplaceAll(s, "_", "-")
	return s
}