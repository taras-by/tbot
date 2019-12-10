package store

import (
	"strings"
)

func Escape(s string) string {
	r := strings.NewReplacer("@", "-", "*", "-", "`", "-", "_", "-")
	return r.Replace(s)
}
