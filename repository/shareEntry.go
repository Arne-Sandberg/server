package repository

import (
	"github.com/freecloudio/freecloud/models"
	log "gopkg.in/clog.v1"
)

// Add used models to enable auto migration for them
func init() {
	databaseModels = append(databaseModels, &models.ShareEntry{})
}

type ShareEntryRepository struct{}

var shareEntryRepository *ShareEntryRepository

func CreateShareEntryRepository() (*ShareEntryRepository, error) {
	if databaseConnection == nil {
		return nil, ErrGormNotInitialized
	}

	if shareEntryRepository != nil {
		return shareEntryRepository, nil
	}

	shareEntryRepository = &ShareEntryRepository{}
	return shareEntryRepository, nil
}

func GetShareEntryRepository() *ShareEntryRepository {
	return shareEntryRepository
}

func (rep *ShareEntryRepository) Create(shareEntry *models.ShareEntry) (err error) {
	err = databaseConnection.Create(shareEntry).Error
	if err != nil {
		log.Error(0, "Could not insert share entry: %v", err)
		return
	}
	return
}

func (rep *ShareEntryRepository) GetByID(shareID int64) (shareEntry *models.ShareEntry, err error) {
	shareEntry = &models.ShareEntry{}
	err = databaseConnection.First(shareEntry, "id = ?", shareID).Error
	if err != nil {
		log.Error(0, "Could not get shareEntry for ID %v: %v", shareID, err)
		return
	}
	return
}

func (rep *ShareEntryRepository) GetByFileID(fileID int64) (shareEntries []*models.ShareEntry, err error) {
	err = databaseConnection.Find(&shareEntries, &models.ShareEntry{FileID: fileID}).Error
	if err != nil {
		log.Error(0, "Could not get shareEntries for FileID %v: %v", fileID, err)
		return
	}
	return
}
