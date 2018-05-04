package grpc_router

import (
	"context"

	"github.com/freecloudio/freecloud/models"
)

type FilesService struct {
}

func (srv *FilesService) ZipFiles(context.Context, *models.ZipRequest) (*models.DefaultResponse, error) {

}

func (srv *FilesService) GetFileInfo(context.Context, *models.PathRequest) (*models.DirectoryContentResponse, error) {

}

func (srv *FilesService) UpdateFileInfo(context.Context, *models.FileInfo) (*models.FileInfoResponse, error) {

}

func (srv *FilesService) CreateFile(context.Context, *models.FileInfo) (*models.FileInfoResponse, error) {

}

func (srv *FilesService) DeleteFile(context.Context, *models.PathRequest) (*models.DefaultResponse, error) {

}

func (srv *FilesService) ShareFile(context.Context, *models.ShareRequest) (*models.DefaultResponse, error) {

}

func (srv *FilesService) SearchForFile(context.Context, *models.SearchRequest) (*models.FileInfoResponse, error) {

}

func (srv *FilesService) GetStarredFiles(context.Context, *models.Authentication) (*models.DirectoryContentResponse, error) {

}

func (srv *FilesService) GetSharedFiles(context.Context, *models.Authentication) (*models.DirectoryContentResponse, error) {

}

func (srv *FilesService) RescanOwnFiles(context.Context, *models.Authentication) (*models.DefaultResponse, error) {

}

func (srv *FilesService) RescanUserFilesByID(context.Context, *models.Authentication) (*models.DefaultResponse, error) {

}

func (srv *FilesService) GetUpdateNotifications(*models.Authentication, models.FilesService_GetUpdateNotificationsServer) error {

}
