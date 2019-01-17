package repository

import (
	"github.com/freecloudio/server/models"
	"github.com/freecloudio/server/utils"
	log "gopkg.in/clog.v1"
)

// Add used models to enable auto migration for them
func init() {
	databaseModels = append(databaseModels, &models.User{})
}

// UserRepository represents the database for storing users
type UserRepository struct{}

// CreateUserRepository creates a new UserRepository IF gorm has been initialized before
func CreateUserRepository() (*UserRepository, error) {
	if databaseConnection == nil {
		return nil, ErrGormNotInitialized
	}
	return &UserRepository{}, nil
}

// Create stores a new user
func (rep *UserRepository) Create(user *models.User) (err error) {
	user.CreatedAt = utils.GetTimestampNow()
	user.UpdatedAt = utils.GetTimestampNow()
	err = databaseConnection.Create(user).Error
	if err != nil {
		log.Error(0, "Could not create user: %v", err)
		return
	}
	return
}

// Delete deletes a user by its userID
func (rep *UserRepository) Delete(userID int64) (err error) {
	err = databaseConnection.Delete(&models.User{ID: userID}).Error
	if err != nil {
		log.Error(0, "Could not delete user: %v", err)
		return
	}
	return
}

// Update updates a user by its user.ID
func (rep *UserRepository) Update(user *models.User) (err error) {
	user.UpdatedAt = utils.GetTimestampNow()
	err = databaseConnection.Save(user).Error
	if err != nil {
		log.Error(0, "Could not update user: %v", err)
		return
	}
	return
}

// GetByID reads and returns an user by userID
func (rep *UserRepository) GetByID(userID int64) (user *models.User, err error) {
	user = &models.User{}
	err = databaseConnection.First(user, "id = ?", userID).Error
	return
}

// GetByEmail reads and return an user by email
func (rep *UserRepository) GetByEmail(email string) (user *models.User, err error) {
	user = &models.User{}
	err = databaseConnection.First(user, &models.User{Email: email}).Error
	return
}

// GetAll reads and returns all stored users
func (rep *UserRepository) GetAll() (users []*models.User, err error) {
	err = databaseConnection.Find(&users).Error
	return
}

// AdminCount returns the amount of stored admins
func (rep *UserRepository) AdminCount() (count int64, err error) {
	err = databaseConnection.Model(&models.User{}).Where(&models.User{IsAdmin: true}).Count(&count).Error
	if err != nil {
		log.Error(0, "Could not get all admins: %v", err)
		count = -1
		return
	}
	return
}

// TotalCount returns the amount of stored users
func (rep *UserRepository) TotalCount() (count int64, err error) {
	err = databaseConnection.Model(&models.User{}).Count(&count).Error
	if err != nil {
		log.Error(0, "Error counting total sessions: %v", err)
		return
	}
	return
}
