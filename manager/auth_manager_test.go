package manager

import (
	"os"
	"reflect"
	"testing"

	"github.com/freecloudio/server/models"
	"github.com/freecloudio/server/repository"
	"github.com/freecloudio/server/restapi/fcerrors"
)

var testAuthSetupFailed = false
var testAuthDataFolder = "testData"
var testAuthUserAdminPW = "12345678"
var testAuthUserAdmin = &models.User{Username: "Admin", FirstName: "Admin", LastName: "User", Email: "admin.user@email.com", IsAdmin: true, Password: testAuthUserAdminPW}
var testAuthUserPW = "87654321"
var testAuthUser = &models.User{Username: "User", FirstName: "User", LastName: "User", Email: "user.user@email.com", IsAdmin: false, Password: testAuthUserPW}

func testAuthCleanup(mgr *AuthManager) {
	// If nil then it runs before req
	if mgr != nil {
		mgr.Close()
		testCloseClearGraph()
	}
	authManager = nil
	os.RemoveAll(testAuthDataFolder)
	testAuthUserAdmin.Password = testAuthUserAdminPW
	testAuthUser.Password = testAuthUserPW
}

func testAuthReq() (sessionRep *repository.SessionRepository, userRep *repository.UserRepository) {
	testAuthCleanup(nil)
	testConnectClearGraph()
	sessionRep, _ = repository.CreateSessionRepository()
	userRep, _ = repository.CreateUserRepository()
	return
}

func testAuthSetup() *AuthManager {
	sessionRep, userRep := testAuthReq()
	mgr := CreateAuthManager(sessionRep, userRep, 24, 1)
	return mgr
}

func testAuthInsert(mgr *AuthManager) {
	mgr.CreateUser(testAuthUserAdmin)
	mgr.CreateUser(testAuthUser)
}

func TestCreateAuthManager(t *testing.T) {
	sessionRep, userRep := testAuthReq()

	mgr := CreateAuthManager(sessionRep, userRep, 24, 1)
	expMgr := &AuthManager{
		sessionRep:             sessionRep,
		userRep:                userRep,
		sessionExpiry:          24,
		sessionCleanupInterval: 1,
	}
	mgr.Close()
	mgr.done = nil

	if !reflect.DeepEqual(mgr, expMgr) {
		t.Errorf("Created authManager and expected authManager not deeply equal: %v != %v", mgr, expMgr)
	}

	if t.Failed() {
		testAuthSetupFailed = true
	}

	testAuthCleanup(nil)
}

func TestGetAuthManager(t *testing.T) {
	if testAuthSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	sessionRep, userRep := testAuthReq()

	mgr := CreateAuthManager(sessionRep, userRep, 24, 1)
	mgrGet := GetAuthManager()

	if !reflect.DeepEqual(mgr, mgrGet) {
		t.Errorf("Created and read system manager are not deeply equal: %v != %v", mgr, mgrGet)
	}

	testAuthCleanup(mgr)
}

func TestCreateUser(t *testing.T) {
	if testAuthSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	mgr := testAuthSetup()
	defer testAuthCleanup(mgr)

	token, err := mgr.CreateUser(testAuthUserAdmin)
	if err != nil {
		t.Errorf("Failed to create admin user: %v", err)
	}
	if token.Token == "" {
		t.Error("Token empty for new admin user")
	}
	token, err = mgr.CreateUser(testAuthUser)
	if err != nil {
		t.Errorf("Failed to create user: %v", err)
	}
	if token.Token == "" {
		t.Error("Token empty for new user")
	}

	_, err = mgr.CreateUser(testAuthUser)
	if err == nil || err.(*fcerrors.FCError).Code != fcerrors.UserExists {
		t.Errorf("Creating already existing user succeeded or error is unequal to 'user exists': %v", err)
	}

	if t.Failed() {
		testAuthSetupFailed = true
	}
}

