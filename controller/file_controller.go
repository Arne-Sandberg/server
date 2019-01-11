package controller

import (
	"net/http"

	"github.com/freecloudio/freecloud/manager"
	"github.com/freecloudio/freecloud/models"
	fileAPI "github.com/freecloudio/freecloud/restapi/operations/file"
	"github.com/go-openapi/runtime/middleware"
)

func FileGetPathInfoHandler(params fileAPI.GetPathInfoParams, principal *models.Principal) middleware.Responder {
	pathInfo, err := manager.GetFileManager().GetPathInfo(principal.User, params.Path)
	if err != nil {
		return fileAPI.NewGetPathInfoDefault(http.StatusBadRequest).WithPayload(&models.Error{Message: err.Error()})
	}

	return fileAPI.NewGetPathInfoOK().WithPayload(pathInfo)
}

func FileCreateHandler(params fileAPI.CreateFileParams, principal *models.Principal) middleware.Responder {
	fileInfo, err := manager.GetFileManager().CreateFile(principal.User, params.CreateFileRequest.FullPath, params.CreateFileRequest.IsDir)
	if err != nil {
		return fileAPI.NewCreateFileDefault(http.StatusInternalServerError).WithPayload(&models.Error{Message: err.Error()})
	}

	return fileAPI.NewCreateFileOK().WithPayload(fileInfo)
}

func FileDeleteHandler(params fileAPI.DeleteFileParams, principal *models.Principal) middleware.Responder {
	err := manager.GetFileManager().DeleteFile(principal.User, params.Path)
	if err != nil {
		return fileAPI.NewCreateFileDefault(http.StatusInternalServerError).WithPayload(&models.Error{Message: err.Error()})
	}

	return fileAPI.NewDeleteFileOK()
}

func FileRescanCurrentUserHandler(params fileAPI.RescanCurrentUserParams, principal *models.Principal) middleware.Responder {
	err := manager.GetFileManager().ScanUserFolderForChanges(principal.User)
	if err != nil {
		return fileAPI.NewRescanCurrentUserDefault(http.StatusInternalServerError).WithPayload(&models.Error{Message: err.Error()})
	}

	return fileAPI.NewRescanCurrentUserOK()
}

func FileRescanUserByIDHandler(params fileAPI.RescanUserByIDParams, principal *models.Principal) middleware.Responder {
	user, err := manager.GetAuthManager().GetUserByID(params.ID)
	if err != nil {
		return fileAPI.NewRescanUserByIDDefault(http.StatusInternalServerError).WithPayload(&models.Error{Message: err.Error()})
	}

	err = manager.GetFileManager().ScanUserFolderForChanges(user)
	if err != nil {
		return fileAPI.NewRescanUserByIDDefault(http.StatusInternalServerError).WithPayload(&models.Error{Message: err.Error()})
	}

	return fileAPI.NewRescanUserByIDOK()
}

func FileGetStarredFileInfos(params fileAPI.GetStarredFileInfosParams, principal *models.Principal) middleware.Responder {
	fileInfos, err := manager.GetFileManager().GetStarredFileInfosForUser(principal.User)
	if err != nil {
		return fileAPI.NewGetStarredFileInfosDefault(http.StatusInternalServerError).WithPayload(&models.Error{Message: err.Error()})
	}

	return fileAPI.NewGetStarredFileInfosOK().WithPayload(&models.FileList{Files: fileInfos})
}
