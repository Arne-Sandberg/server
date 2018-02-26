package fs

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/freecloudio/freecloud/utils"

	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/models"
	log "gopkg.in/clog.v1"
)

type vfsDatabase interface {
	InsertFile(fileInfo *models.FileInfo) (err error)
	RemoveFile(fileInfo *models.FileInfo) (err error)
	UpdateFile(fileInfo *models.FileInfo) (err error)
	DeleteFile(fileInfo *models.FileInfo) (err error)
	// Must return an empty instead of an error if nothing could be found
	GetDirectoryContent(userID int, path, dirName string) (dirInfo *models.FileInfo, content []*models.FileInfo, err error)
	GetDirectoryContentWithID(directoryID int) (content []*models.FileInfo, err error)
	GetFileInfo(userID int, path, fileName string) (fileInfo *models.FileInfo, err error)
	GetFileInfoWithID(fileID int) (fileInfo *models.FileInfo, err error)
}

type VirtualFilesystem struct {
	fs      Filesystem
	db      vfsDatabase
	tmpName string
}

func NewVirtualFilesystem(fs Filesystem, db vfsDatabase, tmpName string) (vfs *VirtualFilesystem, err error) {
	vfs = &VirtualFilesystem{fs, db, tmpName}
	err = vfs.ScanFSForChanges()

	return
}

func (vfs *VirtualFilesystem) ScanFSForChanges() (err error) {
	log.Trace("Get existing users")
	existingUsers, err := auth.GetAllUsers()
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
	//Create user dir if not existing and add it to db
	created, err := vfs.fs.CreateDirIfNotExist(vfs.getUserPath(user))
	if err != nil {
		log.Error(0, "Error creating folder for user id %v", user.ID)
		return
	}
	_, err = vfs.db.GetFileInfo(user.ID, "/", "")
	if created || err != nil {
		err = vfs.db.InsertFile(&models.FileInfo{
			Path:        "/",
			Name:        "",
			IsDir:       true,
			OwnerID:     user.ID,
			LastChanged: time.Now(),
		})
		if err != nil {
			log.Error(0, "Error inserting created root folder for user id %v", user.ID)
			return
		}
	}

	//Create tmp dir for every user
	created, err = vfs.fs.CreateDirIfNotExist(filepath.Join(vfs.getUserPath(user), vfs.tmpName))
	if err != nil {
		log.Error(0, "Error creating tmp folder for user id %v", user.ID)
		return
	}
	_, err = vfs.db.GetFileInfo(user.ID, "/", vfs.tmpName)
	if created || err != nil {
		err = vfs.db.InsertFile(&models.FileInfo{
			Path:        "/",
			Name:        vfs.tmpName,
			IsDir:       true,
			OwnerID:     user.ID,
			LastChanged: time.Now(),
		})
		if err != nil {
			log.Error(0, "Error inserting created tmp folder for user id %v", user.ID)
			return
		}
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
		return pathInfo.Size, fmt.Errorf("Path is not a directory")
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

func (vfs *VirtualFilesystem) getUserPathWithID(userID int) string {
	return "/" + filepath.Join(strconv.Itoa(userID))
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
		err = fmt.Errorf("Could not find parent folder of folder in db: %v", err)
		log.Error(0, "%v", err)
		return
	}

	err = vfs.fs.CreateDirectory(filepath.Join(userPath, path))
	if err != nil {
		err = fmt.Errorf("Error creating directory for user %v: %v", user.ID, err)
		log.Error(0, "%v", err)
		return
	}

	dirInfo := &models.FileInfo{
		Path:        utils.ConvertToSlash(folderPath, true),
		Name:        folderName,
		IsDir:       true,
		OwnerID:     user.ID,
		LastChanged: time.Now(),
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

	if fileInfo.OriginalFileID > 0 {
		var origFileInfo *models.FileInfo
		origFileInfo, err = vfs.db.GetFileInfoWithID(fileInfo.OriginalFileID)
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

		if fileInfo.OriginalFileID > 0 {
			var origFileInfo *models.FileInfo
			origFileInfo, err = vfs.db.GetFileInfoWithID(fileInfo.OriginalFileID)
			if err != nil {
				return
			}
			origUserPath := vfs.getUserPathWithID(origFileInfo.OwnerID)
			paths[it] = filepath.Join(origUserPath, paths[it])
		} else {
			paths[it] = filepath.Join(userPath, paths[it])
		}
	}

	zipPath = filepath.Join(vfs.tmpName, outputName)
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

func (vfs *VirtualFilesystem) UpdateFile(user *models.User, path string, updates map[string]interface{}) (fileInfo *models.FileInfo, err error) {
	filePath, fileName := vfs.splitPath(path)
	fileInfo, err = vfs.db.GetFileInfo(user.ID, filePath, fileName)
	if err != nil {
		return
	}

	var newPath string
	if rawNewPath, ok := updates["path"]; ok == true {
		newPath, ok = rawNewPath.(string)
		if ok != true {
			err = fmt.Errorf("Given path is not a string")
			return
		}
	}

	var newName string
	if rawNewName, ok := updates["name"]; ok == true {
		newName, ok = rawNewName.(string)
		if ok != true {
			err = fmt.Errorf("Given name is not a string")
			return
		}
	}

	var copyFlag bool
	if rawCopy, ok := updates["copy"]; ok == true {
		copyFlag, ok = rawCopy.(bool)
		if ok != true {
			err = fmt.Errorf("Given copy flag is not a bool")
			return
		}
	}

	if newPath == "" || !utils.ValidatePath(newPath) {
		newPath = fileInfo.Path
	}
	if newName == "" || !utils.ValidatePath(newName) {
		newName = fileInfo.Name
	}

	if newPath != fileInfo.Path || newName != fileInfo.Name {
		userPath := vfs.getUserPath(user)
		oldPath := filepath.Join(userPath, fileInfo.Path, fileInfo.Name)

		folderPath, folderName := vfs.splitPath(newPath)
		var folderInfo *models.FileInfo
		folderInfo, err = vfs.db.GetFileInfo(user.ID, folderPath, folderName)
		if err != nil {
			log.Error(0, "Error getting parent for changed file %v%v: %v", fileInfo.Path, fileInfo.Name, err)
			return
		}

		if !copyFlag {
			fileInfo.LastChanged = time.Now()
			if newName != fileInfo.Name {
				fileInfo.Name = newName
				fileInfo.MimeType = mime.TypeByExtension(filepath.Ext(fileInfo.Name))
			}

			err = vfs.moveFileInDB(user, fileInfo, folderInfo)
			if err != nil {
				return
			}

			if fileInfo.OriginalFileID <= 0 {
				newPath := filepath.Join(userPath, fileInfo.Path, fileInfo.Name)
				err = vfs.fs.MoveFile(oldPath, newPath)
				if err != nil {
					log.Error(0, "Error moving file: %v", oldPath, newPath, err)
					return
				}
			}
		} else {
			err = vfs.copyFile(user, fileInfo, newName, folderInfo)
			if err != nil {
				return
			}
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

	if fileInfo.OriginalFileID <= 0 {
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

	if fileInfo.OriginalFileID <= 0 {
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