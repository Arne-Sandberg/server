package grpcRouter

import (
	"context"

	"github.com/freecloudio/freecloud/models"
)

type AuthService struct {
}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (srv *AuthService) Signup(context.Context, *models.UserRequest) (*models.AuthResponse, error) {
	return nil, nil
}

func (srv *AuthService) Login(context.Context, *models.UserRequest) (*models.AuthResponse, error) {
	return nil, nil
}

func (srv *AuthService) Logout(context.Context, *models.EmptyMessage) (*models.DefaultResponse, error) {
	return nil, nil
}
