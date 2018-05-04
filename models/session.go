package models

import (
	"fmt"
	"strconv"
	"time"
)

const SessionTokenLength = 32

type Session struct {
	UserID    uint32 `storm:"index"`
	Token     string `storm:"id,unique"`
	ExpiresAt time.Time
}

func (s Session) GetTokenString() string {
	return fmt.Sprintf("%s%d", s.Token, s.UserID)
}

func ParseSessionTokenString(cookie string) (*Session, error) {
	tok := cookie[:SessionTokenLength]
	userIDStr := cookie[SessionTokenLength:]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return &Session{}, err
	}
	return &Session{UserID: uint32(userID), Token: tok}, nil
}
