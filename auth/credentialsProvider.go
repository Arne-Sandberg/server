package auth

import "github.com/freecloudio/freecloud/models"

// CredentialsProvider is an interface for various credential sources like Databases or alike
type CredentialsProvider interface {
	GetUserByID(userID uint32) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
	GetAllUsers() ([]*models.User, error)
	DeleteUser(userID uint32) error
	GetAdminCount() (int, error)
}
