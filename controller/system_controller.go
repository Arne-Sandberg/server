package controller

import (
	"github.com/go-openapi/runtime/middleware"

	"github.com/freecloudio/server/manager"
	systemAPI "github.com/freecloudio/server/restapi/operations/system"
)

func SystemStatsHandler() middleware.Responder {
	stats := manager.GetStatsManager().GetSystemStats()
	return systemAPI.NewGetSystemStatsOK().WithPayload(stats)
}
