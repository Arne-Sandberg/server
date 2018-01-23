package auth

import "github.com/freecloudio/freecloud/models"

type SessionProvider interface {
	StoreSession(models.Session) error
	RemoveSession(models.Session) error
	SessionIsValid(session models.Session) bool
}
