package grpc_router

import (
	"context"

	"github.com/freecloudio/freecloud/models"
)

type AuthService struct {
}

func (srv *AuthService) Signup(context.Context, *models.UserRequest) (*models.AuthResponse, error) {

}

func (srv *AuthService) Login(context.Context, *models.UserRequest) (*models.AuthResponse, error) {

}

func (srv *AuthService) Logout(context.Context, *models.EmptyMessage) (*models.DefaultResponse, error) {

}
