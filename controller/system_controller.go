package controller

import (
	"github.com/go-openapi/runtime/middleware"

	"github.com/freecloudio/freecloud/manager"
	systemAPI "github.com/freecloudio/freecloud/restapi/operations/system"
)

func SystemStatsHandler() middleware.Responder {
	stats := manager.GetStatsManager().GetSystemStats()
	return systemAPI.NewGetSystemStatsOK().WithPayload(stats)
}
