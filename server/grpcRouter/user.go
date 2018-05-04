package grpcRouter

import (
	"context"

	"github.com/freecloudio/freecloud/models"
)

type UserService struct {
}

func NewUserService() *UserService {
	return &UserService{}
}

func (srv *UserService) GetOwnUser(context.Context, *models.Authentication) (*models.UserResponse, error) {
	return nil, nil
}

func (srv *UserService) GetUserByID(context.Context, *models.UserID) (*models.UserResponse, error) {
	return nil, nil
}

func (srv *UserService) GetUserByEmail(context.Context, *models.UserEmail) (*models.UserResponse, error) {
	return nil, nil
}

func (srv *UserService) UpdateOwnUser(context.Context, *models.User) (*models.UserResponse, error) {
	return nil, nil
}

func (srv *UserService) UpdateUserByID(context.Context, *models.User) (*models.UserResponse, error) {
	return nil, nil
}

func (srv *UserService) DeleteOwnUser(context.Context, *models.Authentication) (*models.DefaultResponse, error) {
	return nil, nil
}

func (srv *UserService) DeleteUserByID(context.Context, *models.UserID) (*models.DefaultResponse, error) {
	return nil, nil
}
