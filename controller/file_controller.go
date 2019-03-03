package controller

/*func FileGetPathInfoHandler(params fileAPI.GetPathInfoParams, principal *models.Principal) middleware.Responder {
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

func FileGetShareEntryByIDHandler(params fileAPI.GetShareEntryByIDParams, principal *models.Principal) middleware.Responder {
	shareEntry, err := manager.GetFileManager().GetShareEntryByID(params.ShareID, principal.User)
	if err != nil {
		return fileAPI.NewGetShareEntryByIDDefault(http.StatusInternalServerError).WithPayload(&models.Error{Message: err.Error()})
	}

	return fileAPI.NewGetShareEntryByIDOK().WithPayload(shareEntry)
}

func FileDeleteShareEntryByIDHandler(params fileAPI.DeleteShareEntryByIDParams, principal *models.Principal) middleware.Responder {
	err := manager.GetFileManager().DeleteShareEntryByID(params.ShareID, principal.User)
	if err != nil {
		return fileAPI.NewDeleteShareEntryByIDDefault(http.StatusInternalServerError).WithPayload(&models.Error{Message: err.Error()})
	}

	return fileAPI.NewDeleteShareEntryByIDOK()
}*/