func TestUserLogin(t *testing.T) {
	if testAuthSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	mgr := testAuthSetup()
	defer testAuthCleanup(mgr)

	testAuthInsert(mgr)

	token, err := mgr.LoginUser(testAuthUserAdmin.Email, testAuthUserAdminPW)
	if err != nil {
		t.Errorf("Failed to verify and get new session for admin user: %v", err)
	}
	if token.Token == "" {
		t.Error("Token empty for logged in admin user")
	}

	_, err = mgr.LoginUser(testAuthUser.Email, "wrongPassword")
	if err == nil || err.(*fcerrors.FCError).Code != fcerrors.BadCredentials {
		t.Errorf("Verifying and creating session with wrong user credentials succeeded or error is not 'bad credentials': %v", err)
	}
}

func TestGetUserByID(t *testing.T) {
	if testAuthSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	mgr := testAuthSetup()
	defer testAuthCleanup(mgr)

	testAuthInsert(mgr)

	readBackUser, err := mgr.GetUserByUsername(testAuthUserAdmin.Username)
	if err != nil {
		t.Errorf("Failed to get admin user by username: %v", err)
	}
	readBackUser.LastSessionAt = 0
	readBackUser.UpdatedAt = 0
	testAuthUserAdmin.UpdatedAt = 0
	if !reflect.DeepEqual(readBackUser, testAuthUserAdmin) {
		t.Errorf("Read back admin user by ID and admin user not deeply equal: %v != %v", readBackUser, testAuthUserAdmin)
	}

	_, err = mgr.GetUserByUsername("NonExisting")
	if err == nil || err.(*fcerrors.FCError).Code != fcerrors.UserNotFound {
		t.Errorf("Getting user with non existing id succeeded or error is not 'user not found': %v", err)
	}
}

func TestGetUserByMail(t *testing.T) {
	if testAuthSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	mgr := testAuthSetup()
	defer testAuthCleanup(mgr)

	testAuthInsert(mgr)

	readBackUser, err := mgr.GetUserByEmail(testAuthUser.Email)
	if err != nil {
		t.Errorf("Failed to get user by email: %v", err)
	}
	readBackUser.LastSessionAt = 0
	readBackUser.UpdatedAt = 0
	testAuthUser.UpdatedAt = 0
	if !reflect.DeepEqual(readBackUser, testAuthUser) {
		t.Errorf("Read back user by email and user not deeply equal: %v != %v", readBackUser, testAuthUser)
	}

	_, err = mgr.GetUserByEmail("not@existing.com")
	if err == nil || err.(*fcerrors.FCError).Code != fcerrors.UserNotFound {
		t.Errorf("Getting user with non existing email succeeded or error is not 'user not found': %v", err)
	}
}

