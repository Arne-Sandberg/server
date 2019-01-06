package controller

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	log "gopkg.in/clog.v1"

	"github.com/freecloudio/freecloud/manager"
	"github.com/freecloudio/freecloud/models"
	authAPI "github.com/freecloudio/freecloud/restapi/operations/auth"
)

func AuthSignupHandler(user *models.User) middleware.Responder {
	session, err := manager.GetAuthManager().CreateUser(user)
	if err == manager.ErrInvalidUserData {
		return authAPI.NewSignupDefault(http.StatusBadRequest).WithPayload(&models.Error{Message: "Invalid user data"})
	} else if err == manager.ErrUserAlreadyExists {
		return authAPI.NewSignupDefault(http.StatusBadRequest).WithPayload(&models.Error{Message: "User already exists"})
	} else if err != nil {
		return authAPI.NewSignupDefault(http.StatusInternalServerError).WithPayload(&models.Error{Message: err.Error()})
	}

	return authAPI.NewSignupOK().WithPayload(&models.Token{Token: session.GetTokenString()})
}

func AuthLoginHandler(email, password string) middleware.Responder {
	session, err := manager.GetAuthManager().NewSession(email, password)
	if err != nil {
		if err != manager.ErrInvalidCredentials {
			log.Warn("Login failed without wrong credentials for user %v: %v", email, err)
		}
		return authAPI.NewLoginDefault(http.StatusUnauthorized).WithPayload(&models.Error{Message: "Wrong credentials or account does not exist"})
	}

	return authAPI.NewSignupOK().WithPayload(&models.Token{Token: session.GetTokenString()})
}

func AuthLogoutHandler(principal *models.Principal) middleware.Responder {
	session, _ := models.ParseSessionTokenString(principal.Token.Token)
	err := manager.GetAuthManager().DeleteSession(session)
	if err != nil {
		log.Error(0, "Failed to remove session during logout: %v", err)
		return authAPI.NewLogoutDefault(http.StatusInternalServerError).WithPayload(&models.Error{Message: "Failed to delete session"})
	}

	return authAPI.NewLogoutOK()
}
