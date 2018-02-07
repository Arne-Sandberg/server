package fs

import (
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/freecloudio/freecloud/config"
	"github.com/freecloudio/freecloud/utils"

	"github.com/freecloudio/freecloud/models"
	"github.com/mholt/archiver"
	log "gopkg.in/clog.v1"
)

// DiskFilesystem implements the Filesystem interface, writing actual files to the disk
type DiskFilesystem struct {
	base    string
	tmpName string
	done    chan struct{}
}

// NewDiskFilesystem sets up a disk filesystem and returns it
func NewDiskFilesystem(baseDir string) (dfs *DiskFilesystem, err error) {
	base, err := filepath.Abs(baseDir)
	if err != nil {
		log.Error(0, "Could not initialize filesystem: %v", err)
		return nil, err
	}

	// Check if the base directory does not exist. If it doesn't, create it.
	baseInfo, err := os.Stat(base)
	if os.IsNotExist(err) {
		log.Info("Base directory does not exist, creating it now")
		err = os.Mkdir(base, 0755)
		if err != nil {
			log.Error(0, "Could not create base directory: %v", err)
			return
		}
	} else if !baseInfo.IsDir() {
		log.Fatal(0, "Base directory does exist but is not a directory")
		return
	} else if err != nil {
		log.Fatal(0, "Could not check if base directory exists: %v", err)
		return
	}

	log.Info("Initialized filesystem at base directory %s", base)
	dfs = &DiskFilesystem{base: base, tmpName: config.GetString("fs.tmp_folder_name"), done: make(chan struct{})}

	go dfs.cleanupTempFolderRoutine(time.Hour * time.Duration(config.GetInt("fs.tmp_data_expiry")))

	return
}

func (dfs *DiskFilesystem) Close() {
	dfs.done <- struct{}{}
}

func (dfs *DiskFilesystem) cleanupTempFolderRoutine(interval time.Duration) {
	log.Trace("Starting temp folder cleaner, running now and every %v", interval)
	dfs.cleanupTempFolder()

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-dfs.done:
			return
		case <-ticker.C:
			dfs.cleanupTempFolder()
		}
	}
}

func (dfs *DiskFilesystem) cleanupTempFolder() {
	log.Trace("Cleaning temp folder")

	infoList, err := ioutil.ReadDir(dfs.base)
	if err != nil {
		log.Warn("Cleaning temp folder failed: %v", err)
	}

	for _, info := range infoList {
		if !info.IsDir() {
			continue
		}

		tmpFolderPath := filepath.Join(dfs.base, info.Name(), dfs.tmpName)
		tmpInfoList, err := ioutil.ReadDir(tmpFolderPath)
		if err != nil {
			log.Warn("Error reading temp folder in %v during temp cleanup: %v", tmpFolderPath, err)
		}

		for _, tmpInfo := range tmpInfoList {
			if time.Now().After(tmpInfo.ModTime().Add(time.Hour * time.Duration(config.GetInt("fs.tmp_data_expiry")))) {
				err = os.RemoveAll(filepath.Join(tmpFolderPath, tmpInfo.Name()))
				if err != nil {
					log.Warn("Error deleting file %v in temp folder %v during temp cleanup: %v", tmpInfo.Name(), tmpFolderPath, err)
					continue
				}
			}
		}
	}
}

// NewFileHandle opens an *os.File handle for writing to.
// Before opening the file, it check the path for sanity.
func (dfs *DiskFilesystem) NewFileHandle(path string) (*os.File, error) {
	if err := dfs.rejectInsanePath(path); err != nil {
		return nil, err
	}
	f, err := os.Create(filepath.Join(dfs.base, path))
	if err != nil {
		log.Error(0, "Could not create file %s: %v", path, err)
		return nil, err
	}
	return f, nil
}

// CreateDirectory creates a new directory at "path".
// Before doing so, it check the path for sanity.
func (dfs *DiskFilesystem) CreateDirectory(path string) error {
	log.Trace("Path for new directory is '%s'", path)
	if err := dfs.rejectInsanePath(path); err != nil {
		return err
	}
	err := os.MkdirAll(filepath.Join(dfs.base, path), 0755)
	if err != nil {
		log.Error(0, "Could not create directory %s: %v", path, err)
	}
	return err
}

