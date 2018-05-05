package grpcRouter

import (
	"context"

	"path/filepath"
	"time"

	"github.com/freecloudio/freecloud/fs"
	"github.com/freecloudio/freecloud/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FilesService struct {
	filesystem *fs.VirtualFilesystem
}

func NewFilesService(fs *fs.VirtualFilesystem) *FilesService {
	return &FilesService{fs}
}

func (srv *FilesService) ZipFiles(ctx context.Context, req *models.ZipRequest) (*models.Path, error) {
	user, _, err := authCheck(req.Auth.Token, false)
	if err != nil {
		return nil, err
	}

	outputFileName := "_" + time.Now().UTC().Format("06-01-02_15-04-05") + ".zip"
	if len(req.FullPaths) == 1 {
		outputFileName = filepath.Base(req.FullPaths[0]) + outputFileName
	} else {
		outputFileName = "fc" + outputFileName
	}

	fullZipPath, err := srv.filesystem.ZipFiles(user, req.FullPaths, outputFileName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Creating zip failed")
	}

	return &models.Path{Path: fullZipPath}, nil
}

func (srv *FilesService) GetFileInfo(ctx context.Context, req *models.PathRequest) (*models.FileInfoResponse, error) {
	user, _, err := authCheck(req.Auth.Token, false)
	if err != nil {
		return nil, err
	}

	fileInfo, content, err := srv.filesystem.ListFilesForUser(user, req.FullPath)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Getting fileInfo for %v failed", req.FullPath)
	}

	return &models.FileInfoResponse{FileInfo: fileInfo, Content: content}, nil
}

func (srv *FilesService) UpdateFileInfo(context.Context, *models.FileInfo) (*models.FileInfo, error) {
	return nil, nil
}

func (srv *FilesService) CreateFile(context.Context, *models.FileInfo) (*models.FileInfo, error) {
	return nil, nil
}

func (srv *FilesService) DeleteFile(context.Context, *models.PathRequest) (*models.EmptyMessage, error) {
	return nil, nil
}

func (srv *FilesService) ShareFile(context.Context, *models.ShareRequest) (*models.EmptyMessage, error) {
	return nil, nil
}

func (srv *FilesService) SearchForFile(context.Context, *models.SearchRequest) (*models.FileInfo, error) {
	return nil, nil
}

func (srv *FilesService) GetStarredFiles(context.Context, *models.Authentication) (*models.FileList, error) {
	return nil, nil
}

func (srv *FilesService) GetSharedFiles(context.Context, *models.Authentication) (*models.FileList, error) {
	return nil, nil
}

func (srv *FilesService) RescanOwnFiles(context.Context, *models.Authentication) (*models.EmptyMessage, error) {
	return nil, nil
}

func (srv *FilesService) RescanUserFilesByID(context.Context, *models.Authentication) (*models.EmptyMessage, error) {
	return nil, nil
}

func (srv *FilesService) GetUpdateNotifications(*models.Authentication, models.FilesService_GetUpdateNotificationsServer) error {
	return nil
}
