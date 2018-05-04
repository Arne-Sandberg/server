package grpc_router

import (
	"context"

	"github.com/freecloudio/freecloud/models"
)

type UserService struct {
}

func (srv *UserService) GetOwnUser(context.Context, *models.Authentication) (*models.UserResponse, error) {

}

func (srv *UserService) GetUserByID(context.Context, *models.UserID) (*models.UserResponse, error) {

}

func (srv *UserService) GetUserByEmail(context.Context, *models.UserEmail) (*models.UserResponse, error) {

}

func (srv *UserService) UpdateOwnUser(context.Context, *models.User) (*models.UserResponse, error) {

}

func (srv *UserService) UpdateUserByID(context.Context, *models.User) (*models.UserResponse, error) {

}

func (srv *UserService) DeleteOwnUser(context.Context, *models.Authentication) (*models.DefaultResponse, error) {

}

func (srv *UserService) DeleteUserByID(context.Context, *models.UserID) (*models.DefaultResponse, error) {

}
