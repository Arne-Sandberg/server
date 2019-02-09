package manager

import (
	"os"
	"reflect"
	"testing"

	"github.com/freecloudio/server/repository"
)

var testAuthDBName = "authTest.db"

func testAuthSetup() (sessionRep *repository.SessionRepository, userRep *repository.UserRepository) {
	repository.InitDatabaseConnection("", "", "", "", 0, testAuthDBName)
	sessionRep, _ = repository.CreateSessionRepository()
	userRep, _ = repository.CreateUserRepository()
	return
}

func testAuthCleanup() {
	if authManager != nil {
		authManager.Close()
	}
	authManager = nil
	os.Remove(testAuthDBName)
}

func TestCreateAuthManager(t *testing.T) {
	defer testAuthCleanup()
	sessionRep, userRep := testAuthSetup()

	mgr := CreateAuthManager(sessionRep, userRep)
	expMgr := &AuthManager{
		sessionRep: sessionRep,
		userRep:    userRep,
	}
	mgr.done = nil

	if !reflect.DeepEqual(mgr, expMgr) {
		t.Errorf("Created authManager and expected authManager not deeply equal: %v != %v", mgr, expMgr)
	}
}

func TestGetAuthManager(t *testing.T) {
	defer testAuthCleanup()
	sessionRep, userRep := testAuthSetup()

	mgr := CreateAuthManager(sessionRep, userRep)
	mgrGet := GetAuthManager()

	if !reflect.DeepEqual(mgr, mgrGet) {
		t.Errorf("Created and read system manager are not deeply equal: %v != %v", mgr, mgrGet)
	}
}
