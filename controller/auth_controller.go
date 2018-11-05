package controller

import (
	"net/http"

	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/models"
	authAPI "github.com/freecloudio/freecloud/restapi/operations/auth"
	"github.com/go-openapi/runtime/middleware"
)

func AuthSignupHandler(user *models.User) middleware.Responder {
	session, err := auth.NewUser(user)
	if err == auth.ErrInvalidUserData {
		return authAPI.NewSignupDefault(http.StatusBadRequest).WithPayload(&models.Error{Message: "Invalid user data"})
	} else if err == auth.ErrUserAlreadyExists {
		return authAPI.NewSignupDefault(http.StatusBadRequest).WithPayload(&models.Error{Message: "User already exists"})
	} else if err != nil {
		return authAPI.NewSignupDefault(http.StatusBadRequest).WithPayload(&models.Error{Message: err.Error()})
	}

	/*err = srv.filesystem.ScanUserFolderForChanges(user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}*/

	return authAPI.NewSignupOK().WithPayload(&models.Token{Token: session.GetTokenString()})
}
