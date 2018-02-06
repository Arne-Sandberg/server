package fs

import (
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/freecloudio/freecloud/models"
	log "gopkg.in/clog.v1"
)

// DiskFilesystem implements the Filesystem interface, writing actual files to the disk
type DiskFilesystem struct {
	base string
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
	dfs = &DiskFilesystem{base}
	return
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

func (dfs *DiskFilesystem) ResolveFilePath(user *models.User, path string) (fullPath string, filename string, err error) {
	fullPath = filepath.Join(dfs.base, dfs.GetUserBaseDirectory(user), path)
	if _, err = os.Stat(fullPath); err == os.ErrNotExist {
		err = ErrFileNotExist
	}
	filename = filepath.Base(fullPath)
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
	info, err := ioutil.ReadDir(filepath.Join(dfs.base, dfs.GetUserBaseDirectory(user), path))
	if err != nil {
		log.Error(0, "Could not list files in %s: %v", path, err)
		return nil, err
	}
	fileInfos := make([]*models.FileInfo, len(info), len(info))
	for i, f := range info {
		fileInfos[i] = &models.FileInfo{
			Path:  filepath.Join(path, f.Name()),
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
