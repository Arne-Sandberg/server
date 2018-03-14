package auth

import (
	"errors"
	"time"

	"github.com/freecloudio/freecloud/config"
	"github.com/freecloudio/freecloud/models"
	"github.com/freecloudio/freecloud/utils"

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
func Init(credentialsProvider CredentialsProvider, sessionProvider SessionProvider) {
	cProvider = credentialsProvider
	sProvider = sessionProvider

	done = make(chan struct{})
	go cleanupExpiredSessionsRoutine(time.Hour * time.Duration(config.GetInt("auth.session_expiry")))
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
func newUnverifiedSession(userID int) *models.Session {
	sess := &models.Session{
		UserID:    userID,
		Token:     utils.RandomString(SessionTokenLength),
		ExpiresAt: time.Now().UTC().Add(time.Hour * time.Duration(config.GetInt("auth.session_expiry"))),
	}
	err := sProvider.StoreSession(sess)
	if err != nil {
		log.Error(0, "Could not store session: %v", err)
	}

	updates := map[string]interface{}{
		"lastSession": time.Now().UTC(),
	}
	_, err = UpdateUser(userID, updates)
	if err != nil {
		log.Error(0, "Could not update user with lastSession %v", err)
	}

	return sess
}

func TotalSessionCount() int {
	return sProvider.TotalSessionCount()
}

// NewUser hashes the user's password, saves it to the database and then creates a new session, so he doesn't have to login again.
func NewUser(user *models.User) (session *models.Session, err error) {
	if !utils.ValidateEmail(user.Email) || !utils.ValidatePassword(user.Password) || !utils.ValidateFirstName(user.FirstName) || !utils.ValidateLastName(user.LastName) {
		err = ErrInvalidUserData
		return
	}

	existingUser, err := cProvider.GetUserByEmail(user.Email)
	if (*existingUser != models.User{}) {
		err = ErrUserAlreadyExists
		return
	}

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

	// If this is the first user (ID 1) he will become an admin
	if user.ID == 1 {
		log.Trace("Make first user an admin")
		user.IsAdmin = true
		err = cProvider.UpdateUser(user)
	}

	// Now, create a session for the user
	return newUnverifiedSession(user.ID), nil
}

func DeleteUser(userID int) (err error) {
	if err = sProvider.RemoveUserSessions(userID); err != nil {
		return
	}

	if err = cProvider.DeleteUser(userID); err != nil {
		return
	}
	return
}

func GetAllUsers() ([]*models.User, error) {
	users, err := cProvider.GetAllUsers()
	if err != nil {
		log.Error(0, "Could not get all users, %v:", err)
		return nil, err
	}
	for i := 0; i < len(users); i++ {
		// Mask out the password
		users[i].Password = ""
	}
	return users, nil
}

// ValidateSession checks if the session is valid.
func ValidateSession(sess *models.Session) (valid bool) {
	return sProvider.SessionIsValid(sess)
}

func GetUserByID(userID int) (*models.User, error) {
	return cProvider.GetUserByID(userID)
}

//RemoveSession removes the session from the session provider
func RemoveSession(sess *models.Session) (err error) {
	return sProvider.RemoveSession(sess)
}

func UpdateUser(userID int, updates map[string]interface{}) (user *models.User, err error) {
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
	if avatarURL, ok := updates["avatarURL"]; ok == true {
		user.AvatarURL, ok = avatarURL.(string)
		if ok != true || !utils.ValidateAvatarURL(user.AvatarURL) {
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
		user.LastSession, ok = lastSession.(time.Time)
		if ok != true {
			err = ErrInvalidUserData
			return
		}
	} else {
		// I expect that the lastSession will only be updated if nothing else of the data is updated.
		// That way the "Updated" only represents changes to the core user data
		user.Updated = time.Now().UTC()
	}

	err = cProvider.UpdateUser(user)
	user.Password = ""

	return
}

func GetAdminCount() (int, error) {
	return cProvider.GetAdminCount()
}
