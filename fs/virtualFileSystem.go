package fs

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/freecloudio/freecloud/utils"

	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/models"
	log "gopkg.in/clog.v1"
)

type vfsDatabase interface {
	InsertFile(fileInfo *models.FileInfo) (err error)
	RemoveFile(fileInfo *models.FileInfo) (err error)
	UpdateFile(fileInfo *models.FileInfo) (err error)
	GetDirectoryContent(userID int, path, dirName string) (dirInfo *models.FileInfo, content []*models.FileInfo, err error)
	GetFileInfo(userID int, path, fileName string) (fileInfo *models.FileInfo, err error)
}

type VirtualFilesystem struct {
	fs Filesystem
	db vfsDatabase
}

func CreateVirtualFileSystem(fs Filesystem, db vfsDatabase) *VirtualFilesystem {
	vfs := &VirtualFilesystem{fs, db}
	//vfs.ScanFSForChanges()

	return vfs
}

func (vfs *VirtualFilesystem) ScanFSForChanges() (err error) {
	existingUsers, err := auth.GetExisingUsers()
	if err != nil {
		log.Error(0, "Could not get exising users: %v", err)
		return
	}

	for _, user := range existingUsers {
		_, err = vfs.scanDirForChanges(&user, "/", "")
		if err != nil {
			log.Error(0, "Could not scan directory for user %v: %v", user.ID, err)
			return
		}
	}
	return
}

func (vfs *VirtualFilesystem) scanDirForChanges(user *models.User, path, name string) (folderSize int64, err error) {
	// Get all needed data, paths, etc.
	userPath := vfs.getUserPath(user)
	fullPath := filepath.Join(userPath, path, name)
	osPathInfo, err := vfs.fs.GetOSFileInfo(fullPath)
	// Return if the scanning dir is a file
	if err != nil || !osPathInfo.IsDir() {
		return osPathInfo.Size(), fmt.Errorf("Path is not a directory")
	}

	// Get dir contents of fs and db
	fsFiles, err := vfs.fs.GetDirectoryContent(fullPath)
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
			fsFile.Path = vfs.removeUserFromPath(user, fsFile.Path)
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

func (vfs *VirtualFilesystem) removeUserFromPath(user *models.User, origPath string) (path string) {
	userPath := vfs.getUserPath(user)
	path = strings.Replace(origPath, userPath, "", 1)
	path = strings.TrimPrefix(path, strconv.Itoa(user.ID))
	return
}

/*
func (vfs *VirtualFilesystem) NewFileHandleForUser(user *models.User, path string) (*os.File, error) {
	// TODO
}

func (vfs *VirtualFilesystem) CreateDirectoryForUser(user *models.User, path string) error {
	// TODO
}

func (vfs *VirtualFilesystem) ListFilesForUser(user *models.User, path string) ([]*models.FileInfo, error) {
	// TODO
}

func (vfs *VirtualFilesystem) GetFileInfo(user *models.User, path string) (fileInfo *models.FileInfo, err error) {
	// TODO
}

func (vfs *VirtualFilesystem) GetDownloadURL(user *models.User, path string) (downloadURL string, err error) {
	// TODO
}

// ZipFiles zips all given files/directories of paths to a zip archive with the given name in the temp folder
func (vfs *VirtualFilesystem) ZipFiles(user *models.User, paths []string, outputName string) (zipPath string, err error) {
	// TODO
}

func (vfs *VirtualFilesystem) UpdateFile(user *models.User, path string, updates map[string]interface{}) (fileInfo *models.FileInfo, err error) {
	// TODO
}
*/
