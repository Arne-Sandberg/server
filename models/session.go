package models

import (
	"fmt"
	"strconv"

	"github.com/golang/protobuf/ptypes/timestamp"
)

const SessionTokenLength = 32

type Session struct {
	UserID    uint32 `storm:"index"`
	Token     string `storm:"id,unique"`
	ExpiresAt *timestamp.Timestamp
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