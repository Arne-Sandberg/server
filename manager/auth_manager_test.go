package manager

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/freecloudio/server/restapi/fcerrors"

	"github.com/freecloudio/server/models"

	"github.com/freecloudio/server/repository"
)

var testAuthSetupFailed = false
var testAuthDataFolder = "testData"
var testAuthDBName = "authTest.db"
var testAuthUserAdminPW = "12345678"
var testAuthUserAdmin = &models.User{FirstName: "Admin", LastName: "User", Email: "admin.user@email.com", IsAdmin: true, Password: testAuthUserAdminPW}
var testAuthUserPW = "87654321"
var testAuthUser = &models.User{FirstName: "User", LastName: "User", Email: "user.user@email.com", IsAdmin: false, Password: testAuthUserPW}

func testAuthCleanup(mgr *AuthManager) {
	if mgr != nil {
		mgr.Close()
	}
	authManager = nil
	os.Remove(testAuthDBName)
	os.RemoveAll(testAuthDataFolder)
	testAuthUserAdmin.Password = testAuthUserAdminPW
	testAuthUser.Password = testAuthUserPW
}

func testAuthReq() (sessionRep *repository.SessionRepository, userRep *repository.UserRepository) {
	testAuthCleanup(nil)
	repository.InitDatabaseConnection("", "", "", "", 0, testAuthDBName)
	sessionRep, _ = repository.CreateSessionRepository()
	userRep, _ = repository.CreateUserRepository()
	return
}

func testAuthSetup() *AuthManager {
	sessionRep, userRep := testAuthReq()
	mgr := CreateAuthManager(sessionRep, userRep)
	shareRep, _ := repository.CreateShareEntryRepository()
	fileInfoRep, _ := repository.CreateFileInfoRepository()
	fileSystemRep, _ := repository.CreateFileSystemRepository(testAuthDataFolder, ".tmp", 1, 1)
	CreateFileManager(fileSystemRep, fileInfoRep, shareRep, ".tmp")
	return mgr
}

func testAuthInsert(mgr *AuthManager) {
	mgr.CreateUser(testAuthUserAdmin)
	mgr.CreateUser(testAuthUser)
}

