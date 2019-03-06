package manager

import (
	"reflect"
	"testing"
	"time"

	"github.com/freecloudio/server/repository"
)

var testSystemVersion = "v0.0.0"
var testSystemDBName = "systemTest.db"

func testSystemCleanup() {
	systemManager = nil
	testCloseClearGraph()
}

func testSystemReq() (sessionRep *repository.SessionRepository, fileInfoRep *repository.FileInfoRepository) {
	testConnectClearGraph()
	sessionRep, _ = repository.CreateSessionRepository()
	fileInfoRep, _ = repository.CreateFileInfoRepository()
	return
}

func TestCreateSystemManager(t *testing.T) {
	defer testSystemCleanup()
	sessionRep, fileInfoRep := testSystemReq()

	mgr := CreateSystemManager(testSystemVersion, sessionRep, fileInfoRep)
	expMgr := &SystemManager{
		version:     testSystemVersion,
		sessionRep:  sessionRep,
		fileInfoRep: fileInfoRep,
	}
	mgr.startTime = time.Time{}

	if !reflect.DeepEqual(mgr, expMgr) {
		t.Errorf("Created systemManager and expected systemManager not deeply equal: %v != %v", mgr, expMgr)
	}
}

func TestGetSystemManager(t *testing.T) {
	defer testSystemCleanup()
	sessionRep, fileInfoRep := testSystemReq()

	mgr := CreateSystemManager(testSystemVersion, sessionRep, fileInfoRep)
	mgrGet := GetSystemManager()

	if !reflect.DeepEqual(mgr, mgrGet) {
		t.Errorf("Created and read system manager are not deeply equal: %v != %v", mgr, mgrGet)
	}
}

func TestGetSystemStats(t *testing.T) {
	defer testSystemCleanup()
	sessionRep, fileInfoRep := testSystemReq()

	mgr := CreateSystemManager(testSystemVersion, sessionRep, fileInfoRep)
	_, err := mgr.GetSystemStats()
	if err != nil {
		t.Fatalf("Failed to get system stats: %#v", err)
	}
}
