package auth

import (
	"errors"
	"time"

	"github.com/freecloudio/freecloud/config"
	"github.com/freecloudio/freecloud/models"
	"github.com/freecloudio/freecloud/utils"

	"github.com/golang/protobuf/ptypes/timestamp"
	log "gopkg.in/clog.v1"
)

const SessionTokenLength = 32

var (
	cProvider CredentialsProvider
	sProvider SessionProvider
	done      chan struct{}

	ErrMissingCredentials = errors.New("auth: Missing credentials")
	ErrInvalidCredentials = errors.New("auth: Invalid credentials")
	ErrInvalidUserData    = errors.New("auth: Invalid user data")
	ErrUserAlreadyExists  = errors.New("auth: User already exists")
)

// Init intializes the auth package. You must call this before using any auth function.
func Init(credentialsProvider CredentialsProvider, sessionProvider SessionProvider, sessionExpiry int) {
	cProvider = credentialsProvider
	sProvider = sessionProvider

	done = make(chan struct{})
	go cleanupExpiredSessionsRoutine(time.Hour * time.Duration(sessionExpiry))
}

func Close() {
	done <- struct{}{}
}

func cleanupExpiredSessionsRoutine(interval time.Duration) {
	log.Trace("Starting old session cleaner, running now and every %v", interval)
	sProvider.CleanupExpiredSessions()

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			sProvider.CleanupExpiredSessions()
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
	user, err := cProvider.GetUserByEmail(email)
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
		return newUnverifiedSession(user.ID), nil
	}
	return &models.Session{}, ErrInvalidCredentials
}

// newUnverifiedSession issues a session token but does not verify the user's password
func newUnverifiedSession(userID uint32) *models.Session {
	sess := &models.Session{
		UserID:    userID,
		Token:     utils.RandomString(SessionTokenLength),
		ExpiresAt: utils.GetTimestampFromTime(time.Now().UTC().Add(time.Hour * time.Duration(config.GetInt("auth.session_expiry")))),
	}
	err := sProvider.StoreSession(sess)
	if err != nil {
		log.Error(0, "Could not store session: %v", err)
	}

	updates := map[string]interface{}{
		"lastSession": utils.GetTimestampNow(),
	}
	_, err = UpdateUser(userID, updates)
	if err != nil {
		log.Error(0, "Could not update user with lastSession %v", err)
	}

	return sess
}

func TotalSessionCount() uint32 {
	return sProvider.TotalSessionCount()
}

// NewUser hashes the user's password, saves it to the database and then creates a new session, so he doesn't have to login again.
func NewUser(user *models.User) (session *models.Session, err error) {
	if !utils.ValidateEmail(user.Email) || !utils.ValidatePassword(user.Password) || !utils.ValidateFirstName(user.FirstName) || !utils.ValidateLastName(user.LastName) {
		err = ErrInvalidUserData
		return
	}

	existingUser, err := cProvider.GetUserByEmail(user.Email)
	if existingUser.Email == user.Email {
		err = ErrUserAlreadyExists
		return
	}

	user.CreatedAt = utils.GetTimestampNow()
	user.UpdatedAt = utils.GetTimestampNow()
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

	// If this is the first user (ID 1) he will become an admin
	if user.ID == 1 {
		log.Trace("Make first user an admin")
		user.IsAdmin = true
		err = cProvider.UpdateUser(user)
	}

	// Now, create a session for the user
	return newUnverifiedSession(user.ID), nil
}

func DeleteUser(userID uint32) (err error) {
	if err = sProvider.RemoveUserSessions(userID); err != nil {
		return
	}

	if err = cProvider.DeleteUser(userID); err != nil {
		return
	}
	return
}

func GetAllUsers(isAdmin bool) ([]*models.User, error) {
	users, err := cProvider.GetAllUsers()
	if err != nil {
		log.Error(0, "Could not get all users, %v:", err)
		return nil, err
	}
	for _, user := range users {
		// Mask out the password
		user.Password = ""

		// For normal users also mask out created, updated and lastSession
		if !isAdmin {
			user.CreatedAt = &timestamp.Timestamp{}
			user.UpdatedAt = &timestamp.Timestamp{}
			user.LastSessionAt = &timestamp.Timestamp{}
		}
	}
	return users, nil
}

// ValidateSession checks if the session is valid.
func ValidateSession(sess *models.Session) (valid bool) {
	return sProvider.SessionIsValid(sess)
}

func GetUserByID(userID uint32) (*models.User, error) {
	return cProvider.GetUserByID(userID)
}

//RemoveSession removes the session from the session provider
func RemoveSession(sess *models.Session) (err error) {
	return sProvider.RemoveSession(sess)
}

func UpdateUser(userID uint32, updates map[string]interface{}) (user *models.User, err error) {
	user, err = GetUserByID(userID)
	if err != nil {
		return
	}

	if email, ok := updates["email"]; ok == true {
		user.Email, ok = email.(string)
		if ok != true || !utils.ValidateEmail(user.Email) {
			err = ErrInvalidUserData
			return
		}
	}
	if firstName, ok := updates["firstName"]; ok == true {
		user.FirstName, ok = firstName.(string)
		if ok != true || !utils.ValidateFirstName(user.FirstName) {
			err = ErrInvalidUserData
			return
		}
	}
	if lastName, ok := updates["lastName"]; ok == true {
		user.LastName, ok = lastName.(string)
		if ok != true || !utils.ValidateLastName(user.LastName) {
			err = ErrInvalidUserData
			return
		}
	}
	if isAdmin, ok := updates["isAdmin"]; ok == true {
		user.IsAdmin, ok = isAdmin.(bool)
		if ok != true {
			err = ErrInvalidUserData
			return
		}
	}
	if password, ok := updates["password"]; ok == true {
		newPassword, ok := password.(string)
		if ok != true || !utils.ValidatePassword(user.Password) {
			err = ErrInvalidUserData
			return
		}
		user.Password, err = HashPassword(newPassword)
		if err != nil {
			err = ErrInvalidUserData
			return
		}
	}
	if lastSession, ok := updates["lastSession"]; ok == true {
		user.LastSessionAt, ok = lastSession.(*timestamp.Timestamp)
		if ok != true {
			err = ErrInvalidUserData
			return
		}
	} else {
		// I expect that the lastSession will only be updated if nothing else of the data is updated.
		// That way the "Updated" only represents changes to the core user data
		user.UpdatedAt = utils.GetTimestampNow()
	}

	err = cProvider.UpdateUser(user)
	user.Password = ""

	return
}

func GetAdminCount() (int, error) {
	return cProvider.GetAdminCount()
}
