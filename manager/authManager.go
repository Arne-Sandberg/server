package manager

import (
	"errors"
	"time"

	log "gopkg.in/clog.v1"

	"github.com/freecloudio/freecloud/manager/auth"
	"github.com/freecloudio/freecloud/models"
	"github.com/freecloudio/freecloud/utils"
)

const (
	sessionExpiry      = 24 * time.Hour
	sessionTokenLength = 32 // characters
)

var (
	ErrMissingCredentials = errors.New("auth: Missing credentials")
	ErrInvalidCredentials = errors.New("auth: Invalid credentials")
	ErrInvalidUserData    = errors.New("auth: Invalid user data")
	ErrUserAlreadyExists  = errors.New("auth: User already exists")
)

// SessionProvider defines a (mostly) CRUD interface for Sessions.
// This should be implemented by some sort of persistant storage like a database
type SessionProvider interface {
	CreateSession(session *models.Session) error
	ReadSessionByToken(token string) (*models.Session, error)
	ReadSessionCount() (int, error)
	DeleteSession(session *models.Session) error
	DeleteSessionsByUser(userID int64) error
	DeleteExpiredSessions() error
	SessionIsValid(session *models.Session) bool
}

// CredentialsProvider defines a (mostly) CRUD interface for credentials and user accounts.
// This is to be implemented by a persistent storage like a database.
type CredentialsProvider interface {
	CreateUser(user *models.User) error
	ReadAdminCount() (int, error)
	ReadAllUsers() ([]*models.User, error)
	ReadUserByID(userID int64) (*models.User, error)
	ReadUserByEmail(email string) (*models.User, error)
	ReadUserExistsByEmail(email string) (bool, error)
	UpdateUser(user *models.User) error
	DeleteUser(userID int64) error
}

// AuthManager has methods for authenticating users.
type AuthManager struct {
	sessionProvider     SessionProvider
	credentialsProvider CredentialsProvider
	done                chan struct{}
}

// NewAuthManager creates a new AuthManager which can be used immediately
func NewAuthManager(sessionProvider SessionProvider, credentialsProvider CredentialsProvider) *AuthManager {
	mgr := &AuthManager{
		sessionProvider:     sessionProvider,
		credentialsProvider: credentialsProvider,
		done:                make(chan struct{}),
	}
	go mgr.cleanupExpiredSessionsRoutine(1 * time.Hour)
	return mgr
}

// Close is used to end running tasks
func (mgr *AuthManager) Close() {
	mgr.done <- struct{}{}
}

func (mgr *AuthManager) cleanupExpiredSessionsRoutine(interval time.Duration) {
	log.Trace("Session cleaner will run every %v", interval)
	mgr.sessionProvider.DeleteExpiredSessions()
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-mgr.done:
			return
		case <-ticker.C:
			log.Trace("Cleaning expired sessions")
			mgr.sessionProvider.DeleteExpiredSessions()
		}
	}
}

// NewSession verifies the user's credentials and then returns a new session.
func (mgr *AuthManager) NewSession(email string, password string) (*models.Session, error) {
	// First, do some sanity checks so we can reduce calls to the credentials provider with obviously wrong data.
	if len(password) == 0 || len(email) == 0 {
		return nil, ErrMissingCredentials
	}

	user, err := mgr.credentialsProvider.ReadUserByEmail(email)
	if err != nil {
		log.Error(0, "Could not get user via email %s: %v", email, err)
		return nil, err
	}
	valid, err := auth.ValidateScryptPassword(password, user.Password)
	if err != nil {
		log.Error(0, "Password verification failed for user %s: %v", user.Email, err)
		return nil, err
	}
	if valid {
		return mgr.newUnverifiedSession(user.ID)
	}
	return &models.Session{}, ErrInvalidCredentials
}

