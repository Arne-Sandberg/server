package models

// Star represensts that a user starred a file
type Star struct {
	FileID int64 `gorm:"primary_key;auto_increment:false"`
	UserID int64 `gorm:"primary_key;auto_increment:false"`
}
