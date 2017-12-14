package auth

import "github.com/riesinger/freecloud/models"

// CredentialsProvider is an interface for various credential sources like Databases or alike
type CredentialsProvider interface {
	VerifyUserPassword(email string, plaintext string) (bool, error)
	NewUser(user models.User) error
}
