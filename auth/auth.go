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
	cProvider CredentialsProvider
	sProvider SessionProvider

	ErrMissingCredentials = errors.New("auth: Missing credentials")
	ErrInvalidCredentials = errors.New("auth: Invalid credentials")
)

// Init intializes the auth package. You must call this before using any auth function.
func Init(credentialsProvider CredentialsProvider, sessionProvider SessionProvider) {
	cProvider = credentialsProvider
	sProvider = sessionProvider
}

// NewSession verifies the user's credentials and then returns a new Session
func NewSession(email string, password string) (models.Session, error) {
	// First, do some sanity checks before verification
	if len(password) == 0 {
		return models.Session{}, ErrMissingCredentials
	}
	// Get the user
	user, err := cProvider.GetUserByEmail(email)
	if err != nil {
		log.Error(0, "Could not get user with email %s: %v", email, err)
		return models.Session{}, err
	}
	// Now, verify the password
	valid, err := ValidatePassword(password, user.Password)
	if err != nil {
		log.Error(0, "Password verification failed: %v", err)
		return models.Session{}, err
	}
	if valid {
		return newUnverifiedSession(user.ID), nil
	}
	return models.Session{}, ErrInvalidCredentials
}

// newUnverifiedSession issues a session token but does not verify the user's password
func newUnverifiedSession(uid int) models.Session {
	sess := models.Session{UID: uid, Token: utils.RandomString(SessionTokenLength)}
	err := sProvider.StoreSession(sess)
	if err != nil {
		log.Error(0, "Could not store session: %v", err)
	}
	return sess
}

// NewUser hashes the user's password, saves it to the database and then creates a new session, so he doesn't have to login again.
func NewUser(user *models.User) (session models.Session, err error) {
	user.Created = time.Now().UTC()
	user.Updated = time.Now().UTC()
	user.Password, err = HashPassword(user.Password)
	if err != nil {
		log.Error(0, "Password hashing failed: %v", err)
		return
	}

	// Save the user. This also fills its ID
	err = cProvider.CreateUser(user)
	if err != nil {
		return
	}

	// Now, create a session for the user
	return newUnverifiedSession(user.ID), nil
}

// ValidateSession checks if the session is valid.
func ValidateSession(sess models.Session) (valid bool) {
	return sProvider.SessionIsValid(sess)
}

func GetUserByID(uid int) (*models.User, error) {
	return cProvider.GetUserByID(uid)
}
