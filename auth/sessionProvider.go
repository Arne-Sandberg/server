package auth

import "github.com/riesinger/freecloud/models"

type SessionProvider interface {
	StoreSession(models.Session) error
	SessionIsValid(session models.Session) bool
}
