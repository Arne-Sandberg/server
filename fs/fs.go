package fs

import (
	"errors"
	"os"

	"github.com/freecloudio/freecloud/models"
)

var (
	// ErrForbiddenPathName indicates a path having weird characters that nobody should use, also these characters are forbidden on Windows
	ErrForbiddenPathName = errors.New("paths cannot contain the following characters: <>:\"\\|?*")
	ErrFileNotExist      = errors.New("file does not exist")
)

// Filesystem is an interface for implementing various filesystem layers, such as a disk
// filesystem and a memory filesystem.
type Filesystem interface {
	Close()
	NewFileHandle(path string) (*os.File, error)
	CreateDirectory(path string) error
	CreateDirIfNotExist(path string) (created bool, err error)
	GetFileInfo(userPath, path, name string) (fileInfo *models.FileInfo, err error)
	GetDirectoryContent(userPath, path string) ([]*models.FileInfo, error)
	// Prepare download (if remote file locationn) and return local path
	GetDownloadPath(path string) string
	// ZipFiles zips all given files/directories of paths to a zip archive with the given name in the temp folder
	ZipFiles(paths []string, outputPath string) (err error)
	MoveFile(oldPath, newPath string) (err error)
}
