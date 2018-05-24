package fs

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/freecloudio/freecloud/utils"

	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/models"
	log "gopkg.in/clog.v1"
	"errors"
)

type vfsDatabase interface {
	InsertFile(fileInfo *models.FileInfo) (err error)
	RemoveFile(fileInfo *models.FileInfo) (err error)
	UpdateFile(fileInfo *models.FileInfo) (err error)
	DeleteFile(fileInfo *models.FileInfo) (err error)

	// Must return an empty instead of an error if nothing could be found
	GetStarredFilesForUser(userID uint32) (starredFilesForuser []*models.FileInfo, err error)
	GetSharedFilesForUser(userID uint32) (sharedFilesForUser []*models.FileInfo, err error)

	GetDirectoryContent(userID uint32, path, dirName string) (dirInfo *models.FileInfo, content []*models.FileInfo, err error)
	GetDirectoryContentWithID(directoryID uint32) (content []*models.FileInfo, err error)
	GetFileInfo(userID uint32, path, fileName string) (fileInfo *models.FileInfo, err error)
	GetFileInfoWithID(fileID uint32) (fileInfo *models.FileInfo, err error)
	SearchForFiles(userID uint32, path, fileName string) (results []*models.FileInfo, err error)
	DeleteUserFiles(userID uint32) (err error)

	InsertShareEntry(shareEntry *models.ShareEntry) (err error)
	GetShareEntryByID(shareID uint32) (shareEntry *models.ShareEntry, err error)
}

const TmpName = ".tmp"

var ErrFileNotFound = errors.New("vfs: File not found")

type VirtualFilesystem struct {
	fs      Filesystem
	db      vfsDatabase
}

func NewVirtualFilesystem(fs Filesystem, db vfsDatabase) (vfs *VirtualFilesystem, err error) {
	vfs = &VirtualFilesystem{fs, db }
	err = vfs.ScanFSForChanges()

	return
}

func (vfs *VirtualFilesystem) Close() {}

func (vfs *VirtualFilesystem) ScanFSForChanges() (err error) {
	log.Trace("Get existing users")
	existingUsers, err := auth.GetAllUsers(true)
	if err != nil {
		log.Error(0, "Could not get exising users: %v", err)
		return
	}

	for _, user := range existingUsers {
		err = vfs.ScanUserFolderForChanges(user)
		if err != nil {
			continue
		}
	}
	return
}

func (vfs *VirtualFilesystem) ScanUserFolderForChanges(user *models.User) (err error) {
	err = vfs.CreateUserFolders(user.ID)
	if err != nil {
		log.Error(0, "Could not create user folders for %v: %v", user.ID, err)
		return err
	}

	_, err = vfs.scanDirForChanges(user, "/", "")
	if err != nil {
		log.Error(0, "Could not scan directory for user %v: %v", user.ID, err)
		return err
	}
	return
}

func (vfs *VirtualFilesystem) scanDirForChanges(user *models.User, path, name string) (folderSize int64, err error) {
	// Get all needed data, paths, etc.
	userPath := vfs.getUserPath(user)
	pathInfo, err := vfs.fs.GetFileInfo(userPath, path, name)
	// Return if the scanning dir is a file
	if err != nil || !pathInfo.IsDir {
		return pathInfo.Size, fmt.Errorf("path is not a directory")
	}

	// Get dir contents of fs and db
	fsPath := filepath.Join(path, name)
	fsFiles, err := vfs.fs.GetDirectoryContent(userPath, fsPath)
	if err != nil {
		return
	}
	dbPathInfo, dbFiles, err := vfs.db.GetDirectoryContent(user.ID, path, name)
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
			err = vfs.db.InsertFile(fsFile)
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
				err = vfs.db.UpdateFile(dbFile)
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
			subFolderSize, err := vfs.scanDirForChanges(user, fsFile.Path, fsFile.Name)
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

		err = vfs.db.RemoveFile(dbFile)
		if err != nil {
			log.Error(0, "Error removing file from db: %v", err)
			return
		}
	}

	// If the size of the folder has changed then update it in the db
	if dbPathInfo.Size != folderSize {
		dbPathInfo.Size = folderSize
		err = vfs.db.UpdateFile(dbPathInfo)
		if err != nil {
			log.Error(0, "Error updating file in db: %v", err)
			return
		}
	}

	return
}

func (vfs *VirtualFilesystem) getUserPath(user *models.User) string {
	return vfs.getUserPathWithID(user.ID)
}

func (vfs *VirtualFilesystem) getUserPathWithID(userID uint32) string {
	return "/" + filepath.Join(strconv.Itoa(int(userID)))
}

