package manager

import (
	"os"
	"path/filepath"
	"time"

	"github.com/freecloudio/server/restapi/fcerrors"
	"github.com/freecloudio/server/utils"

	"github.com/freecloudio/server/models"
	"github.com/freecloudio/server/repository"
	log "gopkg.in/clog.v1"
)

// FileManager has methods for creating/updating/deleting/sharing files
type FileManager struct {
	fileSystemRep      *repository.FileSystemRepository
	fileInfoRep        *repository.FileInfoRepository
	shareEntryRep      *repository.ShareEntryRepository
	tmpName            string
	tmpExpiry          int
	tmpCleanupInterval int
	done               chan struct{}
}

var fileManager *FileManager

// CreateFileManager creates a new singleton FileManager which can be used immediately, tmpExpiry and tmpCleanupInterval are in hours
func CreateFileManager(fileSystemRep *repository.FileSystemRepository, fileInfoRep *repository.FileInfoRepository, shareEntryRep *repository.ShareEntryRepository, tmpName string, tmpExpiry, tmpCleanupInterval int) (*FileManager, error) {
	if fileManager != nil {
		return fileManager, nil
	}

	fileManager = &FileManager{
		fileSystemRep:      fileSystemRep,
		fileInfoRep:        fileInfoRep,
		shareEntryRep:      shareEntryRep,
		tmpName:            tmpName,
		tmpExpiry:          tmpExpiry,
		tmpCleanupInterval: tmpCleanupInterval,
		done:               make(chan struct{}),
	}
	err := fileManager.ScanFSForChanges()
	go fileManager.cleanupExpiredTmpRoutine()
	return fileManager, err
}

// GetFileManager returns the singleton instance of the FileManager
func GetFileManager() *FileManager {
	return fileManager
}

// Close is used to end running tasks
func (mgr *FileManager) Close() {
	mgr.done <- struct{}{}
}

func (mgr *FileManager) cleanupExpiredTmpRoutine() {
	log.Trace("Tmp cleaner will run every %v hours", mgr.tmpCleanupInterval)
	mgr.cleanupTmp()
	ticker := time.NewTicker(time.Hour * time.Duration(mgr.tmpCleanupInterval))
	for {
		select {
		case <-mgr.done:
			return
		case <-ticker.C:
			log.Trace("Cleaning expired tmp")
			mgr.cleanupTmp()
		}
	}
}

func (mgr *FileManager) cleanupTmp() {
	// TODO: Implement this
}

// ScanFSForChanges runs ScanUserFolderForChanges for all users
func (mgr *FileManager) ScanFSForChanges() (err error) {
	/*existingUsers, err := GetAuthManager().GetAllUsers()
	if err != nil {
		log.Error(0, "Could not get exising users: %v", err)
		return
	}

	for _, user := range existingUsers {
		err = mgr.ScanUserFolderForChanges(user)
		if err != nil {
			continue
		}
	}*/
	return
}

// ScanUserFolderForChanges compares the db with the fs and fixes differences in the db for a specific user
func (mgr *FileManager) ScanUserFolderForChanges(user *models.User) (err error) {
	/*err = mgr.CreateUserFolders(user.ID)
	if err != nil {
		log.Error(0, "Could not create user folders for %v: %v", user.ID, err)
		return err
	}

	_, err = mgr.scanDirForChanges(user, "/", "")
	if err != nil {
		log.Error(0, "Could not scan directory for user %v: %v", user.ID, err)
		return err
	}*/
	return
}

