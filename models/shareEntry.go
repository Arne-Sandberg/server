package models

type ShareEntry struct {
	ID           int64 `gorm:"primary_key;auto_increment"`
	OwnerID      int64
	FileID       int64 `gorm:"index"`
	SharedWithID int64
}
