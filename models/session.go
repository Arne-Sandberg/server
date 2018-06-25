package models

import (
	"fmt"
	"strconv"

)

const SessionTokenLength = 32

type Session struct {
	UserID    uint32 `gorm:"index"`
	Token     string `gorm:"primary_key"`
	ExpiresAt int64
}

func (s Session) GetTokenString() string {
	return fmt.Sprintf("%s%d", s.Token, s.UserID)
}

func ParseSessionTokenString(token string) (*Session, error) {
	if len(token) < SessionTokenLength {
		return nil, fmt.Errorf("given token '%s' is not long enough", token)
	}

	tok := token[:SessionTokenLength]
	userIDStr := token[SessionTokenLength:]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return &Session{}, err
	}
	return &Session{UserID: uint32(userID), Token: tok}, nil
}