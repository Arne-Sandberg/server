package models

import (
	"fmt"
	"strconv"
)

// SessionTokenLength defines the length of a token
const SessionTokenLength = 32

// Session represents one session for an user
type Session struct {
	UserID    int64  `gorm:"index"`
	Token     string `gorm:"primary_key"`
	ExpiresAt int64
}

// GetSessionString assembles the session string for the frontend
func (s Session) GetSessionString() string {
	return fmt.Sprintf("%s%d", s.Token, s.UserID)
}

// ParseSessionString parses a given session string into a session
func ParseSessionString(token string) (*Session, error) {
	if len(token) < SessionTokenLength {
		return nil, fmt.Errorf("given token '%s' is not long enough", token)
	}

	tok := token[:SessionTokenLength]
	userIDStr := token[SessionTokenLength:]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return &Session{}, err
	}
	return &Session{UserID: int64(userID), Token: tok}, nil
}
