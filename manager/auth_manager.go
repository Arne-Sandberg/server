package manager

import (
	"errors"
	"time"

	log "gopkg.in/clog.v1"

	"github.com/freecloudio/freecloud/crypt"
	"github.com/freecloudio/freecloud/models"
	"github.com/freecloudio/freecloud/repository"
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

// AuthManager has methods for authenticating users.
type AuthManager struct {
	sessionRep *repository.SessionRepository
	userRep    *repository.UserRepository
	done       chan struct{}
}

var authManager *AuthManager

// CreateAuthManager creates a new singleton AuthManager which can be used immediately
func CreateAuthManager(sessionRep *repository.SessionRepository, userRep *repository.UserRepository) *AuthManager {
	if authManager != nil {
		return authManager
	}

	authManager = &AuthManager{
		sessionRep: sessionRep,
		userRep:    userRep,
		done:       make(chan struct{}),
	}
	go authManager.cleanupExpiredSessionsRoutine(1 * time.Hour)
	return authManager
}

// GetAuthManager returns the singleton instance of the AuthManager
func GetAuthManager() *AuthManager {
	return authManager
}

// Close is used to end running tasks
func (mgr *AuthManager) Close() {
	mgr.done <- struct{}{}
}

func (mgr *AuthManager) cleanupExpiredSessionsRoutine(interval time.Duration) {
	log.Trace("Session cleaner will run every %v", interval)
	mgr.sessionRep.DeleteExpired()
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-mgr.done:
			return
		case <-ticker.C:
			log.Trace("Cleaning expired sessions")
			mgr.sessionRep.DeleteExpired()
		}
	}
}

// NewSession verifies the user's credentials and then returns a new session.
func (mgr *AuthManager) NewSession(email string, password string) (*models.Session, error) {
	// First, do some sanity checks so we can reduce calls to the credentials provider with obviously wrong data.
	if len(password) == 0 || len(email) == 0 {
		return nil, ErrMissingCredentials
	}

	user, err := mgr.userRep.GetByEmail(email)
	if err != nil {
		log.Error(0, "Could not get user via email %s: %v", email, err)
		return nil, err
	}
	valid, err := crypt.ValidateScryptPassword(password, user.Password)
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

	err := mgr.sessionRep.Create(session)
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

	existingUser, err := mgr.userRep.GetByEmail(user.Email)
	if err != nil && !repository.IsRecordNotFoundError(err) {
		// Don't bail out here, since this will be checked again when creating the user in repository
		log.Warn("Could not validate whether user with email %s already exists", user.Email)
	} else if err == nil && existingUser != nil && existingUser.ID > 0 {
		return nil, ErrUserAlreadyExists
	}

	user.Password, err = crypt.HashScrypt(user.Password)
	if err != nil {
		log.Error(0, "Password hashing failed: %v", err)
		return nil, err
	}
	user.IsAdmin = false

	// Save the user. This also fills their ID
	err = mgr.userRep.Create(user)
	if err != nil {
		log.Error(0, "Creating user failed: %v", err)
		return nil, err
	}

	// If this is the first user (ID 1) they will become an admin
	if user.ID == 1 {
		log.Trace("Making first user an admin")
		user.IsAdmin = true
		err = mgr.userRep.Update(user)
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
	if err = mgr.sessionRep.DeleteForUser(userID); err != nil {
		// TODO: log me
		return
	}

	if err = mgr.userRep.Delete(userID); err != nil {
		// TODO: log me
		return
	}

	// Delete data?!

	return
}

func (mgr *AuthManager) GetAllUsers(isAdmin bool) ([]*models.User, error) {
	users, err := mgr.userRep.GetAll()
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
	storedSession, err := mgr.sessionRep.GetByToken(sess.Token)
	if err != nil {
		log.Warn("Could not read session via token, assuming invalid session")
		return false
	}
	if storedSession.UserID == sess.UserID && storedSession.ExpiresAt > time.Now().UTC().Unix() {
		return true
	}
	return false
}

func (mgr *AuthManager) GetUserByID(userID int64) (*models.User, error) {
	return mgr.userRep.GetByID(userID)
}

func (mgr *AuthManager) GetUserByEmail(email string) (*models.User, error) {
	return mgr.userRep.GetByEmail(email)
}

// RemoveSession removes the session from the session provider
func (mgr *AuthManager) DeleteSession(session *models.Session) (err error) {
	return mgr.sessionRep.Delete(session)
}

func (mgr *AuthManager) UpdateLastSession(userID int64) (err error) {
	user, err := mgr.GetUserByID(userID)
	if err != nil {
		return
	}

	user.LastSessionAt = time.Now().UTC().Unix()
	err = mgr.userRep.Update(user)

	return
}

// GetAdminCount returns the count of admin users
func (mgr *AuthManager) GetAdminCount() (int, error) {
	return mgr.userRep.AdminCount()
}

func (mgr *AuthManager) GetSessionCount() (int, error) {
	return mgr.sessionRep.Count()
}
