package grpcRouter

import (
	"context"

	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/fs"
	"github.com/freecloudio/freecloud/models"
	log "gopkg.in/clog.v1"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
)

type AuthService struct {
	filesystem *fs.VirtualFilesystem
}

func NewAuthService(vfs *fs.VirtualFilesystem) *AuthService {
	return &AuthService{vfs}
}

func (srv *AuthService) Signup(ctx context.Context, user *models.User) (*models.Authentication, error) {
	log.Trace("Signing up user: %s %s with email %s", user.FirstName, user.LastName, user.Email)
	session, err := auth.NewUser(user)
	if err == auth.ErrInvalidUserData {
		return nil, status.Error(codes.InvalidArgument,"Invalid user data")
	} else if err == auth.ErrUserAlreadyExists {
		return nil, status.Error(codes.InvalidArgument,"User already exists")
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = srv.filesystem.ScanUserFolderForChanges(user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &models.Authentication{Token: session.GetTokenString()}, nil
}

func (srv *AuthService) Login(ctx context.Context, user *models.User) (resp *models.Authentication, err error) {
	session, err := auth.NewSession(user.Email, user.Password)
	if err == auth.ErrInvalidCredentials {
		return nil, status.Error(codes.Unauthenticated,"Wrong credentials or account does not exist")
	} else if err != nil {
		// TODO: Catch the "not found" error and also return StatusUnauthorized here
		return nil, status.Errorf(codes.Unauthenticated, "Failed to get user %s: %v", user.Email, err)
	}

	return &models.Authentication{Token: session.GetTokenString()}, nil
}

func (srv *AuthService) Logout(ctx context.Context, authReq *models.Authentication) (*models.EmptyMessage, error) {
	_, session, err := authCheck(authReq.Token, false)
	if err != nil {
		return nil, err
	}

	err = auth.RemoveSession(session)
	if err != nil {
		log.Error(0, "Failed to remove session during logout: %v", err)
		return nil, status.Error(codes.Internal, "Failed to delete session")
	}

	return &models.EmptyMessage{}, nil
}

func authCheck(token string, adminReq bool) (user *models.User, session *models.Session, err error) {
	session, err = models.ParseSessionTokenString(token)

	if err != nil {
		err = status.Error(codes.InvalidArgument, "Could not parse session token")
		return
	}

	valid := auth.ValidateSession(session)
	if !valid {
		err = status.Error(codes.Unauthenticated, "Session not valid!")
		return
	}

	user, err = auth.GetUserByID(session.UserID)
	if err != nil {
		log.Error(0, "Filling user data in middleware failed: %v", err)
		err = status.Error(codes.Internal, "Filling user data in middleware failed")
		return
	}

	if adminReq && !user.IsAdmin {
		err = status.Error(codes.PermissionDenied, "Permission denied")
		return
	}

	return
}
