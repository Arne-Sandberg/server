package repository

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/freecloudio/server/models"
)

func TestSessionRepository(t *testing.T) {
	notExpiring := time.Now().UTC().Unix() + 9999
	expiring := time.Now().UTC().Unix() - 9999
	session0 := &models.Session{UserID: 0, Token: "aabbccddeeff", ExpiresAt: notExpiring}
	session1 := &models.Session{UserID: 0, Token: "ffeeddccbbaa", ExpiresAt: notExpiring}
	session2 := &models.Session{UserID: 1, Token: "112233445566", ExpiresAt: notExpiring}
	session3 := &models.Session{UserID: 1, Token: "665544332211", ExpiresAt: expiring}
	dbName := "sessionTest.db"

	cleanDBFiles := func() {
		os.Remove(dbName)
	}

	cleanDBFiles()
	defer cleanDBFiles()

	var rep *SessionRepository

	success := t.Run("create connection and repository", func(t *testing.T) {
		err := InitDatabaseConnection("", "", "", "", 0, dbName)
		if err != nil {
			t.Fatalf("Failed to connect to gorm database: %v", err)
		}

		rep, err = CreateSessionRepository()
		if err != nil {
			t.Fatalf("Failed to create session repository: %v", err)
		}
	})
	if !success {
		t.Skip("Further test skipped due to setup failing")
	}

	t.Run("empty repository", func(t *testing.T) {
		count, err := rep.Count()
		if err != nil {
			t.Fatalf("Failed to get count: %v", err)
		}
		if count > 0 {
			t.Errorf("Count greater than zero for empty session repository: %d", count)
		}
	})

	success = t.Run("create sessions", func(t *testing.T) {
		err := rep.Create(session0)
		if err != nil {
			t.Errorf("Failed to create session: %v", err)
		}
		err = rep.Create(session1)
		if err != nil {
			t.Errorf("Failed to create session: %v", err)
		}
		err = rep.Create(session2)
		if err != nil {
			t.Errorf("Failed to create session: %v", err)
		}
		err = rep.Create(session3)
		if err != nil {
			t.Errorf("Failed to create session: %v", err)
		}
	})
	if !success {
		t.Skip("Skipping further tests due to no created sessions")
	}

	t.Run("correct count after creating sessions", func(t *testing.T) {
		count, err := rep.Count()
		if err != nil {
			t.Fatalf("Failed to get count: %v", err)
		}
		if count != 4 {
			t.Errorf("Count unequal to four for filled session repository: %d", err)
		}
	})

	t.Run("correct read back of created session", func(t *testing.T) {
		readBackSession, err := rep.GetByToken(session0.Token)
		if err != nil {
			t.Errorf("Failed to read back session0: %v", err)
		}
		if !reflect.DeepEqual(readBackSession, session0) {
			t.Error("Read back session0 and session0 not deeply equal")
		}
		readBackSession, err = rep.GetByToken(session1.Token)
		if err != nil {
			t.Errorf("Failed to read back session1: %v", err)
		}
		if !reflect.DeepEqual(readBackSession, session1) {
			t.Error("Read back session1 and session1 not deeply equal")
		}
	})

	delSuccess := t.Run("delete session", func(t *testing.T) {
		err := rep.Delete(session0)
		if err != nil {
			t.Errorf("Failed to delete session0: %v", err)
		}
	})

	if delSuccess {
		t.Run("correct read back after deleting session", func(t *testing.T) {
			_, err := rep.GetByToken(session0.Token)
			if err == nil || !IsRecordNotFoundError(err) {
				t.Errorf("Succeeded to read deleted session or error is not 'record not found': %v", err)
			}
		})
	}

	if delSuccess {
		t.Run("correct count after deleting session", func(t *testing.T) {
			count, err := rep.Count()
			if err != nil {
				t.Errorf("Failed to get count after session deletion: %v", err)
			}
			if count != 3 {
				t.Errorf("Count unequal to three after session deletion: %d", count)
			}
		})
	}

	userClearSuccess := t.Run("delete all sessions for user", func(t *testing.T) {
		err := rep.DeleteAllForUser(session1.UserID)
		if err != nil {
			t.Errorf("Failed to delete all sessions for user of session1: %v", err)
		}
	})

	if userClearSuccess {
		t.Run("correct read back after delete all sessions for user", func(t *testing.T) {
			_, err := rep.GetByToken(session1.Token)
			if err == nil || !IsRecordNotFoundError(err) {
				t.Errorf("Succeeded to read deleted session or error is not 'record not found': %v", err)
			}
		})
	}

	if userClearSuccess {
		t.Run("correct count after delete all sessions for user", func(t *testing.T) {
			count, err := rep.Count()
			if err != nil {
				t.Errorf("Failed to get count after delete all sessions for user: %v", err)
			}
			if count != 2 {
				t.Errorf("Count unequal to two after delete all session for user: %d", count)
			}
		})
	}

	delExpSuccess := t.Run("delete expired sessions", func(t *testing.T) {
		err := rep.DeleteExpired()
		if err != nil {
			t.Errorf("Failed to delete expired sessions: %v", err)
		}
	})

	if delExpSuccess {
		t.Run("correct read back after delete expired sessions", func(t *testing.T) {
			_, err := rep.GetByToken(session3.Token)
			if err == nil || !IsRecordNotFoundError(err) {
				t.Errorf("Succeeded to read expired session or error is not 'record not found': %v", err)
			}
		})
	}

	if delExpSuccess {
		t.Run("correct count after delete all sessions for user", func(t *testing.T) {
			count, err := rep.Count()
			if err != nil {
				t.Errorf("Failed to get count after delete all sessions for user: %v", err)
			}
			if count != 1 {
				t.Errorf("Count unequal to one after delete all session for user: %d", count)
			}
		})
	}
}
