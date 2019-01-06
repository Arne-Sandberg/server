package repository

import (
	"time"

	"github.com/freecloudio/freecloud/models"
	log "gopkg.in/clog.v1"
)

// Add used models to enable auto migration for them
func init() {
	databaseModels = append(databaseModels, &models.Session{})
}

// SessionRepository provides CRUD methods for managing sessions.
// Note: This is not exported on purpose! Reason being that all repositories
// share a common database connection. This means that each repository must
// check if it needs to initialize the connection upon creation.
type SessionRepository struct{}

func CreateSessionRepository() (*SessionRepository, error) {
	if databaseConnection == nil {
		return nil, ErrGormNotInitialized
	}
	return &SessionRepository{}, nil
}

func (rep *SessionRepository) Create(session *models.Session) (err error) {
	err = databaseConnection.Create(session).Error
	if err != nil {
		log.Error(0, "Could not store session: %v", err)
	}
	return
}

func (rep *SessionRepository) Count() (count int, err error) {
	err = databaseConnection.Model(&models.Session{}).Count(&count).Error
	if err != nil {
		log.Error(0, "Error counting total sessions: %v", err)
	}
	return
}

func (rep *SessionRepository) Delete(session *models.Session) (err error) {
	err = databaseConnection.Delete(session).Error
	if err != nil {
		log.Error(0, "Could not delete session: %v", err)
	}
	return
}

func (rep *SessionRepository) DeleteForUser(userID int64) (err error) {
	err = databaseConnection.Where("user_id = ?", userID).Delete(models.Session{}).Error
	if err != nil {
		log.Error(0, "Could not clean all sessions for user %d: %v", userID, err)
	}
	return
}

func (rep *SessionRepository) DeleteExpired() {
	log.Trace("Cleaning old sessions")
	err := databaseConnection.Where("expires_at < ?", time.Now().UTC().Unix()).Delete(&models.Session{}).Error
	if err != nil {
		log.Error(0, "Deleting expired sessions failed: %v", err)
	}
}
