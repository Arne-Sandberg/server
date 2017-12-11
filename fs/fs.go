package fs

import "os"

// Filesystem is an interface for implementing various filesystem layers, such as a disk
// filesystem and a memory filesystem.
type Filesystem interface {
	NewFileHandle(path string) (*os.File, error)
	CreateDirectory(path string) error
	ListFiles(path string) ([]os.FileInfo, error)
}