/*
func (mgr *FileManager) scanDirForChanges(user *models.User, path, name string) (folderSize int64, err error) {
	fsPath := filepath.Join(path, name)

	// Get all needed data, paths, etc.
	userPath := mgr.getUserPath(user)
	pathInfo, err := mgr.fileSystemRep.GetInfo(userPath, fsPath)
	// Return if the scanning dir is a file
	if err != nil || !pathInfo.IsDir {
		return pathInfo.Size, fmt.Errorf("path is not a directory")
	}

	// Get dir contents of fs and db
	fsFiles, err := mgr.fileSystemRep.GetDirectoryInfo(userPath, fsPath)
	if err != nil {
		return
	}
	dbPathInfo, dbFiles, err := mgr.getDirectoryContentByPath(user.ID, path, name)
	if err != nil {
		return
	}

	for _, fsFile := range fsFiles {
		// Find fs file in dbFiles by name
		dbIt := -1
		for it, dbFile := range dbFiles {
			if dbFile.Name == fsFile.Name {
				dbIt = it
				break
			}
		}

		if dbIt == -1 {
			// File not yet in db --> Add it
			fsFile.OwnerID = user.ID
			fsFile.ParentID = dbPathInfo.ID
			err = mgr.fileInfoRep.Create(fsFile)
			if err != nil {
				log.Error(0, "Error inserting into db: %v", err)
				return
			}
		} else {
			// File found in db files --> Check whether an update is needed
			dbFile := dbFiles[dbIt]
			if (!fsFile.IsDir && fsFile.Size != dbFile.Size) || fsFile.LastChanged != dbFile.LastChanged || fsFile.IsDir != dbFile.IsDir {
				dbFile.Size = fsFile.Size
				dbFile.LastChanged = fsFile.LastChanged
				dbFile.IsDir = fsFile.IsDir
				err = mgr.fileInfoRep.Update(dbFile)
				if err != nil {
					log.Error(0, "Error updating file in db: %v", err)
					return
				}
			}

			// Delete file from db list as it is now used
			dbFiles[dbIt] = dbFiles[len(dbFiles)-1]
			dbFiles[len(dbFiles)-1] = nil
			dbFiles = dbFiles[:len(dbFiles)-1]
		}

		// If it is a file directly add the size; If it is an dir then scan it and add the size of the dir
		if !fsFile.IsDir {
			folderSize += fsFile.Size
		} else {
			subFolderSize, err := mgr.scanDirForChanges(user, fsFile.Path, fsFile.Name)
			if err != nil {
				log.Error(0, "Error scanning subfolder: %v", err)
				return folderSize, err
			}
			folderSize += subFolderSize
		}
	}

	// Delete remaining files from dbList in db as they are deleted from the fs
	for _, dbFile := range dbFiles {
		if dbFile.ShareID > 0 {
			continue
		}

		err = mgr.fileInfoRep.Delete(dbFile.ID)
		if err != nil {
			log.Error(0, "Error removing file from db: %v", err)
			return
		}
	}

	// If the size of the folder has changed then update it in the db
	if dbPathInfo.Size != folderSize {
		dbPathInfo.Size = folderSize
		err = mgr.fileInfoRep.Update(dbPathInfo)
		if err != nil {
			log.Error(0, "Error updating file in db: %v", err)
			return
		}
	}

	return
}
*/

func (mgr *FileManager) getUserPath(username string) string {
	return "/" + filepath.Join(username)
}

// GetPathInfo returns the pathInfo - fileInfo and content - of a path
func (mgr *FileManager) GetPathInfo(username, path string) (pathInfo *models.PathInfo, err error) {
	pathInfo = &models.PathInfo{}

	pathInfo.FileInfo, err = mgr.fileInfoRep.GetByPath(username, path)
	if err != nil && repository.IsRecordNotFoundError(err) {
		return nil, fcerrors.Wrap(err, fcerrors.FileNotExists)
	} else if err != nil {
		return nil, fcerrors.Wrap(err, fcerrors.Database)
	}

	if pathInfo.FileInfo.IsDir {
		pathInfo.Content, err = mgr.fileInfoRep.GetDirectoryContentByPath(username, path)
		if err != nil {
			return nil, fcerrors.Wrap(err, fcerrors.Database)
		}
	}

	return
}

