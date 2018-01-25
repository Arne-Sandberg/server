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

// ValidateAvatarURL checks whether the avatar URL is valid
func ValidateAvatarURL(avatarURL string) bool {
	return true // TODO: Check whether there needs to be a check for a valid avatar URL
}
