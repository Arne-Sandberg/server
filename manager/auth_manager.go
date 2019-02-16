package manager

import (
	"time"

	log "gopkg.in/clog.v1"

	"github.com/freecloudio/server/crypt"
	"github.com/freecloudio/server/models"
	"github.com/freecloudio/server/repository"
	"github.com/freecloudio/server/restapi/fcerrors"
	"github.com/freecloudio/server/utils"
)

const (
	sessionTokenLength = 32 // characters
)

// AuthManager has methods for authenticating users.
type AuthManager struct {
	sessionRep             *repository.SessionRepository
	userRep                *repository.UserRepository
	sessionExpiry          int
	sessionCleanupInterval int
	done                   chan struct{}
}

var authManager *AuthManager

// CreateAuthManager creates a new singleton AuthManager which can be used immediately, sessionExpiry and sessionCleanupInterval are in hours
func CreateAuthManager(sessionRep *repository.SessionRepository, userRep *repository.UserRepository, sessionExpiry, sessionCleanupInterval int) *AuthManager {
	if authManager != nil {
		return authManager
	}

	authManager = &AuthManager{
		sessionRep:             sessionRep,
		userRep:                userRep,
		sessionExpiry:          sessionExpiry,
		sessionCleanupInterval: sessionCleanupInterval,
		done:                   make(chan struct{}),
	}
	go authManager.cleanupExpiredSessionsRoutine()
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

func (mgr *AuthManager) cleanupExpiredSessionsRoutine() {
	log.Trace("Session cleaner will run every %v hours", mgr.sessionCleanupInterval)
	mgr.sessionRep.DeleteExpired()
	ticker := time.NewTicker(time.Hour * time.Duration(mgr.sessionCleanupInterval))
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

// CreateUser validates a new user's data, hashes his password and then stores them.
// Also, a new session is returned for the given user.
func (mgr *AuthManager) CreateUser(user *models.User) (session *models.Session, err error) {
	if !utils.ValidateEmail(user.Email) ||
		!utils.ValidatePassword(user.Password) ||
		!utils.ValidateFirstName(user.FirstName) ||
		!utils.ValidateLastName(user.LastName) {
		return nil, fcerrors.New(fcerrors.InvalidUserData)
	}

	user.Email = utils.ConvertToCleanEmail(user.Email)

	existingUser, err := mgr.userRep.GetByEmail(user.Email)
	if err != nil && !repository.IsRecordNotFoundError(err) {
		// Don't bail out here, since this will be checked again when creating the user in repository
		log.Warn("Could not validate whether user with email %s already exists", user.Email)
	} else if err == nil && existingUser != nil && existingUser.ID > 0 {
		return nil, fcerrors.New(fcerrors.UserExists)
	}

	user.Password, err = crypt.HashScrypt(user.Password)
	if err != nil {
		log.Error(0, "Password hashing failed: %v", err)
		return nil, fcerrors.Wrap(err, fcerrors.HashingFailed)
	}
	user.IsAdmin = false

	// Save the user. This also fills their ID
	err = mgr.userRep.Create(user)
	if err != nil {
		log.Error(0, "Creating user failed: %v", err)
		return nil, fcerrors.Wrap(err, fcerrors.Database)
	}

	// If this is the first user (ID 1) they will become an admin
	if user.ID == 1 {
		log.Trace("Making first user an admin")
		user.IsAdmin = true
		err = mgr.userRep.Update(user)
		if err != nil {
			log.Error(0, "Could not make first user an admin: %v", err)
			// Since a system without an admin won't properly work, bail out
			return nil, fcerrors.Wrap(err, fcerrors.Database)
		}
	}

	err = GetFileManager().ScanUserFolderForChanges(user)
	if err != nil {
		log.Error(0, "Failed to scan folder for new user: %v", err)
		return nil, fcerrors.Wrap(err, fcerrors.Filesystem)
	}

	// Now, create a session for the user
	return mgr.createUserSession(user.ID)
}

// LoginUser verifies the user's credentials and then returns a new session.
func (mgr *AuthManager) LoginUser(email string, password string) (*models.Session, error) {
	// First, do some sanity checks so we can reduce calls to the credentials provider with obviously wrong data.
	if !utils.ValidateEmail(email) || !utils.ValidatePassword(password) {
		return nil, fcerrors.New(fcerrors.MissingCredentials)
	}

	email = utils.ConvertToCleanEmail(email)

	user, err := mgr.userRep.GetByEmail(email)
	if repository.IsRecordNotFoundError(err) {
		log.Warn("User not found by email %s", email)
		// we intentionally don't tell the user whether the error was due to bad credentials or the user being nonexistant
		return nil, fcerrors.New(fcerrors.BadCredentials)
	} else if err != nil {
		log.Error(0, "Could not get user via email %s: %v", email, err)
		return nil, fcerrors.Wrap(err, fcerrors.Database)
	}

	valid, err := crypt.ValidateScryptPassword(password, user.Password)
	if err != nil {
		log.Error(0, "Password verification failed for user %s: %v", user.Email, err)
		return nil, fcerrors.Wrap(err, fcerrors.HashingFailed)
	}
	if valid {
		return mgr.createUserSession(user.ID)
	}

	return &models.Session{}, fcerrors.New(fcerrors.BadCredentials)
}

// DeleteUser deletes a user from db and maybe his files
func (mgr *AuthManager) DeleteUser(userID int64) (err error) {
	user, err := mgr.userRep.GetByID(userID)
	if err != nil {
		log.Error(0, "Could not get user '%v' for deletion: %v", userID, err)
		return
	}

	if !user.RetainFilesAfterDeletion {
		err = GetFileManager().DeleteUserFiles(user)
		if err != nil {
			log.Error(0, "Failed to delete files for to be deleted user: %v", err)
			return
		}
	}

	if err = mgr.sessionRep.DeleteAllForUser(userID); err != nil {
		log.Error(0, "Could not delete all sessions for user %d: %v", userID, err)
		err = fcerrors.New(fcerrors.DeleteSession)
		return
	}

	if err = mgr.userRep.Delete(userID); err != nil {
		log.Error(0, "Deleting the user with ID %d failed: %v", userID, err)
		if repository.IsRecordNotFoundError(err) {
			err = fcerrors.New(fcerrors.UserNotFound)
		} else {
			err = fcerrors.New(fcerrors.Database)
		}
		return
	}

	return
}

// GetAllUsers returns all existing users with masked out password
func (mgr *AuthManager) GetAllUsers() ([]*models.User, error) {
	users, err := mgr.userRep.GetAll()
	if err != nil {
		log.Error(0, "Could not get all users, %v:", err)
		return nil, fcerrors.Wrap(err, fcerrors.Database)
	}
	return users, nil
}

// GetUserByID returns a user by ID
func (mgr *AuthManager) GetUserByID(userID int64) (*models.User, error) {
	user, err := mgr.userRep.GetByID(userID)
	if repository.IsRecordNotFoundError(err) {
		return nil, fcerrors.New(fcerrors.UserNotFound)
	} else if err != nil {
		return nil, fcerrors.Wrap(err, fcerrors.Database)
	}
	return user, nil
}

// GetUserByEmail returns a user by email
func (mgr *AuthManager) GetUserByEmail(email string) (*models.User, error) {
	email = utils.ConvertToCleanEmail(email)
	user, err := mgr.userRep.GetByEmail(email)
	if repository.IsRecordNotFoundError(err) {
		return nil, fcerrors.New(fcerrors.UserNotFound)
	} else if err != nil {
		return nil, fcerrors.Wrap(err, fcerrors.Database)
	}
	return user, nil
}

// GetAdminCount returns the count of admin users
func (mgr *AuthManager) GetAdminCount() (int, error) {
	count, err := mgr.userRep.AdminCount()
	return int(count), fcerrors.Wrap(err, fcerrors.Database)
}

func (mgr *AuthManager) createUserSession(userID int64) (*models.Session, error) {
	session := &models.Session{
		UserID:    userID,
		Token:     utils.RandomString(sessionTokenLength),
		ExpiresAt: time.Now().UTC().Add(time.Hour * time.Duration(mgr.sessionExpiry)).Unix(),
	}

	err := mgr.sessionRep.Create(session)
	if err != nil {
		log.Error(0, "Could not store session: %v", err)
		return nil, fcerrors.Wrap(err, fcerrors.Database)
	}

	err = mgr.userRep.UpdateLastSession(userID)
	if err != nil {
		log.Error(0, "Could not update last session of user: %v", err)
		return nil, fcerrors.Wrap(err, fcerrors.Database)
	}

	return session, nil
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

// DeleteSession removes the session from the session provider
func (mgr *AuthManager) DeleteSession(session *models.Session) (err error) {
	return fcerrors.Wrap(mgr.sessionRep.Delete(session), fcerrors.Database)
}

// GetSessionCount return the count of active sessions
func (mgr *AuthManager) GetSessionCount() (int, error) {
	count, err := mgr.sessionRep.Count()
	return count, fcerrors.Wrap(err, fcerrors.Database)
}
