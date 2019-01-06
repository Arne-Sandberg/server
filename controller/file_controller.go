package controller

import (
	"io"
	"net/http"

	"github.com/freecloudio/freecloud/models"
	fileAPI "github.com/freecloudio/freecloud/restapi/operations/file"
	"github.com/go-openapi/runtime/middleware"
	log "gopkg.in/clog.v1"
)

func PathInfoHandler(fullPath string, user *models.User) middleware.Responder {
	pathInfo, err := vfs.GetPathInfo(user, fullPath)
	if err != nil {
		return fileAPI.NewGetPathInfoDefault(http.StatusBadRequest).WithPayload(&models.Error{Message: err.Error()})
	}

	return fileAPI.NewGetPathInfoOK().WithPayload(pathInfo)
}

func CreateFileHandler(fullPath string, isDir bool, user *models.User) middleware.Responder {
	fileInfo, err := vfs.CreateFile(user, fullPath, isDir)
	if err != nil {
		return fileAPI.NewCreateFileDefault(http.StatusInternalServerError).WithPayload(&models.Error{Message: err.Error()})
	}

	return fileAPI.NewCreateFileOK().WithPayload(fileInfo)
}

func FileUploadHandler(path string, upFile io.ReadCloser, user *models.User) middleware.Responder {
	log.Trace("Uploading file to %s", path)

	return fileAPI.NewUploadFileOK()
}
