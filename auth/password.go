package auth

import (
	"github.com/riesinger/freecloud/auth/hashers"
)

// HashPassword uses a hashing function found in hashers to encrypt a password.
func HashPassword(plaintext string) (hash string, err error) {
	hash, err = hashers.HashScrypt(plaintext)
	return
}

// ValidatePassword uses the same hashing function as HashPassword to verify a
// plaintext password against its hash.
func ValidatePassword(plaintext, hash string) (valid bool, err error) {
	valid, err = hashers.ValidateScryptPassword(plaintext, hash)
	return
}
