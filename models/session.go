package models

import (
	"fmt"
	"strconv"
)

const SessionTokenLenght = 32

type Session struct {
	UID   int    `storm:"index"`
	Token string `storm:"id,unique"`
}

func (s Session) GetCookieString() string {
	return fmt.Sprintf("%s%d", s.Token, s.UID)
}

func ParseSessionCookieString(cookie string) (Session, error) {
	tok := cookie[:SessionTokenLenght]
	uidStr := cookie[SessionTokenLenght:]
	uid, err := strconv.Atoi(uidStr)
	if err != nil {
		return Session{}, err
	}
	return Session{uid, tok}, nil
}
