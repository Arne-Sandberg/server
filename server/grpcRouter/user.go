package grpcRouter

import (
	"context"

	"github.com/freecloudio/freecloud/models"
	"github.com/freecloudio/freecloud/auth"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
)

type UserService struct {
}

func NewUserService() *UserService {
	return &UserService{}
}

func (srv *UserService) GetOwnUser(ctx context.Context, authReq *models.Authentication) (*models.User, error) {
	user, _, err := authCheck(authReq.Token, false)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (srv *UserService) GetUserByID(ctx context.Context, req *models.UserIDRequest) (*models.User, error) {
	_, _, err := authCheck(req.Auth.Token, false)
	if err != nil {
		return nil, err
	}

	user, err := auth.GetUserByID(req.UserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error getting user with ID %v", req.UserID)
	}

	return user, nil
}

func (srv *UserService) GetUserByEmail(ctx context.Context, req *models.UserEmailRequest) (*models.User, error) {
	_, _, err := authCheck(req.Auth.Token, false)
	if err != nil {
		return nil, err
	}

	user, err := auth.GetUserByEmail(req.UserEmail)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error getting user with email %v", req.UserEmail)
	}

	return user, nil
}

func (srv *UserService) UpdateOwnUser(ctx context.Context, req *models.UserUpdateRequest) (*models.User, error) {
	user, _, err := authCheck(req.Auth.Token, false)
	if err != nil {
		return nil, err
	}

	if !user.IsAdmin {
		req.UserUpdate.IsAdminOO = nil
	}

	return auth.UpdateUser(user.ID, req.UserUpdate)
}

func (srv *UserService) DeleteOwnUser(ctx context.Context, authReq *models.Authentication) (*models.EmptyMessage, error) {
	user, _, err := authCheck(authReq.Token, false)
	if err != nil {
		return nil, err
	}

	err = auth.DeleteUser(user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Deleting user failed: %v", err)
	}

	return &models.EmptyMessage{}, nil
}

func (srv *UserService) DeleteUserByID(ctx context.Context, req *models.UserIDRequest) (*models.EmptyMessage, error) {
	_, _, err := authCheck(req.Auth.Token, true)
	if err != nil {
		return nil, err
	}

	err = auth.DeleteUser(req.UserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Deleting user failed: %v", err)
	}

	return &models.EmptyMessage{}, nil
}
