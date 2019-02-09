package controller

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/freecloudio/server/manager"
	"github.com/freecloudio/server/models"
	systemAPI "github.com/freecloudio/server/restapi/operations/system"
)

func SystemStatsHandler() middleware.Responder {
	stats, err := manager.GetSystemManager().GetSystemStats()
	if err != nil {
		return systemAPI.NewGetSystemStatsDefault(http.StatusInternalServerError).WithPayload(&models.Error{Message: err.Error()})
	}

	return systemAPI.NewGetSystemStatsOK().WithPayload(stats)
}
