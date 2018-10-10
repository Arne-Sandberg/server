package grpcRouter

import (
	"context"

	"path/filepath"
	"time"

	"github.com/freecloudio/freecloud/fs"
	"github.com/freecloudio/freecloud/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/freecloudio/freecloud/auth"
)

type FilesService struct {
	filesystem *fs.VirtualFilesystem
	notifications chan uint32
}

func NewFilesService(fs *fs.VirtualFilesystem) *FilesService {
	return &FilesService{fs, make(chan uint32, 50)}
}

func (srv *FilesService) ZipFiles(ctx context.Context, req *models.PathsRequest) (*models.Path, error) {
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

	fileInfo, content, err := srv.filesystem.GetDirInfo(user, req.FullPath)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Getting fileInfo for %v failed", req.FullPath)
	}

	return &models.FileInfoResponse{FileInfo: fileInfo, Content: content}, nil
}

func (srv *FilesService) CreateFile(ctx context.Context, req *models.CreateFileRequest) (*models.FileInfo, error) {
	user, _, err := authCheck(req.Auth.Token, false)
	if err != nil {
		return nil, err
	}

	fileInfo, err := srv.filesystem.CreateFile(user, req.FullPath, req.IsDir)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create file: %v", err)
	}

	return fileInfo, nil
}

func (srv *FilesService) UpdateFileInfos(ctx context.Context, req *models.FileInfosUpdateRequest) (*models.EmptyMessage, error) {
	user, _, err := authCheck(req.Auth.Token, false)
	if err != nil {
		return nil, err
	}

	for _, fullPath := range req.FullPaths {
		_, err := srv.filesystem.UpdateFile(user, fullPath, req.FileInfoUpdate)
		if err != nil {
			return nil, status.Error(codes.Internal, "Failed to update user")
		}
	}

	return &models.EmptyMessage{}, nil
}

func (srv *FilesService) DeleteFiles(ctx context.Context, req *models.PathsRequest) (*models.EmptyMessage, error) {
	user, _, err := authCheck(req.Auth.Token, false)
	if err != nil {
		return nil, err
	}

	for _, fullPath := range req.FullPaths {
		err = srv.filesystem.DeleteFile(user, fullPath)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to delete file %v", fullPath)
		}
	}

	return &models.EmptyMessage{}, nil
}

func (srv *FilesService) ShareFiles(ctx context.Context, req *models.ShareRequest) (*models.EmptyMessage, error) {
	fromUser, _, err := authCheck(req.Auth.Token, false)
	if err != nil {
		return nil, err
	}

	for _, fullPath := range req.FullPaths {
		for _, shareWithID := range req.UserIDs {
			toUser, err := auth.GetUserByID(shareWithID)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "Failed to get user %v: %v", shareWithID, err)
			}

			if err := srv.filesystem.ShareFile(fromUser, toUser, fullPath); err != nil {
				return nil, status.Errorf(codes.Internal, "Failed to share %v with %v: %v", fullPath, shareWithID, err)
			}
		}
	}

	return &models.EmptyMessage{}, nil
}

func (srv *FilesService) SearchFiles(ctx context.Context, req *models.SearchRequest) (*models.FileList, error) {
	user, _, err := authCheck(req.Auth.Token, false)
	if err != nil {
		return nil, err
	}

	results, err := srv.filesystem.SearchForFiles(user, req.Term)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to search for %v", req.Term)
	}

	return &models.FileList{FileInfos: results}, nil
}

func (srv *FilesService) GetStarredFiles(ctx context.Context, req *models.Authentication) (*models.FileList, error) {
	user, _, err := authCheck(req.Token, false)
	if err != nil {
		return nil, err
	}

	starredFilesInfo, err := srv.filesystem.ListStarredFilesForUser(user)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to get starred files")
	}

	return &models.FileList{FileInfos: starredFilesInfo}, nil
}

func (srv *FilesService) GetSharedFiles(ctx context.Context, req *models.Authentication) (*models.FileList, error) {
	user, _, err := authCheck(req.Token, false)
	if err != nil {
		return nil, err
	}

	sharedFilesInfo, err := srv.filesystem.ListSharedFilesForUser(user)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to get shared files")
	}

	return &models.FileList{FileInfos: sharedFilesInfo}, nil
}

func (srv *FilesService) RescanOwnFiles(ctx context.Context, req *models.Authentication) (*models.EmptyMessage, error) {
	user, _, err := authCheck(req.Token, false)
	if err != nil {
		return nil, err
	}

	if err := srv.filesystem.ScanUserFolderForChanges(user); err != nil {
		return nil, status.Error(codes.Internal, "Failed to scan folder")
	}

	return &models.EmptyMessage{}, nil
}

func (srv *FilesService) RescanUserFilesByID(ctx context.Context, req *models.UserIDRequest) (*models.EmptyMessage, error) {
	_, _, err := authCheck(req.Auth.Token, true)
	if err != nil {
		return nil, err
	}

	scanUser, err := auth.GetUserByID(req.UserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get user %v", req.UserID)
	}

	if err := srv.filesystem.ScanUserFolderForChanges(scanUser); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to scan folder for %v", req.UserID)
	}

	return &models.EmptyMessage{}, nil
}

func (srv *FilesService) SendUpdateNotification(userID uint32) {
	 srv.notifications <- userID
}

func (srv *FilesService) GetUpdateNotifications(req *models.Authentication, stream models.FilesService_GetUpdateNotificationsServer) error {
	user, _, err := authCheck(req.Token, false)
	if err != nil {
		return err
	}

	msg := &models.EmptyMessage{}
	for userID := range srv.notifications {
		if userID != user.ID {
			continue
		}

		if err := stream.Send(msg); err != nil {
			return err
		}
	}

	return nil
}
