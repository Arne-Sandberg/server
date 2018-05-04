package grpcRouter

import (
	"context"

	"github.com/freecloudio/freecloud/models"
	"github.com/freecloudio/freecloud/utils"
	"github.com/freecloudio/freecloud/stats"
)

type SystemService struct {
}

func NewSystemService() *SystemService {
	return &SystemService{}
}

func (srv *SystemService) GetSystemStats(ctx context.Context, authReq *models.Authentication) (*models.SystemStatsResponse, error) {
	user, _, resp := validateTokenAndFillUserData(authReq.Token)
	if resp != nil {
		return &models.SystemStatsResponse{ Meta: resp }, nil
	}

	if !user.IsAdmin {
		return &models.SystemStatsResponse{ Meta: utils.PbForbidden() }, nil
	}

	return &models.SystemStatsResponse{ Meta: utils.PbOK(), Stats: stats.GetSystemStats() }, nil
}
