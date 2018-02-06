package models

import "time"

type FileInfo struct {
	Path        string    `json:"path,omitempty"`
	Name        string    `json:"name,omitempty"`
	IsDir       bool      `json:"isDir,omitempty"`
	Size        int64     `json:"size,omitempty"`
	OwnerID     int       `json:"ownerID,omitempty"`
	LastChanged time.Time `json:"lastChanged,omitempty"`
	MimeType    string    `json:"mimetype,omitempty"`
}
