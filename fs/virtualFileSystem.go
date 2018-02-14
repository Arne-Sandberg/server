package fs

import (
	"os"

	"github.com/freecloudio/freecloud/models"
)

type vfsDatabase interface {
}

type VirtualFilesystem struct {
	fs Filesystem
	db vfsDatabase
}

func (vfs *VirtualFilesystem) NewFileHandleForUser(user *models.User, path string) (*os.File, error) {

}

func (vfs *VirtualFilesystem) CreateDirectoryForUser(user *models.User, path string) error {

}

func (vfs *VirtualFilesystem) ListFilesForUser(user *models.User, path string) ([]*models.FileInfo, error) {

}

func (vfs *VirtualFilesystem) GetFileInfo(user *models.User, path string) (fileInfo *models.FileInfo, err error) {

}

// ZipFiles zips all given files/directories of paths to a zip archive with the given name in the temp folder
func (vfs *VirtualFilesystem) ZipFiles(user *models.User, paths []string, outputName string) (zipPath string, err error) {

}

func (vfs *VirtualFilesystem) UpdateFile(user *models.User, path string, updates map[string]interface{}) (fileInfo *models.FileInfo, err error) {

}