func TestGetAllUsers(t *testing.T) {
	if testAuthSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	mgr := testAuthSetup()
	defer testAuthCleanup(mgr)

	testAuthInsert(mgr)

	users, err := mgr.GetAllUsers()
	if err != nil {
		t.Errorf("Failed to get all users: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("Lenght of all users unequal to two: %d", len(users))
	}
	for _, user := range users {
		user.LastSessionAt = 0
		user.UpdatedAt = 0
		if user.Username == testAuthUserAdmin.Username {
			testAuthUserAdmin.UpdatedAt = 0
			if !reflect.DeepEqual(user, testAuthUserAdmin) {
				t.Errorf("Read admin user from all users not deeply equal to admin user: %v != %v", user, testAuthUserAdmin)
			}
		} else if user.Username == testAuthUser.Username {
			testAuthUser.UpdatedAt = 0
			if !reflect.DeepEqual(user, testAuthUser) {
				t.Errorf("Read user from all user not deeply equal to user: %v != %v", user, testAuthUser)
			}
		}
	}
}

func TestDeleteUser(t *testing.T) {
	if testAuthSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	mgr := testAuthSetup()
	defer testAuthCleanup(mgr)

	testAuthInsert(mgr)
	token, _ := mgr.LoginUser(testAuthUser.Email, testAuthUserPW)

	err := mgr.DeleteUser(testAuthUser.Username)
	if err != nil {
		t.Errorf("Failed to delete user: %v", err)
	}
	_, err = mgr.GetUserByUsername(testAuthUser.Username)
	if err == nil || err.(*fcerrors.FCError).Code != fcerrors.UserNotFound {
		t.Errorf("Getting deleted user was successfull or error is unequal to 'user not found': %v", err)
	}
	_, err = mgr.LoginUser(testAuthUser.Email, testAuthUserPW)
	if err == nil || err.(*fcerrors.FCError).Code != fcerrors.BadCredentials {
		t.Errorf("Creating new session for deleted user succeeded or error is unequal to 'bad credentials': %v", err)
	}
	_, err = mgr.ValidateToken(token)
	if err == nil {
		t.Error("Session valid for deleted user")
	}
}

func TestValidateSession(t *testing.T) {
	if testAuthSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	mgr := testAuthSetup()
	defer testAuthCleanup(mgr)

	testAuthInsert(mgr)
	token, _ := mgr.LoginUser(testAuthUserAdmin.Email, testAuthUserAdminPW)

	user, err := mgr.ValidateToken(token)
	if err != nil {
		t.Errorf("Failed to validate valid token: %v", err)
	}
	user.LastSessionAt = 0
	user.UpdatedAt = 0
	testAuthUserAdmin.UpdatedAt = 0
	if !reflect.DeepEqual(user, testAuthUserAdmin) {
		t.Errorf("Read back admin user through token validation and admin user not deeply equal: %v != %v", user, testAuthUserAdmin)
	}

	token.Token = "invalidToken"
	_, err = mgr.ValidateToken(token)
	if err == nil {
		t.Error("Succeeded to validate wrong token")
	}
}

func TestGetSessionCount(t *testing.T) {
	if testAuthSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	mgr := testAuthSetup()
	defer testAuthCleanup(mgr)

	testAuthInsert(mgr)

	count, err := mgr.GetSessionCount()
	if err != nil {
		t.Errorf("Failed to get session count: %v", err)
	}
	if count != 2 {
		t.Errorf("Session count unequal to two: %d", count)
	}
	mgr.LoginUser(testAuthUser.Email, testAuthUserPW)
	count, err = mgr.GetSessionCount()
	if err != nil {
		t.Errorf("Failed to get session count after new session: %v", err)
	}
	if count != 3 {
		t.Errorf("Session count unequal to three after new session: %d", count)
	}
}

func TestDeleteSession(t *testing.T) {
	if testAuthSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	mgr := testAuthSetup()
	defer testAuthCleanup(mgr)

	testAuthInsert(mgr)

	count, _ := mgr.GetSessionCount()
	if count != 2 {
		t.Errorf("Session count unequal to two: %d", count)
	}
	token, _ := mgr.LoginUser(testAuthUser.Email, testAuthUserPW)
	count, _ = mgr.GetSessionCount()
	if count != 3 {
		t.Errorf("Session count unequal to three after new session: %d", count)
	}
	err := mgr.DeleteToken(token)
	if err != nil {
		t.Errorf("Failed to delete session: %v", err)
	}
	count, _ = mgr.GetSessionCount()
	if count != 2 {
		t.Errorf("Session count unequal to two after deleting session: %d", count)
	}
}

func TestGetAdminCount(t *testing.T) {
	if testAuthSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	mgr := testAuthSetup()
	defer testAuthCleanup(mgr)

	testAuthInsert(mgr)

	count, err := mgr.GetAdminCount()
	if err != nil {
		t.Errorf("Failed to get admin count: %v", err)
	}
	if count != 1 {
		t.Errorf("Admin count unequal to one: %d", count)
	}
	mgr.DeleteUser(testAuthUserAdmin.Username)
	count, err = mgr.GetAdminCount()
	if err != nil {
		t.Errorf("Failed to get admin count after deleting admin: %v", err)
	}
	if count != 0 {
		t.Errorf("Admin count unequal to one after deleting admin: %d", count)
	}
}
