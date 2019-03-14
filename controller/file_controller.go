package controller

import (
	"net/http"

	"github.com/freecloudio/server/manager"
	"github.com/freecloudio/server/models"
	"github.com/freecloudio/server/restapi/operations/file"
	"github.com/go-openapi/runtime/middleware"
)

func FileGetPathInfoHandler(params file.GetPathInfoParams, principal *models.Principal) middleware.Responder {
	pathInfo, err := manager.GetFileManager().GetPathInfo(principal.User.Username, params.Path)
	if err != nil {
		return file.NewGetPathInfoDefault(http.StatusBadRequest).WithPayload(&models.Error{Message: err.Error()})
	}

	return file.NewGetPathInfoOK().WithPayload(pathInfo)
}

func FileCreateHandler(params file.CreateFileParams, principal *models.Principal) middleware.Responder {
	err := manager.GetFileManager().CreateFile(principal.User.Username, params.CreateFileRequest.FullPath, params.CreateFileRequest.IsDir)
	if err != nil {
		return file.NewCreateFileDefault(http.StatusInternalServerError).WithPayload(&models.Error{Message: err.Error()})
	}

	return file.NewCreateFileOK()
}

/*
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

func FileGetStarredFileInfosHandler(params fileAPI.GetStarredFileInfosParams, principal *models.Principal) middleware.Responder {
	fileInfos, err := manager.GetFileManager().GetStarredFileInfosForUser(principal.User)
	if err != nil {
		return fileAPI.NewGetStarredFileInfosDefault(http.StatusInternalServerError).WithPayload(&models.Error{Message: err.Error()})
	}

	return fileAPI.NewGetStarredFileInfosOK().WithPayload(&models.FileList{Files: fileInfos})
}

func FileZipFilesHandler(params fileAPI.ZipFilesParams, principal *models.Principal) middleware.Responder {
	zipPath, err := manager.GetFileManager().ZipFiles(principal.User, params.Paths.Paths)
	if err != nil {
		return fileAPI.NewZipFilesDefault(http.StatusInternalServerError).WithPayload(&models.Error{Message: err.Error()})
	}

	return fileAPI.NewZipFilesOK().WithPayload(&models.Path{Path: zipPath})
}

func FileShareFilesHandler(params fileAPI.ShareFilesParams, principal *models.Principal) middleware.Responder {
	err := manager.GetFileManager().ShareFiles(principal.User, params.ShareRequest.Users, params.ShareRequest.Paths)
	if err != nil {
		return fileAPI.NewShareFilesDefault(http.StatusInternalServerError).WithPayload(&models.Error{Message: err.Error()})
	}

	return fileAPI.NewShareFilesOK()
}
*/
