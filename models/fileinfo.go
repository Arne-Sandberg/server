package models

type FileInfo struct {
	Path string
	Name string
	IsDir bool
	Size int64
	OwnerID int
}