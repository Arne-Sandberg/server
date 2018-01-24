package fs

import (
	"errors"
	"os"

	"github.com/freecloudio/freecloud/models"
)

var (
	ErrUpwardsNavigation = errors.New("upward navigating directories is not allowed")
)

// Filesystem is an interface for implementing various filesystem layers, such as a disk
// filesystem and a memory filesystem.
type Filesystem interface {
	NewFileHandle(path string) (*os.File, error)
	CreateDirectory(path string) error
	// GetUserBaseDirectory returns the user's base directory name, relative to the filesystem base.
	GetUserBaseDirectory(user *models.User) string
	NewFileHandleForUser(user *models.User, path string) (*os.File, error)
	CreateDirectoryForUser(user *models.User, path string) error
	ListFilesForUser(user *models.User, path string) ([]*models.FileInfo, error)
}
