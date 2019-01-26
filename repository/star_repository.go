package repository

import (
	"github.com/freecloudio/server/models"
	log "gopkg.in/clog.v1"
)

// Add used models to enable auto migration for them
func init() {
	databaseModels = append(databaseModels, &models.Star{})
}

// StarRepository represents the database for storing stars
type StarRepository struct{}

// CreateStarRepository creates a new StarRepository IF gorm has been initialized before
func CreateStarRepository() (*StarRepository, error) {
	if databaseConnection == nil {
		return nil, ErrGormNotInitialized
	}
	return &StarRepository{}, nil
}

// Create stores a new star
func (rep *StarRepository) Create(star *models.Star) (err error) {
	err = databaseConnection.Create(star).Error
	if err != nil {
		log.Error(0, "Could not create star: %v", err)
		return
	}
	return
}

// Delete deletes a star by its fileID and userID
func (rep *StarRepository) Delete(fileID, userID int64) (err error) {
	err = databaseConnection.Delete(&models.Star{FileID: fileID, UserID: userID}).Error
	if err != nil {
		log.Error(0, "Could not delete star: %v", err)
		return
	}
	return
}

// Exists returns whether a file is starred by an user
func (rep *StarRepository) Exists(fileID, userID int64) (exists bool, err error) {
	err = databaseConnection.First(&models.Star{FileID: fileID, UserID: userID}).Error
	if err != nil && !IsRecordNotFoundError(err) {
		log.Error(0, "Could not get star with fileID '%d' and userID '%d': %v", fileID, userID, err)
		return
	} else if err != nil {
		exists = false
		err = nil
	} else {
		exists = true
	}
	return
}

// Count returns the amount if stored stars
func (rep *StarRepository) Count() (count int64, err error) {
	err = databaseConnection.Model(&models.Star{}).Count(&count).Error
	if err != nil {
		log.Error(0, "Error counting total stars: %v", err)
		return
	}
	return
}
