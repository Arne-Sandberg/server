package models

type ShareEntry struct {
	ID    				uint32 `gorm:"primary_key;auto_increment"`
	OwnerID     	uint32
	FileID 				uint32 `gorm:"index"`
	SharedWithID	uint32
}