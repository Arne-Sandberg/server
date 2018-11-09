package controller

import (
	"net/http"

	"github.com/freecloudio/freecloud/models"
	fileAPI "github.com/freecloudio/freecloud/restapi/operations/file"
	"github.com/freecloudio/freecloud/vfs"
	"github.com/go-openapi/runtime/middleware"
)

func PathInfoHandler(fullPath string, user *models.User) middleware.Responder {
	pathInfo, err := vfs.GetPathInfo(user, fullPath)
	if err != nil {
		return fileAPI.NewGetPathInfoDefault(http.StatusBadRequest).WithPayload(&models.Error{Message: err.Error()})
	}

	return fileAPI.NewGetPathInfoOK().WithPayload(pathInfo)
}
