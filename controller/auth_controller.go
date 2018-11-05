package controller

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	log "gopkg.in/clog.v1"

	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/models"
	authAPI "github.com/freecloudio/freecloud/restapi/operations/auth"
	"github.com/freecloudio/freecloud/vfs"
)

func AuthSignupHandler(user *models.User) middleware.Responder {
	session, err := auth.NewUser(user)
	if err == auth.ErrInvalidUserData {
		return authAPI.NewSignupDefault(http.StatusBadRequest).WithPayload(&models.Error{Message: "Invalid user data"})
	} else if err == auth.ErrUserAlreadyExists {
		return authAPI.NewSignupDefault(http.StatusBadRequest).WithPayload(&models.Error{Message: "User already exists"})
	} else if err != nil {
		return authAPI.NewSignupDefault(http.StatusInternalServerError).WithPayload(&models.Error{Message: err.Error()})
	}

	err = vfs.ScanUserFolderForChanges(user)
	if err != nil {
		return authAPI.NewSignupDefault(http.StatusInternalServerError).WithPayload(&models.Error{Message: err.Error()})
	}

	return authAPI.NewSignupOK().WithPayload(&models.Token{Token: session.GetTokenString()})
}

func AuthLoginHandler(email, password string) middleware.Responder {
	session, err := auth.NewSession(email, password)
	if err != nil {
		if err != auth.ErrInvalidCredentials {
			log.Warn("Login failed without wrong credentials for user %v: %v", email, err)
		}
		return authAPI.NewLoginDefault(http.StatusUnauthorized).WithPayload(&models.Error{Message: "Wrong credentials or account does not exist"})
	}

	return authAPI.NewSignupOK().WithPayload(&models.Token{Token: session.GetTokenString()})
}
