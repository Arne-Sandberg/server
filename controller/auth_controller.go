package controller

import (
	"github.com/freecloudio/server/restapi/fcerrors"

	"github.com/go-openapi/runtime/middleware"
	log "gopkg.in/clog.v1"

	"github.com/freecloudio/server/manager"
	"github.com/freecloudio/server/models"
	authAPI "github.com/freecloudio/server/restapi/operations/auth"
	userAPI "github.com/freecloudio/server/restapi/operations/user"
)

func AuthSignupHandler(params authAPI.SignupParams) middleware.Responder {
	session, err := manager.GetAuthManager().CreateUser(params.User)
	if err != nil {
		return authAPI.NewSignupDefault(fcerrors.GetStatusCode(err)).WithPayload(&models.Error{Message: err.Error()})
	}

	return authAPI.NewSignupOK().WithPayload(&models.Token{Token: session.GetSessionString()})
}

func AuthLoginHandler(params authAPI.LoginParams) middleware.Responder {
	email := params.Credentials.Email
	password := params.Credentials.Password

	session, err := manager.GetAuthManager().NewSession(email, password)
	if err != nil {
		return authAPI.NewLoginDefault(fcerrors.GetStatusCode(err)).WithPayload(&models.Error{Message: err.Error()})
	}

	return authAPI.NewSignupOK().WithPayload(&models.Token{Token: session.GetSessionString()})
}

func AuthLogoutHandler(params authAPI.LogoutParams, principal *models.Principal) middleware.Responder {
	session, _ := models.ParseSessionString(principal.Token.Token)
	err := manager.GetAuthManager().DeleteSession(session)
	if err != nil {
		log.Error(0, "Failed to remove session during logout: %v", err)
		return authAPI.NewLogoutDefault(fcerrors.GetStatusCode(err)).WithPayload(&models.Error{Message: err.Error()})
	}

	return authAPI.NewLogoutOK()
}

func AuthGetCurrentUserHandler(params userAPI.GetCurrentUserParams, principal *models.Principal) middleware.Responder {
	return userAPI.NewGetCurrentUserOK().WithPayload(principal.User)
}

func AuthGetUserByIDHandler(params userAPI.GetUserByIDParams, principal *models.Principal) middleware.Responder {
	user, err := manager.GetAuthManager().GetUserByID(params.ID)
	if err != nil {
		return userAPI.NewGetUserByIDDefault(fcerrors.GetStatusCode(err)).WithPayload(&models.Error{Message: err.Error()})
	}

	return userAPI.NewGetUserByIDOK().WithPayload(user)
}

func AuthDeleteCurrentUserHandler(params userAPI.DeleteCurrentUserParams, principal *models.Principal) middleware.Responder {
	err := manager.GetAuthManager().DeleteUser(principal.User.ID)
	if err != nil {
		return userAPI.NewDeleteCurrentUserDefault(fcerrors.GetStatusCode(err)).WithPayload(&models.Error{Message: err.Error()})
	}

	return userAPI.NewDeleteCurrentUserOK()
}

func AuthDeleteUserByIDHandler(params userAPI.DeleteUserByIDParams, principal *models.Principal) middleware.Responder {
	err := manager.GetAuthManager().DeleteUser(params.ID)
	if err != nil {
		return userAPI.NewDeleteUserByIDDefault(fcerrors.GetStatusCode(err)).WithPayload(&models.Error{Message: err.Error()})
	}

	return userAPI.NewDeleteUserByIDOK()
}
