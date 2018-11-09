package controller

import (
	systemAPI "github.com/freecloudio/freecloud/restapi/operations/system"
	"github.com/freecloudio/freecloud/stats"
	"github.com/go-openapi/runtime/middleware"
)

func SystemStatsHandler() middleware.Responder {
	stats := stats.GetSystemStats()
	return systemAPI.NewGetSystemStatsOK().WithPayload(stats)
}