// GetFileInfo returns the fileInfo for an user and path
func (mgr *FileManager) GetFileInfo(username, path string) (fileInfo *models.FileInfo, err error) {
	fileInfo, err = mgr.fileInfoRep.GetByPath(username, path)
	if err != nil && repository.IsRecordNotFoundError(err) {
		return nil, fcerrors.Wrap(err, fcerrors.FileNotExists)
	} else if err != nil {
		return nil, fcerrors.Wrap(err, fcerrors.Database)
	}

	return
}

// CreateUserFolders creates the root and tmp folder for an user
func (mgr *FileManager) CreateUserFolders(username string) error {
	userPath := mgr.getUserPath(username)

	//Create user dir if not existing and add it to the db
	created, err := mgr.fileSystemRep.CreateDirectory(userPath)
	if err != nil {
		return fcerrors.Wrap(err, fcerrors.Filesystem)
	}
	if created {
		err = mgr.fileInfoRep.CreateUserFolder(username)
		if err != nil {
			return fcerrors.Wrap(err, fcerrors.Database)
		}
	}

	// Create tmp dir
	err = mgr.CreateDirectory(username, mgr.tmpName)
	if err != nil {
		return err
	}

	return nil
}

// CreateFile combines CreateDirectory or NewFileHandle + FinishNewFile to create an empty directory or file
func (mgr *FileManager) CreateFile(username string, path string, isDir bool) error {
	if isDir {
		err := mgr.CreateDirectory(username, path)
		if err != nil {
			return err
		}
	} else {
		file, err := mgr.NewFileHandle(username, path)
		defer file.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// NewFileHandle returns a fileHandle from the fs
func (mgr *FileManager) NewFileHandle(username, path string) (*os.File, error) {
	if !utils.ValidatePath(path) {
		return nil, fcerrors.New(fcerrors.PathNotValid)
	}

	filePath, fileName := utils.SplitPath(path)

	owner, ownerParentPath, err := mgr.fileInfoRep.GetOwnerPath(username, filePath)
	if err != nil && repository.IsRecordNotFoundError(err) {
		return nil, fcerrors.New(fcerrors.FileNotExists)
	} else if err != nil {
		return nil, fcerrors.Wrap(err, fcerrors.Database)
	}

	ownerPath := mgr.getUserPath(owner)
	handlePath := filepath.Join(ownerPath, ownerParentPath, fileName)
	fileHandle, err := mgr.fileSystemRep.CreateHandle(handlePath)
	if err != nil {
		return nil, fcerrors.Wrap(err, fcerrors.Filesystem)
	}

	fileInfo, err := mgr.fileSystemRep.GetInfo(ownerPath, filepath.Join(ownerParentPath, fileName))
	if err != nil {
		fileHandle.Close()
		return nil, fcerrors.Wrap(err, fcerrors.Filesystem)
	}

	fileInfo.OwnerUsername = owner
	err = mgr.fileInfoRep.Create(fileInfo)
	if err != nil {
		fileHandle.Close()
		return nil, fcerrors.Wrap(err, fcerrors.Database)
	}

	return fileHandle, nil
}

// CreateDirectory creates a new dir for the user; Returns no error if the dir already exists in the fs
func (mgr *FileManager) CreateDirectory(username, path string) error {
	if !utils.ValidatePath(path) {
		return fcerrors.New(fcerrors.PathNotValid)
	}

	filePath, fileName := utils.SplitPath(path)

	owner, ownerParentPath, err := mgr.fileInfoRep.GetOwnerPath(username, filePath)
	if err != nil && repository.IsRecordNotFoundError(err) {
		return fcerrors.New(fcerrors.FileNotExists)
	} else if err != nil {
		return fcerrors.Wrap(err, fcerrors.Database)
	}

	ownerPath := mgr.getUserPath(owner)
	fsPath := filepath.Join(ownerPath, ownerParentPath, fileName)
	created, err := mgr.fileSystemRep.CreateDirectory(fsPath)
	if !created {
		return nil
	} else if err != nil {
		return fcerrors.Wrap(err, fcerrors.Filesystem)
	}

	dirInfo, err := mgr.fileSystemRep.GetInfo(ownerPath, path)
	if err != nil {
		return fcerrors.Wrap(err, fcerrors.Filesystem)
	}
	dirInfo.OwnerUsername = owner
	err = mgr.fileInfoRep.Create(dirInfo)
	if err != nil {
		return fcerrors.Wrap(err, fcerrors.Database)
	}

	return nil
}

// SearchForFiles searches in the given path for a given term and returns all results that contain the term
func (mgr *FileManager) SearchForFiles(username, path, term string) (results []*models.FileInfo, err error) {
	return mgr.fileInfoRep.Search(username, path, term)
}

// DeleteUserFiles deletes all user files in the db and in the fs
func (mgr *FileManager) DeleteUserFiles(username string) (err error) {
	err = mgr.fileInfoRep.DeleteUserFileInfos(username)
	if err != nil {
		return
	}

	err = mgr.fileSystemRep.Delete(mgr.getUserPath(username))
	if err != nil {
		return
	}
	return
}

/*
func (mgr *FileManager) GetStarredFileInfosForUser(user *models.User) (starredFilesInfo []*models.FileInfo, err error) {
	starredFilesInfo, err = mgr.fileInfoRep.GetStarredFileInfosByUser(user.ID)
	if err != nil {
		return
	}
	return
}

func (mgr *FileManager) ListSharedFilesForUser(user *models.User) (sharedFilesInfo []*models.FileInfo, err error) {
	sharedFilesInfo, err = mgr.fileInfoRep.GetSharedFileInfosByUser(user.ID)
	if err != nil {
		return
	}
	return
}

func (mgr *FileManager) GetDownloadPath(user *models.User, path string) (downloadURL, filename string, err error) {
	fileInfo, err := mgr.GetFileInfo(user, path, false)
	if err != nil {
		return
	}

	downloadURL = mgr.fileSystemRep.GetDownloadPath(filepath.Join(mgr.getUserPathWithID(fileInfo.OwnerID), fileInfo.Path, fileInfo.Name))

	_, fileName := utils.SplitPath(path)
	filename = fileName
	return
}

// ZipFiles zips all given files/directories of paths to a zip archive with the given name in the temp folder
func (mgr *FileManager) ZipFiles(user *models.User, paths []string) (zipPath string, err error) {
	for it := 0; it < len(paths); it++ {
		var fileInfo *models.FileInfo
		fileInfo, err = mgr.GetFileInfo(user, paths[it], false)
		if err != nil {
			return
		}

		paths[it] = filepath.Join(mgr.getUserPathWithID(fileInfo.OwnerID), fileInfo.Path, fileInfo.Name)
	}

	outputName := time.Now().Format("2006.01.02_15:04:05.zip")
	if len(paths) == 1 {
		_, name := utils.SplitPath(paths[0])
		outputName = name + ".zip"
	}
	zipPath = filepath.Join(mgr.tmpName, outputName)
	outputPath := filepath.Join(mgr.getUserPath(user), zipPath)

	err = mgr.fileSystemRep.Zip(paths, outputPath)
	if err != nil {
		return
	}

	err = mgr.FinishNewFile(user, zipPath)
	if err != nil {
		return
	}
	return
}

func (mgr *FileManager) UpdateFile(user *models.User, path string, updatedFileInfo *models.FileInfoUpdate) (fileInfo *models.FileInfo, err error) {
		fileInfo, err = vfs.GetFileInfo(user, path)
		if err != nil {
			return
		}

		var newPath string
		if path, ok := updatedFileInfo.PathOO.(*models.FileInfoUpdate_Path); ok == true && utils.ValidatePath(path.Path) {
			newPath = path.Path
		} else {
			newPath = fileInfo.Path
		}

		var newName string
		if name, ok := updatedFileInfo.NameOO.(*models.FileInfoUpdate_Name); ok == true && utils.ValidatePath(name.Name) {
			newName = name.Name
		} else {
			newName = fileInfo.Name
		}

		var newStarred bool
		if starred, ok := updatedFileInfo.StarredOO.(*models.FileInfoUpdate_Starred); ok == true {
			newStarred = starred.Starred
		} else {
			newStarred = fileInfo.Starred
		}

		var copyFlag bool
		if cpy, ok := updatedFileInfo.CopyOO.(*models.FileInfoUpdate_Copy); ok == true {
			copyFlag = cpy.Copy
		}

		if newPath != fileInfo.Path || newName != fileInfo.Name {
			newFolderInfo, err := vfs.GetFileInfo(user, newPath)
			if err != nil {
				log.Error(0, "Error getting parent for changed file %v%v: %v", fileInfo.Path, fileInfo.Name, err)
				return nil, err
			}

			// Has shareID and into something shared by me or shared with me should be blocked
			if res, _ := vfs.isInSharedByMe(user.ID, 0, newFolderInfo); fileInfo.ShareID > 0 && (res || newFolderInfo.ShareID > 0 || newFolderInfo.OwnerID != user.ID) {
				return nil, ErrSharedIntoShared
			}

			fileInfo.Starred = newStarred

			if !copyFlag {
				err = vfs.moveFile(user, fileInfo, newName, newFolderInfo)
			} else {
				err = vfs.copyFile(user, fileInfo, newName, newFolderInfo)
			}
			if err != nil {
				return nil, err
			}

		} else if newStarred != fileInfo.Starred {
			fileInfo.LastChanged = utils.GetTimestampNow()
			fileInfo.Starred = newStarred
			err = mgr.fileInfoRep.UpdateFile(fileInfo)
			if err != nil {
				log.Error(0, "Error updating file in db: %v", err)
				return
			}
		}

		//TODO: Make asynchronus scan call for dir sizes?!?
	return
}

func (mgr *FileManager) moveFile(user *models.User, fileInfo *models.FileInfo, newName string, newFolderInfo *models.FileInfo) (err error) {
	userPath := mgr.getUserPath(user)
	oldPath := filepath.Join(userPath, fileInfo.Path, fileInfo.Name)

	fileInfo.LastChanged = utils.GetTimestampNow()
	if newName != fileInfo.Name {
		fileInfo.Name = newName
		fileInfo.MimeType = mime.TypeByExtension(filepath.Ext(fileInfo.Name))
	}

	err = mgr.moveFileInDB(user, fileInfo, newFolderInfo)
	if err != nil {
		return
	}

	if fileInfo.ShareID <= 0 {
		newPath := filepath.Join(userPath, fileInfo.Path, fileInfo.Name)
		err = mgr.fileSystemRep.Move(oldPath, newPath)
		if err != nil {
			log.Error(0, "Error moving file from %v to %v: %v", oldPath, newPath, err)
			return
		}
	}

	return nil
}

func (mgr *FileManager) moveFileInDB(user *models.User, fileInfo *models.FileInfo, parentFileInfo *models.FileInfo) (err error) {
	fileInfo.Path = utils.ConvertToSlash(filepath.Join(parentFileInfo.Path, parentFileInfo.Name), true)
	fileInfo.ParentID = parentFileInfo.ID

	err = mgr.fileInfoRep.Update(fileInfo)
	if err != nil {
		log.Error(0, "Error updating file in db: %v", err)
		return
	}

	if fileInfo.IsDir {
		var folderContent []*models.FileInfo
		folderContent, err = mgr.fileInfoRep.GetDirectoryContentByID(user.ID, fileInfo.ID)
		for _, contentInfo := range folderContent {
			err = mgr.moveFileInDB(user, contentInfo, fileInfo)
			if err != nil {
				return
			}
		}
	}
	return
}

func (mgr *FileManager) copyFile(user *models.User, fileInfo *models.FileInfo, newName string, newParentFileInfo *models.FileInfo) (err error) {
	if newName == "" {
		newName = fileInfo.Name
	}
	parentPath := utils.ConvertToSlash(filepath.Join(newParentFileInfo.Path, newParentFileInfo.Name), true)
	newPath := filepath.Join(parentPath, newName)

	if fileInfo.ShareID <= 0 {
		if fileInfo.IsDir {
			err = mgr.CreateDirectoryForUser(user, newPath)
			if err != nil {
				return
			}

			var newFolderInfo *models.FileInfo
			newFolderInfo, err = mgr.GetFileInfo(user, newPath, false)
			if err != nil {
				return
			}

			var folderContent []*models.FileInfo
			folderContent, err = mgr.fileInfoRep.GetDirectoryContentByID(user.ID, fileInfo.ID)
			for _, contentInfo := range folderContent {
				err = mgr.copyFile(user, contentInfo, "", newFolderInfo)
				if err != nil {
					return
				}
			}
		} else {
			userPath := mgr.getUserPath(user)
			oldPath := filepath.Join(fileInfo.Path, fileInfo.Name)
			err = mgr.fileSystemRep.Copy(filepath.Join(userPath, oldPath), filepath.Join(userPath, newPath))
			if err != nil {
				return
			}
			err = mgr.FinishNewFile(user, newPath)
			if err != nil {
				return
			}
		}
	} else {
		newFileInfo := *fileInfo
		newFileInfo.Path = parentPath
		newFileInfo.ParentID = newParentFileInfo.ID
		newFileInfo.Name = newName

		err = mgr.fileInfoRep.Create(&newFileInfo)
		if err != nil {
			return
		}
	}

	return
}

func (mgr *FileManager) DeleteFile(user *models.User, path string) (err error) {
	var fileInfo *models.FileInfo
	fileInfo, err = mgr.GetFileInfo(user, path, false)
	if err != nil {
		return
	}

	err = mgr.deleteFileInDB(fileInfo)
	if err != nil {
		return
	}

	if fileInfo.ShareID <= 0 {
		err = mgr.fileSystemRep.Delete(filepath.Join(mgr.getUserPathWithID(fileInfo.OwnerID), fileInfo.Path, fileInfo.Name))
		if err != nil {
			return
		}
	}

	var shareEntries []*models.ShareEntry
	shareEntries, err = mgr.shareEntryRep.GetByFileID(fileInfo.ID)
	if err != nil {
		return
	}
	for _, shareEntry := range shareEntries {
		err = mgr.shareEntryRep.Delete(shareEntry.ID)
		if err != nil {
			return
		}
	}

	//TODO: Make asynchronus scan call for dir sizes?!?

	return
}

func (mgr *FileManager) deleteFileInDB(fileInfo *models.FileInfo) (err error) {
	err = mgr.fileInfoRep.Delete(fileInfo.ID)

	if fileInfo.IsDir {
		// TODO: Delete in db directly with where parent == ?
	}
	return
}

func (mgr *FileManager) ShareFiles(fromUser *models.User, toUserIDs []int64, paths []string) error {
	type failedShareStruct struct {
		toUserMail string
		path       string
	}
	failedShares := []*failedShareStruct{}

	for _, toUserID := range toUserIDs {
		toUser, err := GetAuthManager().GetUserByID(toUserID)
		if err != nil {
			return err
		}

		for _, path := range paths {
			err := mgr.ShareFile(fromUser, toUser, path)
			if err != nil {
				log.Error(0, "failed to share '%s' to user '%d': %v", path, toUserID, err)
				failedShares = append(failedShares, &failedShareStruct{toUser.Email, path})
			}
		}
	}

	if len(failedShares) > 0 {
		var sb strings.Builder
		for _, failedShare := range failedShares {
			sb.WriteString(fmt.Sprintf("%s: %s\n", failedShare.toUserMail, failedShare.path))
		}

		return fmt.Errorf("failed to share one or mutliple files to an user: %s", sb.String())
	}

	return nil
}

*/