func (mgr *AuthManager) newUnverifiedSession(userID int64) (*models.Session, error) {
	session := &models.Session{
		UserID:    userID,
		Token:     utils.RandomString(sessionTokenLength),
		ExpiresAt: time.Now().UTC().Add(sessionExpiry).Unix(),
	}

	err := mgr.sessionProvider.CreateSession(session)
	if err != nil {
		log.Error(0, "Could not store session: %v", err)
		return nil, err
	}

	err = mgr.UpdateLastSession(userID)
	if err != nil {
		log.Error(0, "Could not store the session as user's latest: %v", err)
		return nil, err
	}

	return session, nil
}

// CreateUser validates a new user's data, hashes his password and then stores them.
// Also, a new session is returned for the given user.
func (mgr *AuthManager) CreateUser(user *models.User) (session *models.Session, err error) {
	if !utils.ValidateEmail(user.Email) ||
		!utils.ValidatePassword(user.Password) ||
		!utils.ValidateFirstName(user.FirstName) ||
		!utils.ValidateLastName(user.LastName) {
		return nil, ErrInvalidUserData
	}

	userExists, err := mgr.credentialsProvider.ReadUserExistsByEmail(user.Email)
	if err != nil {
		// Don't bail out here, since this will be checked again when creating the
		// user in the credentialsProvider.
		log.Warn("Could not validate whether user with email %s already exists", user.Email)
	}
	if userExists {
		return nil, ErrUserAlreadyExists
	}

	user.Password, err = auth.HashScrypt(user.Password)
	if err != nil {
		log.Error(0, "Password hashing failed: %v", err)
		return nil, err
	}

	// Save the user. This also fills their ID
	err = mgr.credentialsProvider.CreateUser(user)
	if err != nil {
		log.Error(0, "Creating user failed: %v", err)
		return nil, err
	}

	// If this is the first user (ID 1) they will become an admin
	if user.ID == 1 {
		log.Trace("Making first user an admin")
		user.IsAdmin = true
		err = mgr.credentialsProvider.UpdateUser(user)
		if err != nil {
			log.Error(0, "Could not make first user an admin: %v", err)
			// Since a system without an admin won't properly work, bail out
			return nil, err
		}
	}

	// Now, create a session for the user
	return mgr.newUnverifiedSession(user.ID)

}

func (mgr *AuthManager) DeleteUser(userID int64) (err error) {
	if err = mgr.sessionProvider.DeleteSessionsByUser(userID); err != nil {
		// TODO: log me
		return
	}

	if err = mgr.credentialsProvider.DeleteUser(userID); err != nil {
		// TODO: log me
		return
	}
	return
}

func (mgr *AuthManager) GetAllUsers(isAdmin bool) ([]*models.User, error) {
	users, err := mgr.credentialsProvider.ReadAllUsers()
	if err != nil {
		log.Error(0, "Could not get all users, %v:", err)
		return nil, err
	}
	for _, user := range users {
		// Mask out the password
		user.Password = ""

		// For normal users also mask out created, updated and lastSession
		if !isAdmin {
			user.CreatedAt = 0
			user.UpdatedAt = 0
			user.LastSessionAt = 0
		}
	}
	return users, nil
}

// ValidateSession checks if the session is valid.
func (mgr *AuthManager) ValidateSession(sess *models.Session) (valid bool) {
	return mgr.sessionProvider.SessionIsValid(sess)
}

func (mgr *AuthManager) ReadUserByID(userID int64) (*models.User, error) {
	return mgr.credentialsProvider.ReadUserByID(userID)
}

func (mgr *AuthManager) ReadUserByEmail(email string) (*models.User, error) {
	return mgr.credentialsProvider.ReadUserByEmail(email)
}

// RemoveSession removes the session from the session provider
func (mgr *AuthManager) DeleteSession(session *models.Session) (err error) {
	return mgr.sessionProvider.DeleteSession(session)
}

func (mgr *AuthManager) UpdateLastSession(userID int64) (err error) {
	user, err := mgr.ReadUserByID(userID)
	if err != nil {
		return
	}

	user.LastSessionAt = time.Now().UTC().Unix()
	err = mgr.credentialsProvider.UpdateUser(user)

	return
}

// ReadAdminCount returns the count of admin users
func (mgr *AuthManager) ReadAdminCount() (int, error) {
	return mgr.credentialsProvider.ReadAdminCount()
}
