package manager

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/freecloudio/server/repository"
)

var testSystemVersion = "v0.0.0"
var testSystemDBName = "systemTest.db"

func testSystemCleanup() {
	systemManager = nil
	authManager = nil
	os.Remove(testSystemDBName)
}

func testSystemReq() {
	repository.InitDatabaseConnection("", "", "", "", 0, testSystemDBName)
	sessionRep, _ := repository.CreateSessionRepository()
	userRep, _ := repository.CreateUserRepository()

	CreateAuthManager(sessionRep, userRep)
}

func TestCreateSystemManager(t *testing.T) {
	defer testSystemCleanup()

	mgr := CreateSystemManager(testSystemVersion)
	expMgr := &SystemManager{
		version: testSystemVersion,
	}
	mgr.startTime = time.Time{}

	if !reflect.DeepEqual(mgr, expMgr) {
		t.Errorf("Created systemManager and expected systemManager not deeply equal: %v != %v", mgr, expMgr)
	}
}

func TestGetSystemManager(t *testing.T) {
	defer testSystemCleanup()

	mgr := CreateSystemManager(testSystemVersion)
	mgrGet := GetSystemManager()

	if !reflect.DeepEqual(mgr, mgrGet) {
		t.Errorf("Created and read system manager are not deeply equal: %v != %v", mgr, mgrGet)
	}
}

func TestGetSystemStats(t *testing.T) {
	defer testSystemCleanup()
	testSystemReq()

	mgr := CreateSystemManager(testSystemVersion)
	_, err := mgr.GetSystemStats()
	if err != nil {
		t.Fatalf("Failed to get system stats: %#v", err)
	}
}
