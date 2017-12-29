package auth

import (
	"errors"
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
func NewSession(uid int, password string) (Session, error) {
	// First, do some sanity checks before verification
	if len(password) == 0 {
		return "", ErrMissingCredentials
	}
	// Get the user
	user, err := provider.GetUserByID(uid)
	if err != nil {
		log.Error(0, "Could not get user with ID %d: %v", uid, err)
		return "", err
	}
	// Now, verify the password
	valid, err := ValidatePassword(password, user.Password)
	if err != nil {
		log.Error(0, "Password verification failed: %v", err)
		return "", err
	}
	if valid {
		return newUnverifiedSession(uid), nil
	}
	return "", ErrInvalidCredentials
}

// newUnverifiedSession issues a session token but does not verify the user's password
func newUnverifiedSession(uid int) Session {
	sess := Session(utils.RandomString(SessionTokenLength))
	sessions[uid] = append(sessions[uid], sess)
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

	return
}
