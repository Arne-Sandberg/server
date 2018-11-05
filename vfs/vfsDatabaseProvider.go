package vfs

import "github.com/freecloudio/freecloud/models"

type VFSDatabaseProvider interface {
	InsertFile(fileInfo *models.FileInfo) (err error)
	RemoveFile(fileInfo *models.FileInfo) (err error)
	UpdateFile(fileInfo *models.FileInfo) (err error)
	DeleteFile(fileInfo *models.FileInfo) (err error)

	// Must return an empty instead of an error if nothing could be found
	GetStarredFilesForUser(userID int64) (starredFilesForuser []*models.FileInfo, err error)
	GetSharedFilesForUser(userID int64) (sharedFilesForUser []*models.FileInfo, err error)

	GetDirectoryContent(userID int64, path, dirName string) (dirInfo *models.FileInfo, content []*models.FileInfo, err error)
	GetDirectoryContentWithID(directoryID int64) (content []*models.FileInfo, err error)
	GetFileInfo(userID int64, path, fileName string) (fileInfo *models.FileInfo, err error)
	GetFileInfoWithID(fileID int64) (fileInfo *models.FileInfo, err error)
	SearchForFiles(userID int64, path, fileName string) (results []*models.FileInfo, err error)
	DeleteUserFiles(userID int64) (err error)

	InsertShareEntry(shareEntry *models.ShareEntry) (err error)
	GetShareEntryByID(shareID int64) (shareEntry *models.ShareEntry, err error)
	GetShareEntriesForFile(fileID int64) (shareEntries []*models.ShareEntry, err error)
}
