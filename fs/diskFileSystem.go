package fs

import (
	"os"
	"path/filepath"

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
