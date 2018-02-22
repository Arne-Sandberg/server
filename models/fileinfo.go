package models

import "time"

type FileInfo struct {
	ID             int       `storm:"id,increment"`
	Path           string    `json:"path,omitempty"`
	Name           string    `json:"name,omitempty"`
	IsDir          bool      `json:"isDir"`
	Size           int64     `json:"size"`
	OwnerID        int       `json:"ownerID,omitempty"`
	LastChanged    time.Time `json:"lastChanged,omitempty"`
	MimeType       string    `json:"mimetype,omitempty"`
	ParentID       int       `json:"parentID,omitempty"`
	OriginalFileID int       `json:"originalFileID,omitempty"`
}
