package auth

const (
	SaltLength = 16
)

func HashPassword(plaintext string) (hash string) {
	// The salt does not need to be cryptographically secure
	salt := utils.RandomString(SaltLength)
}