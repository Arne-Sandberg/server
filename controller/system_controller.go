package controller

import (
	"github.com/freecloudio/server/restapi/fcerrors"

	"github.com/go-openapi/runtime/middleware"

	"github.com/freecloudio/server/manager"
	systemAPI "github.com/freecloudio/server/restapi/operations/system"
)

func SystemStatsHandler() middleware.Responder {
	stats, err := manager.GetSystemManager().GetSystemStats()
	if err != nil {
		return systemAPI.NewGetSystemStatsDefault(fcerrors.GetStatusCode(err)).WithPayload(fcerrors.GetAPIError(err))
	}

	return systemAPI.NewGetSystemStatsOK().WithPayload(stats)
}
