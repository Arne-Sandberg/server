package fs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/riesinger/freecloud/models"
	log "gopkg.in/clog.v1"
)

// DiskFilesystem implements the Filesystem interface, writing actual files to the disk
type DiskFilesystem struct {
	base string
}

// NewDiskFilesystem sets up a disk filesystem and returns it
func NewDiskFilesystem(baseDir string) (*DiskFilesystem, error) {
	base, err := filepath.Abs(baseDir)
	if err != nil {
		log.Error(0, "Could not initialize filesystem: %v", err)
		return nil, err
	}

	// Check if the base directory does not exist. If it doesn't, create it.
	_, err = os.Stat(base)
	if os.IsNotExist(err) {
		log.Info("Base directory does not exist, creating it now")
		err := os.Mkdir(base, 0755)
		if err != nil {
			log.Error(0, "Could not create base directory: %v", err)
			return nil, err
		}
	} else if err != nil {
		log.Warn("Could not check if base directory exists, assuming it does")
	}

	log.Info("Initialized filesystem at base directory %s", base)
	return &DiskFilesystem{base}, nil
}

// NewFileHandle opens an *os.File handle for writing to
func (dfs *DiskFilesystem) NewFileHandle(path string) (*os.File, error) {
	f, err := os.Create(filepath.Join(dfs.base, path))
	if err != nil {
		log.Error(0, "Could not create file %s: %v", path, err)
		return nil, err
	}
	return f, nil
}

func (dfs *DiskFilesystem) ListFiles(path string) ([]os.FileInfo, error) {
	info, err := ioutil.ReadDir(filepath.Join(dfs.base, path))
	if err != nil {
		log.Error(0, "Could not list files in %s: %v", path, err)
		return nil, err
	}
	return info, nil
}

func (dfs *DiskFilesystem) CreateDirectory(path string) error {
	err := os.MkdirAll(filepath.Join(dfs.base, path), 0755)
	if err != nil {
		log.Error(0, "Could not create directory %s: %v", path, err)
	}
	return err
}

func (dfs *DiskFilesystem) GetUserBaseDirectory(user *models.User) string {
	return strconv.Itoa(user.ID)
}

func (dfs *DiskFilesystem) NewFileHandleForUser(user *models.User, path string) (*os.File, error) {
	if err := dfs.createUserDirIfNotExist(user); err != nil {
		return nil, err
	}
	return dfs.NewFileHandle(filepath.Join(dfs.GetUserBaseDirectory(user), path))
}

func (dfs *DiskFilesystem) CreateDirectoryForUser(user *models.User, path string) error {
	return dfs.CreateDirectory(filepath.Join(dfs.GetUserBaseDirectory(user), path))
}

func (dfs *DiskFilesystem) ListFilesForUser(user *models.User, path string) ([]os.FileInfo, error) {
	if err := dfs.createUserDirIfNotExist(user); err != nil {
		return nil, err
	}
	return dfs.ListFiles(filepath.Join(dfs.GetUserBaseDirectory(user), path))
}

func (dfs *DiskFilesystem) createUserDirIfNotExist(user *models.User) error {
	path := filepath.Join(dfs.base, dfs.GetUserBaseDirectory(user))
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		log.Info("User directory does not exist, creating it now")
		// We can safely assume that the actual base directory exists, as it is created on initialization
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
