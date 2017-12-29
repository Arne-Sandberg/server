package db

import (
	"time"

	"github.com/asdine/storm"
	"github.com/pkg/errors"
	"github.com/riesinger/freecloud/auth"
	"github.com/riesinger/freecloud/models"
	log "gopkg.in/clog.v1"
)

type StormDB struct {
	c *storm.DB
}

func NewStormDB() (*StormDB, error) {
	db, err := storm.Open("freecloud.db")
	if err != nil {
		log.Error(0, "Could not open datbase: %v", err)
		return nil, err
	}
	log.Info("Initialized database")
	return &StormDB{c: db}, nil
}

func (db *StormDB) Close() {
	db.c.Close()
}

func (db *StormDB) CreateUser(user *models.User) (err error) {
	user.Created = time.Now().UTC()
	user.Updated = time.Now().UTC()
	// Hash the user's password
	hash, err := auth.HashPassword(user.Password)
	if err != nil {
		log.Error(0, "Password hashing failed: %v", err)
		return
	}
	user.Password = hash
	err = db.c.Save(user)
	if err != nil {
		log.Error(0, "Could not save user: %v", err)
	}
	return
}

func (db *StormDB) GetUserByID(uid int) (user *models.User, err error) {
	err = db.c.One("ID", uid, user)
	return
}

func (db *StormDB) VerifyUserPassword(email string, plaintext string) (valid bool, err error) {

	var user models.User
	err = db.c.One("Email", email, &user)
	if err != nil {
		log.Warn("Could not find user by email %s: %v", email, err)
		valid = false
		err = errors.Wrap(err, "unable to find user")
		return
	}

	// Once we got the user, verify the password
	valid, err = auth.ValidatePassword(plaintext, user.Password)
	if err != nil {
		log.Warn("Could not verify password: %v", err)
		err = errors.Wrap(err, "password verification failed")
		return
	}

	return

}
