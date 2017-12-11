package auth

type CredentialsProvider interface {
	VerifyUserPassword(email string, plaintext string) (bool, error)
}
