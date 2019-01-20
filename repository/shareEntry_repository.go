package repository

import (
	"github.com/freecloudio/server/models"
	log "gopkg.in/clog.v1"
)

// Add used models to enable auto migration for them
func init() {
	databaseModels = append(databaseModels, &models.ShareEntry{})
}

// ShareEntryRepository represents the database for storing users
type ShareEntryRepository struct{}

// CreateShareEntryRepository creates a new ShareEntryRepository IF gorm has been initialized before
func CreateShareEntryRepository() (*ShareEntryRepository, error) {
	if databaseConnection == nil {
		return nil, ErrGormNotInitialized
	}

	return &ShareEntryRepository{}, nil
}

// Create stores a new share entry
func (rep *ShareEntryRepository) Create(shareEntry *models.ShareEntry) (err error) {
	err = databaseConnection.Create(shareEntry).Error
	if err != nil {
		log.Error(0, "Could not insert share entry: %v", err)
		return
	}
	return
}

// Delete deletes a share entry by its shareID
func (rep *ShareEntryRepository) Delete(shareID int64) (err error) {
	err = databaseConnection.Delete(&models.ShareEntry{ID: shareID}).Error
	if err != nil {
		log.Error(0, "Could not delete share entry with ID %v: %v", shareID, err)
		return
	}
	return
}

// GetByID reads and returns a share entry by shareID
func (rep *ShareEntryRepository) GetByID(shareID int64) (shareEntry *models.ShareEntry, err error) {
	shareEntry = &models.ShareEntry{}
	err = databaseConnection.Raw(getByIDQuery, shareID).Scan(shareEntry).Error
	if err != nil {
		log.Error(0, "Could not get shareEntry for ID %v: %v", shareID, err)
		return
	}
	return
}

// GetByFileID reads and returns a share entry by fileID
func (rep *ShareEntryRepository) GetByFileID(fileID int64) (shareEntries []*models.ShareEntry, err error) {
	err = databaseConnection.Raw(getByFileIDQuery, fileID).Scan(&shareEntries).Error
	if err != nil {
		log.Error(0, "Could not get shareEntries for FileID %v: %v", fileID, err)
		return
	}
	return
}

// Count returns the amount of stored share entries
func (rep *ShareEntryRepository) Count() (count int64, err error) {
	err = databaseConnection.Model(&models.ShareEntry{}).Count(&count).Error
	if err != nil {
		log.Error(0, "Error counting share entries: %v", err)
		return
	}
	return
}

var (
	getAllQuery = `
select orig.share_id as id, orig.file_id, orig.owner_id, share.shared_with_id
from (
			select share_entries.id as share_id, file_id, owner_id
			from share_entries
			left outer join file_infos
			on share_entries.file_id = file_infos.id ) as orig
left outer join (
			select share_entries.id as share_id, file_infos.owner_id as shared_with_id
			from share_entries
			left outer join file_infos
			on share_entries.id = file_infos.share_id ) as share
on orig.share_id = share.share_id`
	getByIDQuery     = getAllQuery + " where orig.share_id = ?"
	getByFileIDQuery = getAllQuery + " where orig.file_id = ?"
)