// splitPath splits the given full path into the path and the name of the file/dir
func (vfs *VirtualFilesystem) splitPath(origPath string) (path, name string) {
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

func (vfs *VirtualFilesystem) CreateUserFolders(userID uint32) error {
	userPath := vfs.getUserPathWithID(userID)

	//Create user dir if not existing and add it to the db
	created, err := vfs.fs.CreateDirIfNotExist(userPath)
	if err != nil {
		return fmt.Errorf("failed to create folder for user id %v: %v", userID, err)
	}
	_, err = vfs.db.GetFileInfo(userID, "/", "")
	if created || err != nil {
		err = vfs.db.InsertFile(&models.FileInfo{
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
	created, err = vfs.fs.CreateDirIfNotExist(filepath.Join(userPath, TmpName))
	if err != nil {
		return fmt.Errorf("failed creating tmp folder for user id %v: %v", userID, err)
	}
	_, err = vfs.db.GetFileInfo(userID, "/", TmpName)
	if created || err != nil {
		err = vfs.db.InsertFile(&models.FileInfo{
			Path:        "/",
			Name:        TmpName,
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

func (vfs *VirtualFilesystem) NewFileHandleForUser(user *models.User, path string) (*os.File, error) {
	if !utils.ValidatePath(path) {
		return nil, ErrForbiddenPathName
	}

	return vfs.fs.NewFileHandle(filepath.Join(vfs.getUserPath(user), path))
}

func (vfs *VirtualFilesystem) FinishNewFile(user *models.User, path string) (err error) {
	if !utils.ValidatePath(path) {
		err = ErrForbiddenPathName
		return
	}

	filePath, fileName := vfs.splitPath(path)
	userPath := vfs.getUserPath(user)
	fileInfo, err := vfs.fs.GetFileInfo(userPath, filePath, fileName)
	if err != nil {
		return
	}

	folderPath, folderName := vfs.splitPath(filePath)
	folderInfo, err := vfs.db.GetFileInfo(user.ID, folderPath, folderName)
	if err != nil {
		return
	}

	fileInfo.OwnerID = user.ID
	fileInfo.ParentID = folderInfo.ID
	err = vfs.db.InsertFile(fileInfo)
	if err != nil {
		return
	}

	//TODO: Make asynchronus scan call for dir sizes?!?

	return
}

func (vfs *VirtualFilesystem) CreateDirectoryForUser(user *models.User, path string) (err error) {
	if !utils.ValidatePath(path) {
		err = ErrForbiddenPathName
		return
	}

	folderPath, folderName := vfs.splitPath(path)
	userPath := vfs.getUserPath(user)

	parFolderPath, parFolderName := vfs.splitPath(folderPath)
	parFolderInfo, err := vfs.db.GetFileInfo(user.ID, parFolderPath, parFolderName)
	if err != nil {
		err = fmt.Errorf("could not find parent folder of folder in db: %v", err)
		log.Error(0, "%v", err)
		return
	}

	err = vfs.fs.CreateDirectory(filepath.Join(userPath, path))
	if err != nil {
		err = fmt.Errorf("error creating directory for user %v: %v", user.ID, err)
		log.Error(0, "%v", err)
		return
	}

	dirInfo := &models.FileInfo{
		Path:        utils.ConvertToSlash(folderPath, true),
		Name:        folderName,
		IsDir:       true,
		OwnerID:     user.ID,
		LastChanged: utils.GetTimestampNow(),
		ParentID:    parFolderInfo.ID,
	}
	err = vfs.db.InsertFile(dirInfo)
	if err != nil {
		return
	}

	return
}

func (vfs *VirtualFilesystem) ListFilesForUser(user *models.User, path string) (dirInfo *models.FileInfo, content []*models.FileInfo, err error) {
	folderPath, folderName := vfs.splitPath(path)
	dirInfo, content, err = vfs.db.GetDirectoryContent(user.ID, folderPath, folderName)
	if err != nil {
		return
	}
	return
}

func (vfs *VirtualFilesystem) ListStarredFilesForUser(user *models.User) (starredFilesInfo []*models.FileInfo, err error) {
	starredFilesInfo, err = vfs.db.GetStarredFilesForUser(user.ID)
	if err != nil {
		return
	}
	return
}

func (vfs *VirtualFilesystem) ListSharedFilesForUser(user *models.User) (sharedFilesInfo []*models.FileInfo, err error) {
	sharedFilesInfo, err = vfs.db.GetSharedFilesForUser(user.ID)
	if err != nil {
		return
	}
	return
}

func (vfs *VirtualFilesystem) GetFileInfo(user *models.User, path string) (fileInfo *models.FileInfo, err error) {
	filePath, fileName := vfs.splitPath(path)
	fileInfo, err = vfs.db.GetFileInfo(user.ID, filePath, fileName)
	if err != nil {
		return
	}
	return
}

func (vfs *VirtualFilesystem) GetDownloadPath(user *models.User, path string) (downloadURL, filename string, err error) {
	filePath, fileName := vfs.splitPath(path)
	fileInfo, err := vfs.db.GetFileInfo(user.ID, filePath, fileName)
	if err != nil {
		return
	}

	if fileInfo.ShareID > 0 {
		var shareEntry *models.ShareEntry
		shareEntry, err = vfs.CheckShareEntry(user.ID, fileInfo.ShareID)
		if err != nil {
			return
		}

		var origFileInfo *models.FileInfo
		origFileInfo, err = vfs.db.GetFileInfoWithID(shareEntry.FileID)
		if err != nil {
			return
		}
		userPath := vfs.getUserPathWithID(origFileInfo.OwnerID)
		downloadURL = vfs.fs.GetDownloadPath(filepath.Join(userPath, fileInfo.Path, fileInfo.Name))
	} else {
		userPath := vfs.getUserPath(user)
		downloadURL = vfs.fs.GetDownloadPath(filepath.Join(userPath, path))
	}

	filename = fileName
	return
}

// ZipFiles zips all given files/directories of paths to a zip archive with the given name in the temp folder
func (vfs *VirtualFilesystem) ZipFiles(user *models.User, paths []string, outputName string) (zipPath string, err error) {
	if !utils.ValidatePath(outputName) {
		err = ErrForbiddenPathName
		return
	}

	userPath := vfs.getUserPath(user)
	for it := 0; it < len(paths); it++ {
		filePath, fileName := vfs.splitPath(paths[it])
		var fileInfo *models.FileInfo
		fileInfo, err = vfs.db.GetFileInfo(user.ID, filePath, fileName)
		if err != nil {
			return
		}

		if fileInfo.ShareID > 0 {
			var shareEntry *models.ShareEntry
			shareEntry, err = vfs.db.GetShareEntryByID(fileInfo.ShareID)
			if err != nil {
				return
			}

			var origFileInfo *models.FileInfo
			origFileInfo, err = vfs.db.GetFileInfoWithID(shareEntry.FileID)
			if err != nil {
				return
			}

			origUserPath := vfs.getUserPathWithID(origFileInfo.OwnerID)
			paths[it] = filepath.Join(origUserPath, paths[it])
		} else {
			paths[it] = filepath.Join(userPath, paths[it])
		}
	}

	zipPath = filepath.Join(TmpName, outputName)
	outputPath := filepath.Join(userPath, zipPath)

	err = vfs.fs.ZipFiles(paths, outputPath)
	if err != nil {
		return
	}

	err = vfs.FinishNewFile(user, zipPath)
	if err != nil {
		return
	}
	return
}

func (vfs *VirtualFilesystem) UpdateFile(user *models.User, path string, updatedFileInfo *models.FileInfoUpdate) (fileInfo *models.FileInfo, err error) {
	filePath, fileName := vfs.splitPath(path)
	fileInfo, err = vfs.db.GetFileInfo(user.ID, filePath, fileName)
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
		userPath := vfs.getUserPath(user)
		oldPath := filepath.Join(userPath, fileInfo.Path, fileInfo.Name)

		newFolderPath, newFolderName := vfs.splitPath(newPath)
		var newFolderInfo *models.FileInfo
		newFolderInfo, err = vfs.db.GetFileInfo(user.ID, newFolderPath, newFolderName)
		if err != nil {
			log.Error(0, "Error getting parent for changed file %v%v: %v", fileInfo.Path, fileInfo.Name, err)
			return
		}

		fileInfo.Starred = newStarred

		if !copyFlag {
			fileInfo.LastChanged = utils.GetTimestampNow()
			if newName != fileInfo.Name {
				fileInfo.Name = newName
				fileInfo.MimeType = mime.TypeByExtension(filepath.Ext(fileInfo.Name))
			}

			err = vfs.moveFileInDB(user, fileInfo, newFolderInfo)
			if err != nil {
				return
			}

			if fileInfo.ShareID <= 0 {
				newPath := filepath.Join(userPath, fileInfo.Path, fileInfo.Name)
				err = vfs.fs.MoveFile(oldPath, newPath)
				if err != nil {
					log.Error(0, "Error moving file from %v to %v: %v", oldPath, newPath, err)
					return
				}
			}
		} else {
			err = vfs.copyFile(user, fileInfo, newName, newFolderInfo)
			if err != nil {
				return
			}
		}
	} else if newStarred != fileInfo.Starred {
		fileInfo.LastChanged = utils.GetTimestampNow()
		fileInfo.Starred = newStarred
		err = vfs.db.UpdateFile(fileInfo)
		if err != nil {
			log.Error(0, "Error updating file in db: %v", err)
			return
		}
	}

	//TODO: Make asynchronus scan call for dir sizes?!?

	return
}

func (vfs *VirtualFilesystem) moveFileInDB(user *models.User, fileInfo *models.FileInfo, parentFileInfo *models.FileInfo) (err error) {
	fileInfo.Path = utils.ConvertToSlash(filepath.Join(parentFileInfo.Path, parentFileInfo.Name), true)
	fileInfo.ParentID = parentFileInfo.ID

	err = vfs.db.UpdateFile(fileInfo)
	if err != nil {
		log.Error(0, "Error updating file in db: %v", err)
		return
	}

	if fileInfo.IsDir {
		var folderContent []*models.FileInfo
		folderContent, err = vfs.db.GetDirectoryContentWithID(fileInfo.ID)
		for _, contentInfo := range folderContent {
			err = vfs.moveFileInDB(user, contentInfo, fileInfo)
			if err != nil {
				return
			}
		}
	}
	return
}

func (vfs *VirtualFilesystem) copyFile(user *models.User, fileInfo *models.FileInfo, newName string, newParentFileInfo *models.FileInfo) (err error) {
	if newName == "" {
		newName = fileInfo.Name
	}
	parentPath := utils.ConvertToSlash(filepath.Join(newParentFileInfo.Path, newParentFileInfo.Name), true)
	newPath := filepath.Join(parentPath, newName)

	if fileInfo.ShareID <= 0 {
		if fileInfo.IsDir {
			err = vfs.CreateDirectoryForUser(user, newPath)
			if err != nil {
				return
			}

			var newFolderInfo *models.FileInfo
			newFolderInfo, err = vfs.db.GetFileInfo(user.ID, parentPath, newName)
			if err != nil {
				return
			}

			var folderContent []*models.FileInfo
			folderContent, err = vfs.db.GetDirectoryContentWithID(fileInfo.ID)
			for _, contentInfo := range folderContent {
				err = vfs.copyFile(user, contentInfo, "", newFolderInfo)
				if err != nil {
					return
				}
			}
		} else {
			userPath := vfs.getUserPath(user)
			oldPath := filepath.Join(fileInfo.Path, fileInfo.Name)
			err = vfs.fs.CopyFile(filepath.Join(userPath, oldPath), filepath.Join(userPath, newPath))
			if err != nil {
				return
			}
			err = vfs.FinishNewFile(user, newPath)
			if err != nil {
				return
			}
		}
	} else {
		newFileInfo := *fileInfo
		newFileInfo.Path = parentPath
		newFileInfo.ParentID = newParentFileInfo.ID
		newFileInfo.Name = newName

		err = vfs.db.InsertFile(&newFileInfo)
		if err != nil {
			return
		}
	}

	return
}

func (vfs *VirtualFilesystem) DeleteFile(user *models.User, path string) (err error) {
	filePath, fileName := vfs.splitPath(path)
	var fileInfo *models.FileInfo
	fileInfo, err = vfs.db.GetFileInfo(user.ID, filePath, fileName)
	if err != nil {
		return
	}

	err = vfs.deleteFileInDB(fileInfo)
	if err != nil {
		return
	}

	if fileInfo.ShareID <= 0 {
		err = vfs.fs.DeleteFile(filepath.Join(vfs.getUserPath(user), path))
		if err != nil {
			return
		}
	}

	//TODO: Make asynchronus scan call for dir sizes?!?

	return
}

func (vfs *VirtualFilesystem) deleteFileInDB(fileInfo *models.FileInfo) (err error) {
	err = vfs.db.DeleteFile(fileInfo)

	if fileInfo.IsDir {
		var folderContent []*models.FileInfo
		folderContent, err = vfs.db.GetDirectoryContentWithID(fileInfo.ID)
		for _, contentInfo := range folderContent {
			err = vfs.deleteFileInDB(contentInfo)
			if err != nil {
				return
			}
		}
	}
	return
}

func (vfs *VirtualFilesystem) SearchForFiles(user *models.User, path string) (results []*models.FileInfo, err error) {
	filePath, fileName := vfs.splitPath(path)
	return vfs.db.SearchForFiles(user.ID, filePath, fileName)
}

func (vfs *VirtualFilesystem) DeleteUserFiles(user *models.User) (err error) {
	err = vfs.db.DeleteUserFiles(user.ID)
	if err != nil {
		return
	}

	err = vfs.fs.DeleteFile(vfs.getUserPath(user))
	if err != nil {
		return
	}
	return
}

func (vfs *VirtualFilesystem) ShareFile(fromUser, toUser *models.User, path string) (err error) {
	filePath, fileName := vfs.splitPath(path)
	fileInfo, err := vfs.db.GetFileInfo(fromUser.ID, filePath, fileName)
	if err != nil {
		return
	}

	if fileInfo.ShareID > 0 {
		return fmt.Errorf("sharing shared files is forbidden")
	}

	shareEntry := &models.ShareEntry{
		OwnerID:			fromUser.ID,
		SharedWithID:	toUser.ID,
		FileID:				fileInfo.ID,
	}
	err = vfs.db.InsertShareEntry(shareEntry)
	if err != nil {
		return
	}

	sharedParentInfo, err := vfs.db.GetFileInfo(toUser.ID, "/", "")
	if err != nil {
		return
	}

	sharedFileInfo := &models.FileInfo{
		Path:           "/",
		Name:           fileInfo.Name,
		IsDir:          fileInfo.IsDir,
		Size:           fileInfo.Size,
		OwnerID:        toUser.ID,
		LastChanged:    utils.GetTimestampNow(),
		MimeType:       fileInfo.MimeType,
		ParentID:       sharedParentInfo.ID,
		ShareID:				shareEntry.ID,
		Starred:        false,
	}

	err = vfs.db.InsertFile(sharedFileInfo)
	return
}

func (vfs *VirtualFilesystem) CheckPathAccess(userID uint32, requestedPath string) (finalFileInfo *models.FileInfo, err error) {
	filePath, fileName := vfs.splitPath(requestedPath)
	var fileInfo *models.FileInfo
	fileInfo, err = vfs.db.GetFileInfo(userID, filePath, fileName)

	if err == nil && fileInfo.ShareID <= 0 {  // File exists in db for user and is owned by him: Directly return info
		finalFileInfo = fileInfo
		return
	} else if err == nil && fileInfo.ShareID > 0 {  // File exists in db for user and is shared with him: Get orig file and return it
		shareEntry, err := vfs.CheckShareEntry(userID, fileInfo.ShareID)
		if err != nil {
			return
		}

		finalFileInfo = &models.FileInfo{}
		finalFileInfo, err = vfs.db.GetFileInfoWithID(shareEntry.FileID)
		if err != nil {
			return
		}

		return
	} else {  // File does not exist in db: Check recusively if it is in a shared folder otherwise return not found
		var sharedParentInfo *models.FileInfo = nil
		var removedPath string
		parentPath := filePath
		parentName := ""
		for {
			removedPath = filepath.Join(removedPath, parentName)
			parentPath, parentName = vfs.splitPath(parentPath)

			var fileInfo *models.FileInfo
			fileInfo, err = vfs.db.GetFileInfo(userID, parentPath, parentName)

			if err == nil && fileInfo.ShareID <= 0 {  // Found existing parent but it is not shared --> Requested file does not exist
				break
			} else if err == nil && fileInfo.ShareID > 0 {  // Found existing parent that is shared --> Check share and remember parent
				shareEntry, err := vfs.CheckShareEntry(userID, fileInfo.ShareID)
				if err != nil {
					return
				}

				sharedParentInfo = &models.FileInfo{}
				finalFileInfo, err = vfs.db.GetFileInfoWithID(shareEntry.FileID)
				if err != nil {
					return
				}

				break
			}
		}

		if sharedParentInfo == nil {
			return nil, ErrFileNotFound
		}

		finalFileInfo = &models.FileInfo{}
		finalFileInfo, err = vfs.db.GetFileInfo(sharedParentInfo.OwnerID, filepath.Join(sharedParentInfo.Path, sharedParentInfo.Name, removedPath), fileName)
		if err != nil {
			return
		}
	}

	return
}

func (vfs *VirtualFilesystem) CheckShareEntry(userID, shareID uint32) (shareEntry *models.ShareEntry, err error) {
	shareEntry, err = vfs.db.GetShareEntryByID(shareID)
	if err != nil {
		return
	}

	if shareEntry.SharedWithID != userID {
		err = fmt.Errorf("user of shareEntry not matching with requested user")
		return
	}

	return
}
