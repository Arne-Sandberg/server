package repository

import (
	"github.com/freecloudio/freecloud/models"
	"github.com/freecloudio/freecloud/utils"
	log "gopkg.in/clog.v1"
)

// Add used models to enable auto migration for them
func init() {
	databaseModels = append(databaseModels, &models.User{})
}

type UserRepository struct{}

func CreateUserRepository() (*UserRepository, error) {
	if databaseConnection == nil {
		return nil, ErrGormNotInitialized
	}
	return &UserRepository{}, nil
}

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

func (rep *UserRepository) Delete(userID int64) (err error) {
	err = databaseConnection.Delete(&models.User{ID: userID}).Error
	if err != nil {
		log.Error(0, "Could not delete user: %v", err)
		return
	}
	return
}

func (rep *UserRepository) Update(user *models.User) (err error) {
	user.UpdatedAt = utils.GetTimestampNow()
	err = databaseConnection.Save(user).Error
	if err != nil {
		log.Error(0, "Could not update user: %v", err)
		return
	}
	return
}

func (rep *UserRepository) GetByID(userID int64) (user *models.User, err error) {
	user = &models.User{}
	err = databaseConnection.First(&user, "id = ?", userID).Error
	return
}

func (rep *UserRepository) GetByEmail(email string) (user *models.User, err error) {
	user = &models.User{}
	err = databaseConnection.First(user, &models.User{Email: email}).Error
	return
}

func (rep *UserRepository) GetAll() (users []*models.User, err error) {
	err = databaseConnection.Find(&users).Error
	return
}

func (rep *UserRepository) AdminCount() (count int, err error) {
	var admins []*models.User
	err = databaseConnection.Find(&admins, &models.User{IsAdmin: true}).Error
	if err != nil {
		log.Error(0, "Could not get all admins: %v", err)
		count = -1
		return
	}
	count = len(admins)
	return
}

func (rep *SessionRepository) TotalCount() (count int64, err error) {
	err = databaseConnection.Model(&models.Session{}).Count(&count).Error
	if err != nil {
		log.Error(0, "Error counting total sessions: %v", err)
		return
	}
	return
}