// GetUserBaseDirectory returns the user directory's name relative to the base directory.
func (dfs *DiskFilesystem) GetUserBaseDirectory(user *models.User) string {
	return strconv.Itoa(user.ID)
}

// NewFileHandleForUser opens an *os.File in the user directory handle for writing to.
// It relies on NewFileHandle for checking the path's sanity.
func (dfs *DiskFilesystem) NewFileHandleForUser(user *models.User, path string) (*os.File, error) {
	if err := dfs.createUserDirIfNotExist(user); err != nil {
		return nil, err
	}
	return dfs.NewFileHandle(filepath.Join(dfs.GetUserBaseDirectory(user), path))
}

// CreateDirectoryForUser creates a new directory at "path" (relative to the user's directory).
// It relies on CreateDirectory for checking the path's sanity.
func (dfs *DiskFilesystem) CreateDirectoryForUser(user *models.User, path string) error {
	// We don't need to check whether the user directory exists, as it will get created automatically if it doesn't.
	return dfs.CreateDirectory(filepath.Join(dfs.GetUserBaseDirectory(user), path))
}

// ResolveFilePath resolves the given path for a user and returns the full path, if it is a file the filename or an error if the file does not exist
func (dfs *DiskFilesystem) ResolveFilePath(user *models.User, path string) (fullPath string, filename string, err error) {
	fullPath = dfs.getFullPath(user, path)
	fileInfo, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		err = ErrFileNotExist
		return
	} else if err != nil {
		err = fmt.Errorf("Error resolving file path: %v", err)
		return
	}

	if fileInfo.IsDir() {
		filename = ""
	} else {
		filename = filepath.Base(fullPath)
	}

	return
}

// ListFilesForUser returns a list of all files and folders in the given "path" (relative to the user's directory).
// Before doing so, it checks the path for sanity.
func (dfs *DiskFilesystem) ListFilesForUser(user *models.User, path string) ([]*models.FileInfo, error) {
	if err := dfs.rejectInsanePath(path); err != nil {
		return nil, err
	}
	if err := dfs.createUserDirIfNotExist(user); err != nil {
		return nil, err
	}
	info, err := ioutil.ReadDir(dfs.getFullPath(user, path))
	if err != nil {
		log.Error(0, "Could not list files in %s: %v", path, err)
		return nil, err
	}

	if path == "" {
		path = "/"
	}
	path = utils.ConvertToSlash(path)

	fileInfos := make([]*models.FileInfo, len(info), len(info))
	for i, f := range info {
		fileInfos[i] = &models.FileInfo{
			Path:  path,
			Name:  f.Name(),
			IsDir: f.IsDir(),
			Size:  f.Size(),
			// TODO: This might not be valid once we enable file sharing between users
			OwnerID:     user.ID,
			LastChanged: f.ModTime(),
			MimeType:    mime.TypeByExtension(filepath.Ext(f.Name())),
		}
	}
	return fileInfos, nil
}

func (dfs *DiskFilesystem) GetFileInfo(user *models.User, path string) (fileInfo *models.FileInfo, err error) {
	if err = dfs.rejectInsanePath(path); err != nil {
		return
	}
	if err = dfs.createUserDirIfNotExist(user); err != nil {
		return
	}

	var osFileInfo os.FileInfo
	osFileInfo, err = os.Stat(dfs.getFullPath(user, path))
	if os.IsNotExist(err) {
		err = ErrFileNotExist
		return
	}

	filePath := filepath.Dir(path)
	if filePath == "." {
		filePath = "/"
	}
	filePath = utils.ConvertToSlash(filePath)

	fileInfo = &models.FileInfo{
		Path:  filePath,
		Name:  osFileInfo.Name(),
		IsDir: osFileInfo.IsDir(),
		Size:  osFileInfo.Size(),
		// TODO: This might not be valid once we enable file sharing between users
		OwnerID:     user.ID,
		LastChanged: osFileInfo.ModTime(),
		MimeType:    mime.TypeByExtension(filepath.Ext(osFileInfo.Name())),
	}

	return
}

