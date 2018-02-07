package auth

import "github.com/freecloudio/freecloud/models"

type SessionProvider interface {
	CleanupExpiredSessions()
	StoreSession(models.Session) error
	RemoveSession(models.Session) error
	SessionIsValid(session models.Session) bool
}
