package models

// SessionTokenLength defines the length of a token
const SessionTokenLength = 32

// Session represents one session for an user
type Session struct {
	Token     string `fc_neo:"token" fc_neo_unique:""`
	ExpiresAt int64  `fc_neo:"expires_at"`
}
