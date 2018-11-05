package auth

import "github.com/freecloudio/freecloud/models"

type SessionProvider interface {
	CleanupExpiredSessions()
	StoreSession(*models.Session) error
	RemoveSession(*models.Session) error
	SessionIsValid(session *models.Session) bool
	TotalSessionCount() (int64, error)
	RemoveUserSessions(userID int64) error
}
