package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/riesinger/freecloud/models"
	"github.com/riesinger/freecloud/utils"

	log "gopkg.in/clog.v1"
)

const SessionTokenLength = 32

var (
	provider CredentialsProvider

	ErrMissingCredentials = errors.New("auth: Missing credentials")
	ErrInvalidCredentials = errors.New("auth: Invalid credentials")
)

// Init intializes the auth package. You must call this before using any auth function.
func Init(cprovider CredentialsProvider) {
	provider = cprovider
	sessions = make(map[int][]Session)
}

// Session represents a user session, denoted in a cryptographically secure string
type Session string

var sessions map[int][]Session

// NewSession verifies the user's credentials and then returns a new Session
func NewSession(email string, password string) (Session, int, error) {
	// First, do some sanity checks before verification
	if len(password) == 0 {
		return "", -1, ErrMissingCredentials
	}
	// Get the user
	user, err := provider.GetUserByEmail(email)
	if err != nil {
		log.Error(0, "Could not get user with email %s: %v", email, err)
		return "", -1, err
	}
	// Now, verify the password
	valid, err := ValidatePassword(password, user.Password)
	if err != nil {
		log.Error(0, "Password verification failed: %v", err)
		return "", -1, err
	}
	if valid {
		return newUnverifiedSession(user.ID), user.ID, nil
	}
	return "", -1, ErrInvalidCredentials
}

// newUnverifiedSession issues a session token but does not verify the user's password
func newUnverifiedSession(uid int) Session {
	sess := Session(utils.RandomString(SessionTokenLength))
	sessions[uid] = append(sessions[uid], sess)
	log.Trace("Sessions: %v", sessions)
	return sess
}

// NewUser hashes the user's password, saves it to the database and then creates a new session, so he doesn't have to login again.
func NewUser(user *models.User) (session Session, err error) {
	user.Created = time.Now().UTC()
	user.Updated = time.Now().UTC()
	user.Password, err = HashPassword(user.Password)
	if err != nil {
		log.Error(0, "Password hashing failed: %v", err)
		return
	}

	// Save the user. This also fills its ID
	err = provider.CreateUser(user)
	if err != nil {
		return
	}

	// Now, create a session for the user
	return newUnverifiedSession(user.ID), nil
}

// ValidateSession checks if the session is valid.
func ValidateSession(userID int, sess Session) (valid bool, err error) {
	userSessions, ok := sessions[userID]
	if !ok {
		err = fmt.Errorf("no sessions for this user")
		return
	}
	for _, v := range userSessions {
		if v == sess {
			valid = true
			err = nil
			return
		}
	}
	return
}