// createUserDirIfNotExist checks whether the base directory for the given user exists and creates it otherwise.
// This does not do any sanity checking, as the base path should always be sane.
func (dfs *DiskFilesystem) createUserDirIfNotExist(user *models.User) error {
	path := filepath.Join(dfs.base, dfs.GetUserBaseDirectory(user))
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		log.Info("User directory does not exist, creating it now")
		// We can safely assume that the actual base directory exists, as it is created on initialization
		// TODO: check whether these permissions make sense
		err := os.Mkdir(path, 0755)
		if err != nil {
			log.Error(0, "Could not create user directory: %v", err)
			return err
		}
	} else if err != nil {
		log.Warn("Could not check if user directory exists, assuming it does: %v", err)
		return nil
	}
	return nil
}

// rejectInsanePath does a sanity check on a given path and returns:
// - ErrUpwardsNavigation if upwards navigation is detected
// - ErrForbiddenPathName if there are weird characters in the path
// - nil otherwise
func (dfs *DiskFilesystem) rejectInsanePath(path string) error {
	if strings.Contains(path, "../") || strings.Contains(path, "/..") || strings.Contains(path, "~") || strings.Contains(path, "\\..") || strings.Contains(path, "..\\") {
		return ErrUpwardsNavigation
	} else if strings.ContainsAny(path, forbiddenPathCharacters) {
		return ErrForbiddenPathName
	}
	return nil
}

// ZipFiles zips all given absolute paths to a zip archive with the given name in the temp folder
func (dfs *DiskFilesystem) ZipFiles(user *models.User, paths []string, outputName string) (zipPath string, err error) {
	for it := 0; it < len(paths); it++ {
		paths[it], _, err = dfs.ResolveFilePath(user, paths[it])
		if err != nil {
			return
		}
	}

	err = dfs.CreateDirectoryForUser(user, dfs.tmpName)
	if err != nil {
		return
	}

	zipPath = filepath.Join(dfs.tmpName, outputName)
	fullZipPath := filepath.Join(dfs.base, dfs.GetUserBaseDirectory(user), zipPath)
	if err != nil {
		return
	}

	err = archiver.Zip.Make(fullZipPath, paths)

	return
}

func (dfs *DiskFilesystem) UpdateFile(user *models.User, path string, updates map[string]interface{}) (fileInfo *models.FileInfo, err error) {
	if err = dfs.rejectInsanePath(path); err != nil || dfs.getFullPath(user, path) == dfs.getFullPath(user, "/") {
		err = ErrForbiddenPathName
		return
	}

	var newPath, newName string
	fileInfo, err = dfs.GetFileInfo(user, path)
	if err != nil {
		return
	}

	if rawNewPath, ok := updates["path"]; ok == true {
		newPath, ok = rawNewPath.(string)
		if ok != true {
			err = fmt.Errorf("Given path is not a string")
			return
		}
	}

	if rawNewName, ok := updates["name"]; ok == true {
		newName, ok = rawNewName.(string)
		if ok != true {
			err = fmt.Errorf("Given name is not a string")
			return
		}
	}

	if err = dfs.rejectInsanePath(newPath); err != nil || newPath == "" {
		newPath = fileInfo.Path
	}
	if err = dfs.rejectInsanePath(newName); err != nil || newName == "" {
		newName = fileInfo.Name
	}
	newCompletePath := dfs.getFullPath(user, filepath.Join(newPath, newName))
	oldCompletePath := dfs.getFullPath(user, filepath.Join(fileInfo.Path, fileInfo.Name))

	err = os.Rename(oldCompletePath, newCompletePath)
	if err != nil {
		err = fmt.Errorf("Renaming %v failed", path)
		return
	}

	fileInfo.Path = newPath
	fileInfo.Name = newName

	return
}

func (dfs *DiskFilesystem) getFullPath(user *models.User, path string) string {
	return filepath.Join(dfs.base, dfs.GetUserBaseDirectory(user), path)
}
