package models

import (
	"fmt"
	"strconv"
	"time"
)

const SessionTokenLength = 32

type Session struct {
	UID       int    `storm:"index"`
	Token     string `storm:"id,unique"`
	ExpiresAt time.Time
}

func (s Session) GetTokenString() string {
	return fmt.Sprintf("%s%d", s.Token, s.UID)
}

func ParseSessionTokenString(cookie string) (Session, error) {
	tok := cookie[:SessionTokenLength]
	uidStr := cookie[SessionTokenLength:]
	uid, err := strconv.Atoi(uidStr)
	if err != nil {
		return Session{}, err
	}
	return Session{UID: uid, Token: tok}, nil
}
