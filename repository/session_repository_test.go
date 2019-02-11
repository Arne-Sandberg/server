package repository

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/freecloudio/server/models"
)

var testSessionSetupFailed = false
var testSessionDBName = "sessionTest.db"
var testSessionNotExpiring = time.Now().UTC().Unix() + 99999999
var testSessionExpiring = time.Now().UTC().Unix() - 99999999
var testSession0 = &models.Session{UserID: 0, Token: "aabbccddeeff", ExpiresAt: testSessionNotExpiring}
var testSession1 = &models.Session{UserID: 0, Token: "ffeeddccbbaa", ExpiresAt: testSessionNotExpiring}
var testSession2 = &models.Session{UserID: 1, Token: "112233445566", ExpiresAt: testSessionNotExpiring}
var testSession3 = &models.Session{UserID: 1, Token: "665544332211", ExpiresAt: testSessionExpiring}

func testSessionCleanup() {
	os.Remove(testSessionDBName)
}

func testSessionSetup() *SessionRepository {
	testSessionCleanup()
	InitDatabaseConnection("", "", "", "", 0, testSessionDBName)
	rep, _ := CreateSessionRepository()
	return rep
}

func testSessionInsert(rep *SessionRepository) {
	rep.Create(testSession0)
	rep.Create(testSession1)
	rep.Create(testSession2)
	rep.Create(testSession3)
}

func TestCreateSessionRepository(t *testing.T) {
	testSessionCleanup()
	defer testSessionCleanup()

	err := InitDatabaseConnection("", "", "", "", 0, testSessionDBName)
	if err != nil {
		t.Errorf("Failed to connect to gorm database: %v", err)
	}

	_, err = CreateSessionRepository()
	if err != nil {
		t.Errorf("Failed to create session repository: %v", err)
	}

	if t.Failed() {
		testSessionSetupFailed = true
	}
}

func TestCreateSession(t *testing.T) {
	if testSessionSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testSessionCleanup()
	rep := testSessionSetup()

	err := rep.Create(testSession0)
	if err != nil {
		t.Errorf("Failed to create session: %v", err)
	}

	if t.Failed() {
		testSessionSetupFailed = true
	}
}

func TestCountSessions(t *testing.T) {
	if testSessionSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testSessionCleanup()
	rep := testSessionSetup()

	count, err := rep.Count()
	if err != nil {
		t.Fatalf("Failed to get count: %v", err)
	}
	if count > 0 {
		t.Errorf("Count greater than zero for empty session repository: %d", count)
	}

	testSessionInsert(rep)

	count, err = rep.Count()
	if err != nil {
		t.Fatalf("Failed to get count: %v", err)
	}
	if count != 4 {
		t.Errorf("Count unequal to four for filled session repository: %d", err)
	}
}

func TestSessionGetByToken(t *testing.T) {
	if testSessionSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testSessionCleanup()
	rep := testSessionSetup()

	testSessionInsert(rep)

	readBackSession, err := rep.GetByToken(testSession0.Token)
	if err != nil {
		t.Errorf("Failed to read back session0: %v", err)
	}
	if !reflect.DeepEqual(readBackSession, testSession0) {
		t.Errorf("Read back session0 and session0 not deeply equal: %v != %v", readBackSession, testSession0)
	}
	readBackSession, err = rep.GetByToken(testSession1.Token)
	if err != nil {
		t.Errorf("Failed to read back session1: %v", err)
	}
	if !reflect.DeepEqual(readBackSession, testSession1) {
		t.Errorf("Read back session1 and session1 not deeply equal: %v != %v", readBackSession, testSession1)
	}
}

func TestDeleteSession(t *testing.T) {
	if testSessionSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testSessionCleanup()
	rep := testSessionSetup()

	testSessionInsert(rep)

	err := rep.Delete(testSession0)
	if err != nil {
		t.Errorf("Failed to delete session0: %v", err)
	}

	_, err = rep.GetByToken(testSession0.Token)
	if err == nil || !IsRecordNotFoundError(err) {
		t.Errorf("Succeeded to read deleted session or error is not 'record not found': %v", err)
	}

	count, err := rep.Count()
	if err != nil {
		t.Errorf("Failed to get count after session deletion: %v", err)
	}
	if count != 3 {
		t.Errorf("Count unequal to three after session deletion: %d", count)
	}
}

func TestDeleteAllForUserSessions(t *testing.T) {
	if testSessionSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testSessionCleanup()
	rep := testSessionSetup()

	testSessionInsert(rep)

	err := rep.DeleteAllForUser(testSession1.UserID)
	if err != nil {
		t.Errorf("Failed to delete all sessions for user of session1: %v", err)
	}

	_, err = rep.GetByToken(testSession1.Token)
	if err == nil || !IsRecordNotFoundError(err) {
		t.Errorf("Succeeded to read deleted session or error is not 'record not found': %v", err)
	}

	count, err := rep.Count()
	if err != nil {
		t.Errorf("Failed to get count after delete all sessions for user: %v", err)
	}
	if count != 2 {
		t.Errorf("Count unequal to two after delete all session for user: %d", count)
	}
}

func TestDeleteExpiredSessions(t *testing.T) {
	if testSessionSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testSessionCleanup()
	rep := testSessionSetup()

	testSessionInsert(rep)

	err := rep.DeleteExpired()
	if err != nil {
		t.Errorf("Failed to delete expired sessions: %v", err)
	}

	_, err = rep.GetByToken(testSession3.Token)
	if err == nil || !IsRecordNotFoundError(err) {
		t.Errorf("Succeeded to read expired session or error is not 'record not found': %v", err)
	}

	count, err := rep.Count()
	if err != nil {
		t.Errorf("Failed to get count after delete all sessions for user: %v", err)
	}
	if count != 3 {
		t.Errorf("Count unequal to three after delete all session for user: %d", count)
	}
}
