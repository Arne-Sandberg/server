package grpc_router

import (
	"context"

	"github.com/freecloudio/freecloud/models"
)

type SystemService struct {
}

func (srv *SystemService) GetSystemStats(context.Context, *models.Authentication) (*models.SystemStatsResponse, error) {

}
