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
}

type VirtualFilesystem struct {
	fs Filesystem
	db vfsDatabase
}

func NewVirtualFilesystem(fs Filesystem, db vfsDatabase) (vfs *VirtualFilesystem, err error) {
	vfs = &VirtualFilesystem{fs, db}
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
	osPathInfo, err := vfs.fs.GetOSFileInfo(filepath.Join(userPath, path, name))
	// Return if the scanning dir is a file
	if err != nil || !osPathInfo.IsDir() {
		return osPathInfo.Size(), fmt.Errorf("Path is not a directory")
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
	return "/" + filepath.Join(strconv.Itoa(user.ID))
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
	// TODO
	return nil, nil
}

func (vfs *VirtualFilesystem) CreateDirectoryForUser(user *models.User, path string) error {
	// TODO
	return nil
}

func (vfs *VirtualFilesystem) ListFilesForUser(user *models.User, path string) ([]*models.FileInfo, error) {
	// TODO
	return nil, nil
}

func (vfs *VirtualFilesystem) GetFileInfo(user *models.User, path string) (fileInfo *models.FileInfo, err error) {
	// TODO
	return nil, nil
}

func (vfs *VirtualFilesystem) GetDownloadURL(user *models.User, path string) (downloadURL, filename string, err error) {
	// TODO
	return "", "", nil
}

// ZipFiles zips all given files/directories of paths to a zip archive with the given name in the temp folder
func (vfs *VirtualFilesystem) ZipFiles(user *models.User, paths []string, outputName string) (zipPath string, err error) {
	// TODO
	return "", nil
}

func (vfs *VirtualFilesystem) UpdateFile(user *models.User, path string, updates map[string]interface{}) (fileInfo *models.FileInfo, err error) {
	// TODO
	return nil, nil
}
