package utils

import (
	"strings"
)

// ValidateEmail uses simple checks to validate the email
func ValidateEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// ValidatePassword checks whether the password is longer than 6 characters
func ValidatePassword(password string) bool {
	return len(password) > 6
}
