package repository

import (
	"time"

	"github.com/freecloudio/server/models"
	log "gopkg.in/clog.v1"
)

// Add used models to enable auto migration for them
func init() {
	databaseModels = append(databaseModels, &models.Session{})
}

// SessionRepository represents the database for storing sessions
type SessionRepository struct{}

// CreateSessionRepository creates a new SessionRepository IF gorm has been initialized before
func CreateSessionRepository() (*SessionRepository, error) {
	if sqlDatabaseConnection == nil {
		return nil, ErrGormNotInitialized
	}
	return &SessionRepository{}, nil
}

// Create stores a new session
func (rep *SessionRepository) Create(session *models.Session) (err error) {
	err = sqlDatabaseConnection.Create(session).Error
	if err != nil {
		log.Error(0, "Could not store session: %v", err)
	}
	return
}

// Count returns the amount of stored sessions
func (rep *SessionRepository) Count() (count int, err error) {
	err = sqlDatabaseConnection.Model(&models.Session{}).Count(&count).Error
	if err != nil {
		log.Error(0, "Error counting total sessions: %v", err)
	}
	return
}

// Delete deletes a given session
func (rep *SessionRepository) Delete(session *models.Session) (err error) {
	err = sqlDatabaseConnection.Delete(session).Error
	if err != nil {
		log.Error(0, "Could not delete session: %v", err)
	}
	return
}

// DeleteAllForUser deletes all session for one user
func (rep *SessionRepository) DeleteAllForUser(userID int64) (err error) {
	err = sqlDatabaseConnection.Where("user_id = ?", userID).Delete(models.Session{}).Error
	if err != nil {
		log.Error(0, "Could not clean all sessions for user %d: %v", userID, err)
	}
	return
}

// DeleteExpired deletes all expired sessions
func (rep *SessionRepository) DeleteExpired() (err error) {
	log.Trace("Cleaning old sessions")
	err = sqlDatabaseConnection.Where("expires_at < ?", time.Now().UTC().Unix()).Delete(&models.Session{}).Error
	if err != nil {
		log.Error(0, "Deleting expired sessions failed: %v", err)
	}
	return
}

// GetByToken reads and returns a session by token
func (rep *SessionRepository) GetByToken(token string) (session *models.Session, err error) {
	session = &models.Session{}
	err = sqlDatabaseConnection.First(session, "token = ?", token).Error
	if err != nil {
		log.Error(0, "Could not get session by token %s: %v", token, err)
	}
	return
}
