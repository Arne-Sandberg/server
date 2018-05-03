package models

import "time"

// User represents a single end-user.
type User struct {
	ID          int       `storm:"id,increment" json:"id"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	Email       string    `storm:"unique" json:"email"`
	Password    string    `json:"password,omitempty"`
	AvatarURL   string    `json:"avatarURL"`
	IsAdmin     bool      `json:"isAdmin"`
	Created     time.Time `json:"created,omitempty"`
	Updated     time.Time `json:"updated,omitempty"`
	LastSession time.Time `json:"lastSession,omitempty"`
}
