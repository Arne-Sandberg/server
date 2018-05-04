package grpcRouter

import (
	"context"
	"github.com/freecloudio/freecloud/models"
	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/utils"
	"github.com/freecloudio/freecloud/fs"
	log "gopkg.in/clog.v1"
)

type AuthService struct {
	filesystem *fs.VirtualFilesystem
}

func NewAuthService(vfs *fs.VirtualFilesystem) *AuthService {
	return &AuthService{ vfs }
}

func (srv *AuthService) Signup(ctx context.Context, user *models.User) (resp *models.AuthResponse, err error) {
	log.Trace("Signing up user: %s %s with email %s", user.FirstName, user.LastName, user.Email)
	session, err := auth.NewUser(user)
	if err == auth.ErrInvalidUserData {
		return &models.AuthResponse{ Meta: utils.PbBadRequest("Invalid user data") }, nil
	} else if err == auth.ErrUserAlreadyExists {
		return &models.AuthResponse{ Meta: utils.PbBadRequest("User already exists") }, nil
	} else if err != nil {
		return
	}

	err = srv.filesystem.ScanUserFolderForChanges(user)
	if err != nil {
		return
	}

	resp = &models.AuthResponse{
		Meta: utils.PbCreated(),
		Token: session.GetTokenString(),
	}
	return
}

func (srv *AuthService) Login(ctx context.Context, user *models.User) (resp *models.AuthResponse, err error) {
	session, err := auth.NewSession(user.Email, user.Password)
	if err == auth.ErrInvalidCredentials {
		return &models.AuthResponse{ Meta: utils.PbUnauthorized("Wrong credentials or account does not exist") }, nil
	} else if err != nil {
		// TODO: Catch the "not found" error and also return StatusUnauthorized here
		return &models.AuthResponse{ Meta: utils.PbUnauthorized("Failed to get user %s: %v", user.Email, err) }, nil
	}

	resp = &models.AuthResponse{
		Meta: utils.PbOK(),
		Token: session.GetTokenString(),
	}
	return
}

func (srv *AuthService) Logout(ctx context.Context, authReq *models.Authentication) (*models.DefaultResponse, error) {
	session, err := models.ParseSessionTokenString(authReq.Token)
	if err != nil {
		return utils.PbBadRequest("Failed to parse token"), nil
	}

	err = auth.RemoveSession(session)
	if err != nil {
		log.Error(0, "Failed to remove session during logout: %v", err)
		return utils.PbInternalServerError("Failed to delete session"), nil
	}

	return utils.PbOK(), nil
}
