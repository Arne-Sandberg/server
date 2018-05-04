package grpcRouter

import (
	"context"

	"github.com/freecloudio/freecloud/models"
)

type FilesService struct {
}

func NewFilesService() *FilesService {
	return &FilesService{}
}

func (srv *FilesService) ZipFiles(context.Context, *models.ZipRequest) (*models.DefaultResponse, error) {
	return nil, nil
}

func (srv *FilesService) GetFileInfo(context.Context, *models.PathRequest) (*models.DirectoryContentResponse, error) {
	return nil, nil
}

func (srv *FilesService) UpdateFileInfo(context.Context, *models.FileInfo) (*models.FileInfoResponse, error) {
	return nil, nil
}

func (srv *FilesService) CreateFile(context.Context, *models.FileInfo) (*models.FileInfoResponse, error) {
	return nil, nil
}

func (srv *FilesService) DeleteFile(context.Context, *models.PathRequest) (*models.DefaultResponse, error) {
	return nil, nil
}

func (srv *FilesService) ShareFile(context.Context, *models.ShareRequest) (*models.DefaultResponse, error) {
	return nil, nil
}

func (srv *FilesService) SearchForFile(context.Context, *models.SearchRequest) (*models.FileInfoResponse, error) {
	return nil, nil
}

func (srv *FilesService) GetStarredFiles(context.Context, *models.Authentication) (*models.DirectoryContentResponse, error) {
	return nil, nil
}

func (srv *FilesService) GetSharedFiles(context.Context, *models.Authentication) (*models.DirectoryContentResponse, error) {
	return nil, nil
}

func (srv *FilesService) RescanOwnFiles(context.Context, *models.Authentication) (*models.DefaultResponse, error) {
	return nil, nil
}

func (srv *FilesService) RescanUserFilesByID(context.Context, *models.Authentication) (*models.DefaultResponse, error) {
	return nil, nil
}

func (srv *FilesService) GetUpdateNotifications(*models.Authentication, models.FilesService_GetUpdateNotificationsServer) error {
	return nil
}
