package utils

import (
	"regexp"
	"strings"
)

var usernameRegex = regexp.MustCompile(`^([a-zA-Z0-9\-_]+\.?)*$`)

// ValidateUsername checks whether the username does not contain illegal characters
func ValidateUsername(username string) bool {
	return username != "" && usernameRegex.MatchString(username)
}

// ValidateEmail uses simple checks to validate the email
func ValidateEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// ValidatePassword checks whether the password is longer than 6 characters
func ValidatePassword(password string) bool {
	return len(password) >= 6
}

// ValidateFirstName checks whether the first name is filled
func ValidateFirstName(firstName string) bool {
	return len(firstName) > 0
}

// ValidateLastName checks whether the last name is filled
func ValidateLastName(lastName string) bool {
	return len(lastName) > 0
}

const (
	forbiddenPathCharacters = "<>:\"|?*"
)

// ValidatePath checks whether the path contains any forbidden characters
func ValidatePath(path string) bool {
	return !(strings.Contains(path, "..") || strings.Contains(path, "~") || strings.ContainsAny(path, forbiddenPathCharacters))
}
