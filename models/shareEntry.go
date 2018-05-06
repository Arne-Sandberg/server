package models

type ShareEntry struct {
	ID    				uint32 `storm:"id,increment"`
	OwnerID     	uint32
	FileID 				uint32
	SharedWithID	uint32
}
