package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
	// Must return an empty instead of an error if nothing could be found
	GetDirectoryContent(userID int, path, dirName string) (dirInfo *models.FileInfo, content []*models.FileInfo, err error)
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
	existingUsers, err := auth.GetExisingUsers()
	if err != nil {
		log.Error(0, "Could not get exising users: %v", err)
		return
	}

	for _, user := range existingUsers {
		//Create user dir if not existing and add it to db
		created, err := vfs.fs.CreateDirIfNotExist(vfs.getUserPath(user))
		if err != nil {
			log.Error(0, "Error creating folder for user id %v", user.ID)
			continue
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
				continue
			}
		}

		_, err = vfs.scanDirForChanges(user, "/", "")
		if err != nil {
			log.Error(0, "Could not scan directory for user %v: %v", user.ID, err)
			return err
		}
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
	if origPath == "/" || origPath == "." || origPath == "" {
		return "/", ""
	}

	path = utils.ConvertToSlash(filepath.Dir(origPath))
	if path[len(path)-1] != '/' {
		path += "/"
	}
	name = filepath.Base(origPath)
	return
}

func (vfs *VirtualFilesystem) NewFileHandleForUser(user *models.User, path string) (*os.File, error) {
	return vfs.fs.NewFileHandle(filepath.Join(vfs.getUserPath(user), path))
}

func (vfs *VirtualFilesystem) FinishFileHandle(user *models.User, path string) (err error) {
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

	//TODO: Make asynchronus scan call?!?

	return
}

func (vfs *VirtualFilesystem) CreateDirectoryForUser(user *models.User, path string) (err error) {
	folderPath, folderName := vfs.splitPath(path)
	userPath := vfs.getUserPath(user)

	parFolderPath, parFolderName := vfs.splitPath(folderPath)
	parFolderInfo, err := vfs.db.GetFileInfo(user.ID, parFolderPath, parFolderName)
	if err != nil {
		log.Error(0, "Could not find parent folder of finishes fileHandle in db: %v", err)
		return
	}

	err = vfs.fs.CreateDirectory(filepath.Join(userPath, path))
	if err != nil {
		log.Error(0, "Error creating directory for user %v: %v", user.ID, err)
		return
	}

	dirInfo := &models.FileInfo{
		Path:        folderPath,
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
	return
}

func (vfs *VirtualFilesystem) UpdateFile(user *models.User, path string, updates map[string]interface{}) (fileInfo *models.FileInfo, err error) {
	if !utils.ValidatePath(path) {
		err = ErrForbiddenPathName
		return
	}

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

	if newPath == "" || !utils.ValidatePath(newPath) {
		newPath = fileInfo.Path
	}
	if newName == "" || !utils.ValidatePath(newName) {
		newName = fileInfo.Name
	}

	if newPath != fileInfo.Path || newName != fileInfo.Name {
		// TODO: Update file in db

		if fileInfo.OriginalFileID <= 0 {
			userPath := vfs.getUserPath(user)
			newPath := filepath.Join(userPath, newPath, newName)
			oldPath := filepath.Join(userPath, fileInfo.Path, fileInfo.Name)
			err = vfs.fs.MoveFile(oldPath, newPath)
			if err != nil {
				log.Error(0, "Error moving file from %v tp %v: %v", oldPath, newPath, err)
				return
			}
		}
	}
	return
}
