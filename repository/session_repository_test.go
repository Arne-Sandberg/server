package repository

import (
	"reflect"
	"testing"
	"time"

	"github.com/freecloudio/server/models"
)

var testSessionSetupFailed = false
var testSessionDBName = "sessionTest.db"
var testSessionNotExpiring = time.Now().UTC().Unix() + 99999999
var testSessionExpiring = time.Now().UTC().Unix() - 99999999
var testSession0User = &models.User{Username: "0"}
var testSession0 = &models.Session{Token: "aabbccddeeff", ExpiresAt: testSessionNotExpiring}
var testSession1User = &models.User{Username: "1"}
var testSession1 = &models.Session{Token: "ffeeddccbbaa", ExpiresAt: testSessionNotExpiring}
var testSession2 = &models.Session{Token: "112233445566", ExpiresAt: testSessionNotExpiring}
var testSession3 = &models.Session{Token: "665544332211", ExpiresAt: testSessionExpiring}

func testSessionSetup() *SessionRepository {
	testConnectClearGraph()

	userRep, _ := CreateUserRepository()
	userRep.Create(testSession0User)
	userRep.Create(testSession1User)

	rep, _ := CreateSessionRepository()
	return rep
}

func testSessionInsert(rep *SessionRepository) {
	rep.Create(testSession0, testSession0User.Username)
	rep.Create(testSession1, testSession0User.Username)
	rep.Create(testSession2, testSession1User.Username)
	rep.Create(testSession3, testSession1User.Username)
}

func TestCreateSessionRepository(t *testing.T) {
	defer testConnectClearGraph()
	testConnectClearGraph()

	_, err := CreateSessionRepository()
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
	defer testCloseClearGraph()
	rep := testSessionSetup()

	err := rep.Create(testSession0, testSession0User.Username)
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
	defer testCloseClearGraph()
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
	defer testCloseClearGraph()
	rep := testSessionSetup()

	testSessionInsert(rep)

	readBackSession, readBackUser, err := rep.GetWithUserByToken(testSession0.Token)
	if err != nil {
		t.Errorf("Failed to read back session0: %v", err)
	}
	if !reflect.DeepEqual(readBackSession, testSession0) {
		t.Errorf("Read back session0 and session0 not deeply equal: %v != %v", readBackSession, testSession0)
	}
	if !reflect.DeepEqual(readBackUser, testSession0User) {
		t.Errorf("Read back session0 user and session0 user are not deeply equal: %v != %v", readBackUser, testSession0User)
	}
	readBackSession, readBackUser, err = rep.GetWithUserByToken(testSession2.Token)
	if err != nil {
		t.Errorf("Failed to read back session2: %v", err)
	}
	if !reflect.DeepEqual(readBackSession, testSession2) {
		t.Errorf("Read back session2 and session2 not deeply equal: %v != %v", readBackSession, testSession2)
	}
	if !reflect.DeepEqual(readBackUser, testSession1User) {
		t.Errorf("Read back session2 user and session1 user are not deeply equal: %v != %v", readBackUser, testSession1User)
	}
}

func TestDeleteSession(t *testing.T) {
	if testSessionSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testCloseClearGraph()
	rep := testSessionSetup()

	testSessionInsert(rep)

	err := rep.Delete(testSession0)
	if err != nil {
		t.Errorf("Failed to delete session0: %v", err)
	}

	_, _, err = rep.GetWithUserByToken(testSession0.Token)
	if err == nil || err.Error() != "result contains no records" {
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
	defer testCloseClearGraph()
	rep := testSessionSetup()

	testSessionInsert(rep)

	err := rep.DeleteAllForUser(testSession0User.Username)
	if err != nil {
		t.Errorf("Failed to delete all sessions for user of session1: %v", err)
	}

	_, _, err = rep.GetWithUserByToken(testSession1.Token)
	if err == nil || err.Error() != "result contains no records" {
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
	defer testCloseClearGraph()
	rep := testSessionSetup()

	testSessionInsert(rep)

	err := rep.DeleteExpired()
	if err != nil {
		t.Errorf("Failed to delete expired sessions: %v", err)
	}

	_, _, err = rep.GetWithUserByToken(testSession3.Token)
	if err == nil || err.Error() != "result contains no records" {
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
