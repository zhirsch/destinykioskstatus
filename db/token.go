package db

import "time"

type Token struct {
	Value   string
	Ready   time.Time
	Expires time.Time
}

func (t Token) IsReady() bool {
	return time.Now().After(t.Ready)
}

func (t Token) IsExpired() bool {
	return time.Now().After(t.Expires)
}