func TestCreateAuthManager(t *testing.T) {
	sessionRep, userRep := testAuthReq()

	mgr := CreateAuthManager(sessionRep, userRep)
	expMgr := &AuthManager{
		sessionRep: sessionRep,
		userRep:    userRep,
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

	mgr := CreateAuthManager(sessionRep, userRep)
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

	sess, err := mgr.CreateUser(testAuthUserAdmin)
	if err != nil {
		t.Errorf("Failed to create admin user: %v", err)
	}
	if sess.UserID != testAuthUserAdmin.ID {
		t.Errorf("Returned session for created admin user not for created user: %v != %v", sess.UserID, testAuthUserAdmin.ID)
	}
	sess, err = mgr.CreateUser(testAuthUser)
	if err != nil {
		t.Errorf("Failed to create user: %v", err)
	}
	if sess.UserID != testAuthUser.ID {
		t.Errorf("Returned session for created user not for created user: %v != %v", sess.UserID, testAuthUser.ID)
	}

	_, err = mgr.CreateUser(testAuthUser)
	if err == nil || err.(*fcerrors.FCError).Code != fcerrors.UserExists {
		t.Errorf("Creating already existing user succeeded or error is unequal to 'user exists': %v", err)
	}

	if t.Failed() {
		testAuthSetupFailed = true
	}
}

func TestNewSession(t *testing.T) {
	if testAuthSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	mgr := testAuthSetup()
	defer testAuthCleanup(mgr)

	testAuthInsert(mgr)

	sess, err := mgr.NewSession(testAuthUserAdmin.Email, testAuthUserAdminPW)
	if err != nil {
		t.Errorf("Failed to verify and get new session for admin user: %v", err)
	}
	if sess.UserID != testAuthUserAdmin.ID {
		t.Errorf("New verified session is not for correct user: %v != %v", sess.UserID, testAuthUserAdmin.ID)
	}

	_, err = mgr.NewSession(testAuthUser.Email, "wrongPassword")
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

	readBackUser, err := mgr.GetUserByID(testAuthUserAdmin.ID)
	if err != nil {
		t.Errorf("Failed to get admin user by ID: %v", err)
	}
	readBackUser.LastSessionAt = 0
	readBackUser.UpdatedAt = 0
	testAuthUserAdmin.UpdatedAt = 0
	if !reflect.DeepEqual(readBackUser, testAuthUserAdmin) {
		t.Errorf("Read back admin user by ID and admin user not deeply equal: %v != %v", readBackUser, testAuthUserAdmin)
	}

	_, err = mgr.GetUserByID(9999)
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
		if user.ID == testAuthUserAdmin.ID {
			testAuthUserAdmin.UpdatedAt = 0
			if !reflect.DeepEqual(user, testAuthUserAdmin) {
				t.Errorf("Read admin user from all users not deeply equal to admin user: %v != %v", user, testAuthUserAdmin)
			}
		} else if user.ID == testAuthUser.ID {
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
	sess, _ := mgr.NewSession(testAuthUser.Email, testAuthUserPW)

	err := mgr.DeleteUser(testAuthUser.ID)
	if err != nil {
		t.Errorf("Failed to delete user: %v", err)
	}
	_, err = mgr.GetUserByID(testAuthUser.ID)
	if err == nil || err.(*fcerrors.FCError).Code != fcerrors.UserNotFound {
		t.Errorf("Getting deleted user was successfull or error is unequal to 'user not found': %v", err)
	}
	_, err = mgr.NewSession(testAuthUser.Email, testAuthUserPW)
	if err == nil || err.(*fcerrors.FCError).Code != fcerrors.BadCredentials {
		t.Errorf("Creating new session for deleted user succeeded or error is unequal to 'bad credentials': %v", err)
	}
	valid := mgr.ValidateSession(sess)
	if valid {
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
	sess, _ := mgr.NewSession(testAuthUserAdmin.Email, testAuthUserAdminPW)

	res := mgr.ValidateSession(sess)
	if !res {
		t.Error("Failed to validate valid session")
	}

	sess.UserID = testAuthUser.ID
	res = mgr.ValidateSession(sess)
	if res {
		t.Error("Succeeded to validate session for wrong user")
	}

	sess.Token = "invalidToken"
	res = mgr.ValidateSession(sess)
	if res {
		t.Error("Succeeded to validate session with wrong token")
	}
}

func TestUpdateLastSession(t *testing.T) {
	if testAuthSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	mgr := testAuthSetup()
	defer testAuthCleanup(mgr)

	testAuthInsert(mgr)

	before, _ := mgr.GetUserByID(testAuthUserAdmin.ID)
	time.Sleep(3 * time.Second)
	mgr.UpdateLastSession(testAuthUserAdmin.ID)
	after, _ := mgr.GetUserByID(testAuthUserAdmin.ID)
	if before.LastSessionAt >= after.LastSessionAt {
		t.Errorf("Last session after update last session not greater than before: %v <= %v", before.LastSessionAt, after.LastSessionAt)
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
	mgr.NewSession(testAuthUser.Email, testAuthUserPW)
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
	sess, _ := mgr.NewSession(testAuthUser.Email, testAuthUserPW)
	count, _ = mgr.GetSessionCount()
	if count != 3 {
		t.Errorf("Session count unequal to three after new session: %d", count)
	}
	err := mgr.DeleteSession(sess)
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
	mgr.DeleteUser(testAuthUserAdmin.ID)
	count, err = mgr.GetAdminCount()
	if err != nil {
		t.Errorf("Failed to get admin count after deleting admin: %v", err)
	}
	if count != 0 {
		t.Errorf("Admin count unequal to one after deleting admin: %d", count)
	}
}
