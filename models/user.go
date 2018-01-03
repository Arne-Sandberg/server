package models

import "time"

// User represents a single end-user.
type User struct {
	ID          int    `storm:"id,increment"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Email       string `storm:"unique" json:"email"`
	Password    string `json:"password"`
	AvatarURL   string `json:"avatarURL"`
	SignedIn    bool
	IsAdmin     bool
	Created     time.Time
	Updated     time.Time
	LastSession time.Time
}
