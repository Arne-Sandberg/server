package grpcRouter

import (
	"context"

	"github.com/freecloudio/freecloud/models"
	"github.com/freecloudio/freecloud/stats"
)

type SystemService struct {
}

func NewSystemService() *SystemService {
	return &SystemService{}
}

func (srv *SystemService) GetSystemStats(ctx context.Context, authReq *models.Authentication) (*models.SystemStats, error) {
	_, _, err := authCheck(authReq.Token, true)
	if err != nil {
		return nil, err
	}

	return stats.GetSystemStats(), nil
}
