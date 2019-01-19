package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/freecloudio/server/config"
	"github.com/freecloudio/server/repository"
	"github.com/freecloudio/server/utils"

	"errors"
	"mime"

	"github.com/freecloudio/server/models"
	log "gopkg.in/clog.v1"
)

const tmpName = ".tmp"

var (
	ErrFileNotFound     = errors.New("vfs: File not found")
	ErrSharedIntoShared = errors.New("vfs: Moving or copying shared file into shared folder")
	// ErrForbiddenPathName indicates a path having weird characters that nobody should use, also these characters are forbidden on Windows
	ErrForbiddenPathName = errors.New("paths cannot contain the following characters: <>:\"\\|?*")
	ErrFileNotExist      = errors.New("file does not exist")
)

type FileManager struct {
	fileSystemRep *repository.FileSystemRepository
	fileInfoRep   *repository.FileInfoRepository
	shareEntryRep *repository.ShareEntryRepository
}

var fileManager *FileManager

func CreateFileManager(fileSystemRep *repository.FileSystemRepository, fileInfoRep *repository.FileInfoRepository, shareEntryRep *repository.ShareEntryRepository) (*FileManager, error) {
	if fileManager != nil {
		return fileManager, nil
	}

	fileManager = &FileManager{
		fileSystemRep: fileSystemRep,
		fileInfoRep:   fileInfoRep,
		shareEntryRep: shareEntryRep,
	}
	err := fileManager.ScanFSForChanges()

	return fileManager, err
}

func GetFileManager() *FileManager {
	return fileManager
}

func (mgr *FileManager) Close() {}

func (mgr *FileManager) ScanFSForChanges() (err error) {
	existingUsers, err := GetAuthManager().GetAllUsers(true)
	if err != nil {
		log.Error(0, "Could not get exising users: %v", err)
		return
	}

	for _, user := range existingUsers {
		err = mgr.ScanUserFolderForChanges(user)
		if err != nil {
			continue
		}
	}
	return
}

func (mgr *FileManager) ScanUserFolderForChanges(user *models.User) (err error) {
	err = mgr.CreateUserFolders(user.ID)
	if err != nil {
		log.Error(0, "Could not create user folders for %v: %v", user.ID, err)
		return err
	}

	_, err = mgr.scanDirForChanges(user, "/", "")
	if err != nil {
		log.Error(0, "Could not scan directory for user %v: %v", user.ID, err)
		return err
	}
	return
}

