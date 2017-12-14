package auth

import (
	"errors"

	log "gopkg.in/clog.v1"
)

const SessionTokenLength = 16

var (
	provider CredentialsProvider

	ErrMissingCredentials = errors.New("auth: Missing credentials")
	ErrInvalidCredentials = errors.New("auth: Invalid credentials")
)

// Init intializes the auth package. You must call this before using any auth function.
func Init(cprovider CredentialsProvider) {
	provider = cprovider
	sessions = make(map[Session]struct{})
}

// Session represents a user session, denoted in a cryptographically secure string
type Session string

// TODO: we should probably store the sessions on the database, as re-logging every time the server
// restarts is kind of tedious.
var sessions map[Session]struct{}

// NewSession verifies the user's credentials and then returns a new Session
func NewSession(email string, password string) (Session, error) {
	// First, do some sanity checks before verification
	if len(email) == 0 || len(password) == 0 {
		return "", ErrMissingCredentials
	}
	// Now, verify the password using the credentials provider
	validCredentials, err := provider.VerifyUserPassword(email, password)
	if err != nil {
		log.Error(0, "Could not create new session, because call to credentials provider failed: %v", err)
		return "", err
	}
	if validCredentials {
		return Session(make([]byte, SessionTokenLength, SessionTokenLength)), nil
	}
	return "", ErrInvalidCredentials
}

// // ValidateSession checks if the session is valid.
// func ValidateSession(sess Session) (valid bool, err error) {

// }
