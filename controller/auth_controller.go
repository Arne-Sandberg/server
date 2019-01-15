package controller

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	log "gopkg.in/clog.v1"

	"github.com/freecloudio/server/manager"
	"github.com/freecloudio/server/models"
	authAPI "github.com/freecloudio/server/restapi/operations/auth"
	userAPI "github.com/freecloudio/server/restapi/operations/user"
)

func AuthSignupHandler(params authAPI.SignupParams) middleware.Responder {
	session, err := manager.GetAuthManager().CreateUser(params.User)
	if err == manager.ErrInvalidUserData {
		return authAPI.NewSignupDefault(http.StatusBadRequest).WithPayload(&models.Error{Message: "Invalid user data"})
	} else if err == manager.ErrUserAlreadyExists {
		return authAPI.NewSignupDefault(http.StatusBadRequest).WithPayload(&models.Error{Message: "User already exists"})
	} else if err != nil {
		return authAPI.NewSignupDefault(http.StatusInternalServerError).WithPayload(&models.Error{Message: err.Error()})
	}

	return authAPI.NewSignupOK().WithPayload(&models.Token{Token: session.GetTokenString()})
}

func AuthLoginHandler(params authAPI.LoginParams) middleware.Responder {
	email := params.Credentials.Email
	password := params.Credentials.Password

	session, err := manager.GetAuthManager().NewSession(email, password)
	if err != nil {
		if err != manager.ErrInvalidCredentials {
			log.Warn("Login failed without wrong credentials for user %v: %v", email, err)
		}
		return authAPI.NewLoginDefault(http.StatusUnauthorized).WithPayload(&models.Error{Message: "Wrong credentials or account does not exist"})
	}

	return authAPI.NewSignupOK().WithPayload(&models.Token{Token: session.GetTokenString()})
}

func AuthLogoutHandler(params authAPI.LogoutParams, principal *models.Principal) middleware.Responder {
	session, _ := models.ParseSessionTokenString(principal.Token.Token)
	err := manager.GetAuthManager().DeleteSession(session)
	if err != nil {
		log.Error(0, "Failed to remove session during logout: %v", err)
		return authAPI.NewLogoutDefault(http.StatusInternalServerError).WithPayload(&models.Error{Message: "Failed to delete session"})
	}

	return authAPI.NewLogoutOK()
}

func AuthGetCurrentUserHandler(params userAPI.GetCurrentUserParams, principal *models.Principal) middleware.Responder {
	return userAPI.NewGetCurrentUserOK().WithPayload(principal.User)
}

func AuthGetUserByIDHandler(params userAPI.GetUserByIDParams, principal *models.Principal) middleware.Responder {
	user, err := manager.GetAuthManager().GetUserByID(params.ID)
	if err != nil {
		return userAPI.NewGetUserByIDDefault(http.StatusInternalServerError).WithPayload(&models.Error{Message: "Could not get user with id"})
	}

	return userAPI.NewGetUserByIDOK().WithPayload(user)
}

func AuthDeleteCurrentUserHandler(params userAPI.DeleteCurrentUserParams, principal *models.Principal) middleware.Responder {
	err := manager.GetAuthManager().DeleteUser(principal.User.ID)
	if err != nil {
		return userAPI.NewDeleteCurrentUserDefault(http.StatusInternalServerError).WithPayload(&models.Error{Message: "Could not delete own user"})
	}

	return userAPI.NewDeleteCurrentUserOK()
}

func AuthDeleteUserByIDHandler(params userAPI.DeleteUserByIDParams, principal *models.Principal) middleware.Responder {
	err := manager.GetAuthManager().DeleteUser(params.ID)
	if err != nil {
		return userAPI.NewDeleteUserByIDDefault(http.StatusInternalServerError).WithPayload(&models.Error{Message: "Could not delete user by id"})
	}

	return userAPI.NewDeleteUserByIDOK()
}