func (mgr *FileManager) scanDirForChanges(user *models.User, path, name string) (folderSize int64, err error) {
	// Get all needed data, paths, etc.
	userPath := mgr.getUserPath(user)
	pathInfo, err := mgr.fileSystemRep.GetInfo(userPath, path, name)
	// Return if the scanning dir is a file
	if err != nil || !pathInfo.IsDir {
		return pathInfo.Size, fmt.Errorf("path is not a directory")
	}

	// Get dir contents of fs and db
	fsPath := filepath.Join(path, name)
	fsFiles, err := mgr.fileSystemRep.GetDirectoryContent(userPath, fsPath)
	if err != nil {
		return
	}
	dbPathInfo, dbFiles, err := mgr.fileInfoRep.GetDirectoryContentByPath(user.ID, path, name)
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

func (mgr *FileManager) getUserPath(user *models.User) string {
	return mgr.getUserPathWithID(user.ID)
}

func (mgr *FileManager) getUserPathWithID(userID int64) string {
	return "/" + filepath.Join(strconv.Itoa(int(userID)))
}

// splitPath splits the given full path into the path and the name of the file/dir
func (mgr *FileManager) splitPath(origPath string) (path, name string) {
	origPath = utils.ConvertToSlash(origPath, false)
	if origPath == "/." || origPath == "/" || origPath == "." || origPath == "" {
		return "/", ""
	}

	if strings.HasSuffix(origPath, "/") {
		origPath = origPath[:len(origPath)-1]
	}

	path = utils.ConvertToSlash(filepath.Dir(origPath), true)
	if strings.HasSuffix(path, "./") {
		path = path[:len(path)-2]
	}

	name = filepath.Base(origPath)
	return
}

func (mgr *FileManager) CreateUserFolders(userID int64) error {
	userPath := mgr.getUserPathWithID(userID)

	//Create user dir if not existing and add it to the db
	created, err := mgr.fileSystemRep.CreateDirIfNotExist(userPath)
	if err != nil {
		return fmt.Errorf("failed to create folder for user id %v: %v", userID, err)
	}
	_, err = mgr.fileInfoRep.GetByPath(userID, "/", "")
	if created || repository.IsRecordNotFoundError(err) {
		err = mgr.fileInfoRep.Create(&models.FileInfo{
			Path:        "/",
			Name:        "",
			IsDir:       true,
			OwnerID:     userID,
			LastChanged: utils.GetTimestampNow(),
		})
		if err != nil {
			return fmt.Errorf("failed inserting created root folder for user id %v: %v", userID, err)
		}
	}

	//Create tmp dir for if not existing and add it to the db
	created, err = mgr.fileSystemRep.CreateDirIfNotExist(filepath.Join(userPath, tmpName))
	if err != nil {
		return fmt.Errorf("failed creating tmp folder for user id %v: %v", userID, err)
	}
	_, err = mgr.fileInfoRep.GetByPath(userID, "/", tmpName)
	if created || repository.IsRecordNotFoundError(err) {
		err = mgr.fileInfoRep.Create(&models.FileInfo{
			Path:        "/",
			Name:        tmpName,
			IsDir:       true,
			OwnerID:     userID,
			LastChanged: utils.GetTimestampNow(),
		})
		if err != nil {
			return fmt.Errorf("failed inserting created tmp folder for user id %v: %v", userID, err)
		}
	}

	return nil
}

func (mgr *FileManager) NewFileHandleForUser(user *models.User, path string) (*os.File, error) {
	if !utils.ValidatePath(path) {
		return nil, ErrForbiddenPathName
	}

	filePath, fileName := mgr.splitPath(path)
	folderInfo, err := mgr.GetFileInfo(user, filePath, false)
	if err != nil {
		return nil, err
	}

	return mgr.fileSystemRep.NewHandle(filepath.Join(mgr.getUserPathWithID(folderInfo.OwnerID), folderInfo.Path, folderInfo.Name, fileName))
}

func (mgr *FileManager) FinishNewFile(user *models.User, path string) (err error) {
	if !utils.ValidatePath(path) {
		err = ErrForbiddenPathName
		return
	}

	filePath, fileName := mgr.splitPath(path)
	folderInfo, err := mgr.GetFileInfo(user, filePath, false)
	if err != nil {
		return
	}

	userPath := mgr.getUserPathWithID(folderInfo.OwnerID)
	fileInfo, err := mgr.fileSystemRep.GetInfo(userPath, filepath.Join(folderInfo.Path, folderInfo.Name), fileName)
	if err != nil {
		return
	}

	fileInfo.OwnerID = folderInfo.OwnerID
	fileInfo.ParentID = folderInfo.ID
	err = mgr.fileInfoRep.Create(fileInfo)
	if err != nil {
		return
	}

	//TODO: Make asynchronus scan call for dir sizes?!?

	return
}

func (mgr *FileManager) CreateFile(user *models.User, path string, isDir bool) (fileInfo *models.FileInfo, err error) {
	if exisFileInfo, _ := mgr.GetFileInfo(user, path, true); exisFileInfo != nil && exisFileInfo.ID > 0 {
		return nil, fmt.Errorf("file %v already exists", path)
	}

	if isDir {
		err := mgr.CreateDirectoryForUser(user, path)
		if err != nil {
			return nil, fmt.Errorf("directory creation failed for path '%s': %v", path, err)
		}
	} else {
		file, err := mgr.NewFileHandleForUser(user, path)
		defer file.Close()
		if err != nil {
			return nil, fmt.Errorf("creating new file handle failed for path '%s': %v", path, err)
		}
		err = mgr.FinishNewFile(user, path)
		if err != nil {
			return nil, fmt.Errorf("finishing created file failed for path '%s': %v", path, err)
		}
	}

	fileInfo, err = mgr.GetFileInfo(user, path, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get fileInfo of created file '%s': %v", path, err)
	}

	return fileInfo, nil
}

func (mgr *FileManager) CreateDirectoryForUser(user *models.User, path string) (err error) {
	if !utils.ValidatePath(path) {
		err = ErrForbiddenPathName
		return
	}

	folderPath, folderName := mgr.splitPath(path)
	userPath := mgr.getUserPath(user)

	parFolderInfo, err := mgr.GetFileInfo(user, folderPath, false)
	if err != nil {
		err = fmt.Errorf("could not find parent folder of creating folder in db: %v", err)
		log.Error(0, "%v", err)
		return
	}

	err = mgr.fileSystemRep.CreateDirectory(filepath.Join(userPath, path))
	if err != nil {
		err = fmt.Errorf("error creating directory for user %v: %v", user.ID, err)
		log.Error(0, "%v", err)
		return
	}

	dirInfo := &models.FileInfo{
		Path:        utils.ConvertToSlash(filepath.Join(parFolderInfo.Path, parFolderInfo.Name), true),
		Name:        folderName,
		IsDir:       true,
		OwnerID:     parFolderInfo.OwnerID,
		LastChanged: utils.GetTimestampNow(),
		ParentID:    parFolderInfo.ID,
	}
	err = mgr.fileInfoRep.Create(dirInfo)
	if err != nil {
		return
	}

	return
}

func (mgr *FileManager) GetPathInfo(user *models.User, path string) (*models.PathInfo, error) {
	dirInfo, err := mgr.GetFileInfo(user, path, true)
	if err != nil {
		return nil, err
	}

	var content []*models.FileInfo
	if dirInfo.IsDir {
		content, err = mgr.fileInfoRep.GetDirectoryContentByID(dirInfo.ID)
		if err != nil {
			return nil, err
		}

		if dirInfo.ShareID > 0 || dirInfo.OwnerID != user.ID {
			for _, file := range content {
				file.Path = path
			}
		}
	}

	return &models.PathInfo{FileInfo: dirInfo, Content: content}, nil
}

func (mgr *FileManager) GetStarredFileInfosForUser(user *models.User) (starredFilesInfo []*models.FileInfo, err error) {
	starredFilesInfo, err = mgr.fileInfoRep.GetStarredFileInfosForUser(user.ID)
	if err != nil {
		return
	}
	return
}

func (mgr *FileManager) ListSharedFilesForUser(user *models.User) (sharedFilesInfo []*models.FileInfo, err error) {
	sharedFilesInfo, err = mgr.fileInfoRep.GetSharedFileInfosForUser(user.ID)
	if err != nil {
		return
	}
	return
}

// GetFileInfo returns the stored fileInfo for a given path and user resolving shared folders
// Set adaptSharePath to true if the returned path of the fileInfo should be from the root of the requesting user
// Set adaptSharePath to false if it should stay the orig path of the sharing user
func (mgr *FileManager) GetFileInfo(user *models.User, requestedPath string, adaptSharedPath bool) (*models.FileInfo, error) {
	filePath, fileName := mgr.splitPath(requestedPath)
	fileInfo, err := mgr.fileInfoRep.GetByPath(user.ID, filePath, fileName)

	if err == nil && fileInfo.ShareID <= 0 { // File exists in db for user and is owned by him: Directly return info
		return fileInfo, nil
	} else if err == nil && fileInfo.ShareID > 0 { // File exists in db for user and is shared with him: Get orig file and return it
		shareEntry, err := mgr.getCheckedShareEntry(user.ID, fileInfo.ShareID)
		if err != nil {
			return nil, err
		}

		finalFileInfo := &models.FileInfo{}
		finalFileInfo, err = mgr.fileInfoRep.GetByID(shareEntry.FileID)
		if err != nil {
			return nil, err
		}

		if adaptSharedPath {
			finalFileInfo.Path = filePath
		}

		return finalFileInfo, err
	} else { // File does not exist in db: Check recusively if it is in a shared folder otherwise return not found
		var sharedParentInfo *models.FileInfo
		var removedPath string
		parentPath := filePath
		parentName := ""
		for {
			removedPath = filepath.Join(removedPath, parentName)
			parentPath, parentName = mgr.splitPath(parentPath)

			fileInfo, err = mgr.fileInfoRep.GetByPath(user.ID, parentPath, parentName)

			if err == nil && fileInfo.ShareID <= 0 { // Found existing parent but it is not shared --> Requested file does not exist
				break
			} else if err == nil && fileInfo.ShareID > 0 { // Found existing parent that is shared --> Check share and remember parent
				shareEntry, err := mgr.getCheckedShareEntry(user.ID, fileInfo.ShareID)
				if err != nil {
					return nil, err
				}

				sharedParentInfo = &models.FileInfo{}
				sharedParentInfo, err = mgr.fileInfoRep.GetByID(shareEntry.FileID)
				if err != nil {
					return nil, err
				}

				break
			}
		}

		if sharedParentInfo == nil {
			return nil, ErrFileNotFound
		}

		finalFileInfo := &models.FileInfo{}
		finalPath := utils.ConvertToSlash(filepath.Join(sharedParentInfo.Path, sharedParentInfo.Name, removedPath), true)
		finalFileInfo, err = mgr.fileInfoRep.GetByPath(sharedParentInfo.OwnerID, finalPath, fileName)
		if err != nil {
			return nil, err
		}

		if adaptSharedPath {
			finalFileInfo.Path = filePath
		}

		return finalFileInfo, nil
	}
}

func (mgr *FileManager) GetDownloadPath(user *models.User, path string) (downloadURL, filename string, err error) {
	fileInfo, err := mgr.GetFileInfo(user, path, false)
	if err != nil {
		return
	}

	downloadURL = mgr.fileSystemRep.GetDownloadPath(filepath.Join(mgr.getUserPathWithID(fileInfo.OwnerID), fileInfo.Path, fileInfo.Name))

	_, fileName := mgr.splitPath(path)
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
		_, name := mgr.splitPath(paths[0])
		outputName = name + ".zip"
	}
	zipPath = filepath.Join(tmpName, outputName)
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
	/*
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
	*/
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
		folderContent, err = mgr.fileInfoRep.GetDirectoryContentByID(fileInfo.ID)
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
			folderContent, err = mgr.fileInfoRep.GetDirectoryContentByID(fileInfo.ID)
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
		var folderContent []*models.FileInfo
		folderContent, err = mgr.fileInfoRep.GetDirectoryContentByID(fileInfo.ID)
		for _, contentInfo := range folderContent {
			err = mgr.deleteFileInDB(contentInfo)
			if err != nil {
				return
			}
		}
	}
	return
}

