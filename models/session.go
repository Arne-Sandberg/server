package models

// SessionTokenLength defines the length of a token
const SessionTokenLength = 32

// Session represents one session for an user
type Session struct {
	Token     string `fc_neo:"token"`
	ExpiresAt int64  `fc_neo:"expires_at"`
}
