package auth

import (
	"errors"
	"time"

	"github.com/freecloudio/freecloud/models"
	"github.com/freecloudio/freecloud/utils"

	"github.com/jinzhu/gorm"
	log "gopkg.in/clog.v1"
)

const SessionTokenLength = 32

var (
	authInst              *auth
	ErrMissingCredentials = errors.New("auth: Missing credentials")
	ErrInvalidCredentials = errors.New("auth: Invalid credentials")
	ErrInvalidUserData    = errors.New("auth: Invalid user data")
	ErrUserAlreadyExists  = errors.New("auth: User already exists")
)

type auth struct {
	cProvider     CredentialsProvider
	sProvider     SessionProvider
	sessionExpiry time.Duration
	done          chan struct{}
}

// Init intializes the auth package. You must call this before using any auth function.
func Init(credentialsProvider CredentialsProvider, sessionProvider SessionProvider, sessionExp int) {
	authInst = &auth{
		cProvider:     credentialsProvider,
		sProvider:     sessionProvider,
		sessionExpiry: time.Hour * time.Duration(sessionExp),
		done:          make(chan struct{}),
	}

	go cleanupExpiredSessionsRoutine(authInst.sessionExpiry)
}

func Close() {
	authInst.done <- struct{}{}
}

func cleanupExpiredSessionsRoutine(interval time.Duration) {
	log.Trace("Starting old session cleaner, running now and every %v", interval)
	authInst.sProvider.CleanupExpiredSessions()

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-authInst.done:
			return
		case <-ticker.C:
			authInst.sProvider.CleanupExpiredSessions()
		}
	}
}

// NewSession verifies the user's credentials and then returns a new Session
func NewSession(email string, password string) (*models.Session, error) {
	// First, do some sanity checks before verification
	if len(password) == 0 {
		return &models.Session{}, ErrMissingCredentials
	}
	// Get the user
	user, err := authInst.cProvider.GetUserByEmail(email)
	if err != nil {
		log.Error(0, "Could not get user with email %s: %v", email, err)
		return &models.Session{}, err
	}
	// Now, verify the password
	valid, err := ValidatePassword(password, user.Password)
	if err != nil {
		log.Error(0, "Password verification failed: %v", err)
		return &models.Session{}, err
	}
	if valid {
		return newUnverifiedSession(user.ID)
	}
	return &models.Session{}, ErrInvalidCredentials
}

// newUnverifiedSession issues a session token but does not verify the user's password
func newUnverifiedSession(userID int64) (sess *models.Session, err error) {
	sess = &models.Session{
		UserID:    userID,
		Token:     utils.RandomString(SessionTokenLength),
		ExpiresAt: time.Now().UTC().Add(authInst.sessionExpiry).Unix(),
	}
	err = authInst.sProvider.StoreSession(sess)
	if err != nil {
		log.Error(0, "Could not store session: %v", err)
		return
	}

	err = UpdateLastSession(userID)
	if err != nil {
		log.Error(0, "Could not update user with lastSession %v", err)
		return
	}

	return
}

func TotalSessionCount() int64 {
	count, _ := authInst.sProvider.TotalSessionCount()
	return count
}

// NewUser hashes the user's password, saves it to the database and then creates a new session, so he doesn't have to login again.
func NewUser(user *models.User) (session *models.Session, err error) {
	if !utils.ValidateEmail(user.Email) || !utils.ValidatePassword(user.Password) || !utils.ValidateFirstName(user.FirstName) || !utils.ValidateLastName(user.LastName) {
		err = ErrInvalidUserData
		return
	}

	_, err = authInst.cProvider.GetUserByEmail(user.Email)
	if !gorm.IsRecordNotFoundError(err) {
		err = ErrUserAlreadyExists
		return
	}

	user.Password, err = HashPassword(user.Password)
	if err != nil {
		log.Error(0, "Password hashing failed: %v", err)
		return
	}

	// Save the user. This also fills its ID
	err = authInst.cProvider.CreateUser(user)
	if err != nil {
		return
	}

	// If this is the first user (ID 1) he will become an admin
	if user.ID == 1 {
		log.Trace("Make first user an admin")
		user.IsAdmin = true
		err = authInst.cProvider.UpdateUser(user)
	}

	// Now, create a session for the user
	return newUnverifiedSession(user.ID)
}

func DeleteUser(userID int64) (err error) {
	if err = authInst.sProvider.RemoveUserSessions(userID); err != nil {
		return
	}

	if err = authInst.cProvider.DeleteUser(userID); err != nil {
		return
	}
	return
}

func GetAllUsers(isAdmin bool) ([]*models.User, error) {
	users, err := authInst.cProvider.GetAllUsers()
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
func ValidateSession(sess *models.Session) (valid bool) {
	return authInst.sProvider.SessionIsValid(sess)
}

func GetUserByID(userID int64) (*models.User, error) {
	return authInst.cProvider.GetUserByID(userID)
}

func GetUserByEmail(email string) (*models.User, error) {
	return authInst.cProvider.GetUserByEmail(email)
}

//RemoveSession removes the session from the session provider
func RemoveSession(sess *models.Session) (err error) {
	return authInst.sProvider.RemoveSession(sess)
}

func UpdateLastSession(userID int64) (err error) {
	user, err := GetUserByID(userID)
	if err != nil {
		return
	}

	user.LastSessionAt = utils.GetTimestampNow()
	err = authInst.cProvider.UpdateUser(user)

	return
}

func UpdateUser(userID int64, updatedUser *models.UserUpdate) (user *models.User, err error) {
	/*
		user, err = GetUserByID(userID)
		if err != nil {
			return
		}

		if email, ok := updatedUser.EmailOO.(*models.UserUpdate_Email); ok == true {
			user.Email = email.Email
			if !utils.ValidateEmail(user.Email) {
				err = ErrInvalidUserData
				return
			}
		}

		if firstName, ok := updatedUser.FirstNameOO.(*models.UserUpdate_FirstName); ok == true {
			user.FirstName = firstName.FirstName
			if !utils.ValidateFirstName(user.FirstName) {
				err = ErrInvalidUserData
				return
			}
		}

		if lastName, ok := updatedUser.LastNameOO.(*models.UserUpdate_LastName); ok == true {
			user.LastName = lastName.LastName
			if !utils.ValidateLastName(user.LastName) {
				err = ErrInvalidUserData
				return
			}
		}

		if isAdmin, ok := updatedUser.IsAdminOO.(*models.UserUpdate_IsAdmin); ok == true {
			user.IsAdmin = isAdmin.IsAdmin
		}

		if password, ok := updatedUser.PasswordOO.(*models.UserUpdate_Password); ok == true {
			if ok != true || !utils.ValidatePassword(password.Password) {
				err = ErrInvalidUserData
				return
			}
			user.Password, err = HashPassword(password.Password)
			if err != nil {
				err = ErrInvalidUserData
				return
			}
		}

		user.UpdatedAt = utils.GetTimestampNow()

		err = cProvider.UpdateUser(user)
		user.Password = ""
	*/
	return
}

func GetAdminCount() (int, error) {
	return authInst.cProvider.GetAdminCount()
}
