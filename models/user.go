package models

import "time"

// User represents a single end-user.
type User struct {
	ID int
	FirstName string
	LastName  string
	Email     string
	Password  string
	AvatarURL string
	SignedIn bool
	Created   time.Time
}