func (mgr *FileManager) SearchForFiles(user *models.User, path string) (results []*models.FileInfo, err error) {
	filePath, fileName := mgr.splitPath(path)
	return mgr.fileInfoRep.SearchForFileInfo(user.ID, filePath, fileName)
}

func (mgr *FileManager) DeleteUserFiles(user *models.User) (err error) {
	err = mgr.fileInfoRep.DeleteUserFileInfos(user.ID)
	if err != nil {
		return
	}

	err = mgr.fileSystemRep.Delete(mgr.getUserPath(user))
	if err != nil {
		return
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

func (mgr *FileManager) ShareFile(fromUser, toUser *models.User, path string) (err error) {
	filePath, fileName := mgr.splitPath(path)
	// Get fileInfo without resolving shared files
	fileInfo, err := mgr.fileInfoRep.GetByPath(fromUser.ID, filePath, fileName)
	if err != nil {
		return
	}

	if fileInfo.ShareID > 0 {
		return fmt.Errorf("sharing shared files is forbidden")
	}

	if res, _ := mgr.isInSharedByMe(fromUser.ID, toUser.ID, fileInfo); res {
		return fmt.Errorf("file is already shared with this user")
	}

	if exisFileInfo, _ := mgr.fileInfoRep.GetByPath(toUser.ID, "/", fileInfo.Name); exisFileInfo != nil && exisFileInfo.ID > 0 {
		return fmt.Errorf("file already exists at target user")
	}

	shareEntry := &models.ShareEntry{
		OwnerID:      fromUser.ID,
		SharedWithID: toUser.ID,
		FileID:       fileInfo.ID,
	}
	err = mgr.shareEntryRep.Create(shareEntry)
	if err != nil {
		return
	}

	sharedParentInfo, err := mgr.fileInfoRep.GetByPath(toUser.ID, "/", "")
	if err != nil {
		return
	}

	sharedFileInfo := &models.FileInfo{
		Path:        "/",
		Name:        fileInfo.Name,
		IsDir:       fileInfo.IsDir,
		Size:        fileInfo.Size,
		OwnerID:     toUser.ID,
		LastChanged: utils.GetTimestampNow(),
		MimeType:    fileInfo.MimeType,
		ParentID:    sharedParentInfo.ID,
		ShareID:     shareEntry.ID,
		Starred:     false,
	}

	err = mgr.fileInfoRep.Create(sharedFileInfo)
	return
}

func (mgr *FileManager) isInSharedByMe(userID, withUserID int64, fileInfo *models.FileInfo) (bool, error) {
	parentID := fileInfo.ID

	for parentID != 0 {
		fileInfo, err := mgr.fileInfoRep.GetByID(parentID)
		if err != nil {
			return false, err
		}

		if shareEntries, _ := mgr.shareEntryRep.GetByFileID(fileInfo.ID); len(shareEntries) < 0 {
			if withUserID <= 0 {
				return true, nil
			}

			for _, entry := range shareEntries {
				if entry.SharedWithID == withUserID {
					return true, nil
				}
			}
		}

		parentID = fileInfo.ParentID
	}

	return false, nil
}

func (mgr *FileManager) getCheckedShareEntry(userID, shareID int64) (shareEntry *models.ShareEntry, err error) {
	shareEntry, err = mgr.shareEntryRep.GetByID(shareID)
	if err != nil {
		return
	}

	if shareEntry.SharedWithID != userID {
		err = fmt.Errorf("user of shareEntry not matching with requested user")
		return
	}

	return
}

func (mgr *FileManager) GetAvatarForUser(userID int64) (string, error) {
	p := filepath.Join(config.GetString("fs.base_directory"), config.GetString("fs.avatar_directory"), string(userID))
	_, err := os.Stat(p)
	if err != nil {
		log.Error(0, "DB says user %d has avatar, but the file was not found: %v", userID, err)
		return "", err
	}
	return p, nil
}

func (mgr *FileManager) NewAvatarFileHandleForuser(userID int64) (*os.File, error) {
	return mgr.fileSystemRep.NewHandle(filepath.Join(config.GetString("fs.base_directory"), config.GetString("fs.avatar_directory"), string(userID)))
}
