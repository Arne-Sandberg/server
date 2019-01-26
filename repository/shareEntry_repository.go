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

// GetByIDForUser reads and returns a share entry by shareID and whether the userID is owner or shared_with
func (rep *ShareEntryRepository) GetByIDForUser(shareID int64, userID int64) (shareEntry *models.ShareEntry, err error) {
	shareEntry = &models.ShareEntry{}
	err = databaseConnection.Raw(getByIDAndUserQuery, shareID, userID, userID).Scan(shareEntry).Error
	if err != nil {
		log.Error(0, "Could not get shareEntry for ID %v and user %v: %v", shareID, userID, err)
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
	fromPart = `
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
	whereShareIDPart = " where orig.share_id = ?"
	whereFileIDPart  = " where orig.file_id = ?"
	andUserIDPart    = " and (orig.owner_id = ? or share.shared_with_id = ?)"

	getAllQuery         = "select orig.share_id as id, orig.file_id, orig.owner_id, share.shared_with_id" + fromPart // No variables
	getByIDQuery        = getAllQuery + whereShareIDPart                                                             // Only ShareID variable
	getByIDAndUserQuery = getByIDQuery + andUserIDPart                                                               // ShareID and TWO times UserID variables
	getByFileIDQuery    = getAllQuery + whereFileIDPart                                                              // Only FileID variable
)
