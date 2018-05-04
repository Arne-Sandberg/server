package grpcRouter

import (
	"context"

	"github.com/freecloudio/freecloud/models"
)

type SystemService struct {
}

func NewSystemService() *SystemService {
	return &SystemService{}
}

func (srv *SystemService) GetSystemStats(context.Context, *models.Authentication) (*models.SystemStatsResponse, error) {
	return nil, nil
}
