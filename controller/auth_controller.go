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
	token, err := manager.GetAuthManager().CreateUser(params.User)
	if err != nil {
		return authAPI.NewSignupDefault(fcerrors.GetStatusCode(err)).WithPayload(fcerrors.GetAPIError(err))
	}

	return authAPI.NewSignupOK().WithPayload(token)
}

func AuthLoginHandler(params authAPI.LoginParams) middleware.Responder {
	email := params.Credentials.UsernameOrEmail
	password := params.Credentials.Password

	token, err := manager.GetAuthManager().LoginUser(email, password)
	if err != nil {
		return authAPI.NewLoginDefault(fcerrors.GetStatusCode(err)).WithPayload(fcerrors.GetAPIError(err))
	}

	return authAPI.NewSignupOK().WithPayload(token)
}

func AuthLogoutHandler(params authAPI.LogoutParams, principal *models.Principal) middleware.Responder {
	session := &models.Session{Token: string(principal.Token)}
	err := manager.GetAuthManager().DeleteSession(session)
	if err != nil {
		log.Error(0, "Failed to remove session during logout: %v", err)
		return authAPI.NewLogoutDefault(fcerrors.GetStatusCode(err)).WithPayload(fcerrors.GetAPIError(err))
	}

	return authAPI.NewLogoutOK()
}

func AuthGetCurrentUserHandler(params userAPI.GetCurrentUserParams, principal *models.Principal) middleware.Responder {
	return userAPI.NewGetCurrentUserOK().WithPayload(principal.User)
}

func AuthGetUserByUsernameHandler(params userAPI.GetUserByUsernameParams, principal *models.Principal) middleware.Responder {
	user, err := manager.GetAuthManager().GetUserByUsername(params.Username)
	if err != nil {
		return userAPI.NewGetUserByUsernameDefault(fcerrors.GetStatusCode(err)).WithPayload(fcerrors.GetAPIError(err))
	}

	return userAPI.NewGetUserByUsernameOK().WithPayload(user)
}

func AuthDeleteCurrentUserHandler(params userAPI.DeleteCurrentUserParams, principal *models.Principal) middleware.Responder {
	err := manager.GetAuthManager().DeleteUser(principal.User.Username)
	if err != nil {
		return userAPI.NewDeleteCurrentUserDefault(fcerrors.GetStatusCode(err)).WithPayload(fcerrors.GetAPIError(err))
	}

	return userAPI.NewDeleteCurrentUserOK()
}

func AuthDeleteUserByUsernameHandler(params userAPI.DeleteUserByUsernameParams, principal *models.Principal) middleware.Responder {
	err := manager.GetAuthManager().DeleteUser(params.Username)
	if err != nil {
		return userAPI.NewDeleteUserByUsernameDefault(fcerrors.GetStatusCode(err)).WithPayload(fcerrors.GetAPIError(err))
	}

	return userAPI.NewDeleteUserByUsernameOK()
}
