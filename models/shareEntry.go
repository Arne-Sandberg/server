package models

type ShareEntry struct {
	ID    				uint32 `xorm:"pk autoincr"`
	OwnerID     	uint32
	FileID 				uint32
	SharedWithID	uint32
}
