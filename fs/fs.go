package fs

import (
	"errors"
	"os"

	"github.com/freecloudio/freecloud/models"
)

var (
	// ErrUpwardsNavigation gets raised when a possible upwards navigations gets detected (such as paths containing "../"" or "~")
	ErrUpwardsNavigation = errors.New("upward navigating directories is not allowed")
	// ErrForbiddenPathName indicates a path having weird characters that nobody should use, also these characters are forbidden on Windows
	ErrForbiddenPathName = errors.New("paths cannot contain the following characters: <>:\"\\|?*")
	ErrFileNotExist      = errors.New("file does not exist")
)

const (
	forbiddenPathCharacters = "<>:\"|?*"
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
	// ResolveFilePath returns the full path for a given file and user.
	// This is used in the download handler
	ResolveFilePath(user *models.User, path string) (fullPath string, filename string, err error)
}
